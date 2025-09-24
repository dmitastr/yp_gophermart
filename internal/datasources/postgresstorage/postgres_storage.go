package postgresstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dmitastr/yp_gophermart/internal/config"
	"github.com/dmitastr/yp_gophermart/internal/datasources/postgresstorage/retrypolicy"
	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	_ "github.com/dmitastr/yp_gophermart/internal/errors"
	"github.com/dmitastr/yp_gophermart/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/tracelog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/avast/retry-go/v4"
)

type PostgresStorage struct {
	pool        *pgxpool.Pool
	retryPolicy *retrypolicy.RetryPolicy
}

func NewPostgresStorage(ctx context.Context, cfg *config.Config) (*PostgresStorage, error) {
	dbConfig, err := pgxpool.ParseConfig(cfg.DatabaseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db with url=%s: %v", cfg.DatabaseURI, err)
	}
	logger.Infof("Database connection established successfully")

	m, err := migrate.New(
		"file://database/migrations",
		cfg.DatabaseURI)
	if err != nil {
		return nil, err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}
	logger.Infof("Database migration succeeded")

	rp := retrypolicy.NewRetryPolicy(3, time.Second, true)

	return &PostgresStorage{pool: pool, retryPolicy: rp}, nil
}

func (p *PostgresStorage) executeTxWithRetry(ctx context.Context, tx pgx.Tx, query string, args pgx.NamedArgs, lock bool) (err error) {
	err = retry.Do(
		func() error {
			return p.executeTx(ctx, tx, query, args, lock)
		},
		p.retryPolicy.GetOptions(ctx)...,
	)

	return
}

func (p *PostgresStorage) executeWithRetry(ctx context.Context, op retry.RetryableFunc) (err error) {
	err = retry.Do(
		op,
		p.retryPolicy.GetOptions(ctx)...,
	)

	return
}

func (p *PostgresStorage) executeTx(ctx context.Context, tx pgx.Tx, query string, args pgx.NamedArgs, lock bool) (err error) {
	if lock {
		lockQuery := `SELECT pg_advisory_xact_lock(('x' || md5(@username))::bit(64)::bigint)`
		if _, err := tx.Exec(ctx, lockQuery, args); err != nil {
			return fmt.Errorf("error locking rows: %v", err)
		}
	}
	if _, err := tx.Exec(ctx, query, args); err != nil {
		return fmt.Errorf("error inserting order: %v", err)
	}

	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && !errors.Is(rErr, pgx.ErrTxClosed) {
			err = errors.Join(err, rErr)
			logger.Errorf("Error rolling back transaction: %v", rErr)
		}
	}()

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil

}

func (p *PostgresStorage) InsertUser(ctx context.Context, user *models.User) (err error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}

	query := `INSERT INTO users (username, password, created_at) VALUES (@name, @pass, @created_at)`
	args := user.ToNamedArgs()
	if err := p.executeTxWithRetry(ctx, tx, query, args, false); err != nil {
		return err
	}

	return
}

func (p *PostgresStorage) GetUser(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	query := `SELECT username, password FROM users WHERE username = $1`

	err := p.executeWithRetry(ctx, func() error {
		return p.pool.QueryRow(ctx, query, username).Scan(&user.Name, &user.Password)
	})

	return &user, err
}

func (p *PostgresStorage) UpdateUser(ctx context.Context, user *models.User) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}

	query := `UPDATE users SET password=@pass WHERE username = @name`
	if err := p.executeTxWithRetry(ctx, tx, query, user.ToNamedArgs(), false); err != nil {
		return err
	}

	return nil
}

func (p *PostgresStorage) GetOrders(ctx context.Context, username string) ([]models.Order, error) {
	var rows pgx.Rows
	var err error

	query := `SELECT order_id, status, accrual, uploaded_at, username
				FROM orders 
                WHERE username = $1 
                ORDER BY uploaded_at DESC`

	if retryErr := p.executeWithRetry(ctx, func() error {
		rows, err = p.pool.Query(ctx, query, username)
		if err != nil {
			return fmt.Errorf("error getting orders: %w", err)
		}
		return nil
	}); retryErr != nil {
		return nil, retryErr
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Order])
}

func (p *PostgresStorage) GetOrder(ctx context.Context, orderID models.OrderID) (*models.Order, error) {
	query := `SELECT order_id, status, accrual, uploaded_at, username
				FROM orders 
                WHERE order_id = $1`

	var order models.Order

	err := p.executeWithRetry(ctx, func() error {
		return p.pool.QueryRow(ctx, query, orderID).Scan(&order.OrderID, &order.Status, &order.Accrual, &order.UploadedAt, &order.Username)
	})

	if err != nil {
		return nil, fmt.Errorf("error getting order: %v", err)
	}

	return &order, nil
}

func (p *PostgresStorage) PostOrder(ctx context.Context, order *models.Order) (err error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}

	query := `INSERT INTO orders (order_id, status, accrual, uploaded_at, username) 
	VALUES (@order_id, @status, @accrual, @uploaded_at, @username) 
	ON CONFLICT (order_id, username)  DO UPDATE SET 
	status = @status, 
    accrual = @accrual`

	if err := p.executeTxWithRetry(ctx, tx, query, order.ToNamedArgs(), false); err != nil {
		return err
	}
	return nil

}

func (p *PostgresStorage) GetBalance(ctx context.Context, username string) (*models.Balance, error) {
	query := `SELECT username, current, withdrawn
				FROM balance
                WHERE username = $1`

	var balance models.Balance
	var withdrawn sql.NullFloat64
	var current sql.NullFloat64

	err := p.executeWithRetry(ctx, func() error {
		return p.pool.QueryRow(ctx, query, username).Scan(&balance.Username, &current, &withdrawn)
	})

	if err != nil {
		return nil, err
	}

	if withdrawn.Valid {
		balance.Withdrawn = withdrawn.Float64
	}
	if current.Valid {
		balance.Current = current.Float64
	}

	return &balance, nil

}

func (p *PostgresStorage) PostWithdraw(ctx context.Context, withdraw *models.Withdraw) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}

	query := `
			WITH balance_cte AS (
			  SELECT 
				current
			  FROM balance
			  WHERE username=@username
			)
			INSERT INTO withdrawals (username, order_id, sum, processed_at)
			SELECT @username, @order_id, @sum, @processed_at
				FROM balance_cte
			WHERE balance_cte.current >= @sum 
			ON CONFLICT (order_id, username)  DO NOTHING`

	if err := p.executeTxWithRetry(ctx, tx, query, withdraw.ToNamedArgs(), true); err != nil {
		return err
	}

	return nil
}

func (p *PostgresStorage) GetWithdrawals(ctx context.Context, username string) ([]models.Withdraw, error) {
	var rows pgx.Rows
	var err error

	query := `SELECT order_id, sum, processed_at, username
				FROM withdrawals 
                WHERE username = $1 
                ORDER BY processed_at DESC`

	retryErr := p.executeWithRetry(ctx, func() error {
		rows, err = p.pool.Query(ctx, query, username)
		if err != nil {
			return fmt.Errorf("error getting witdrawals: %w", err)
		}
		return nil
	})
	if retryErr != nil {
		return nil, retryErr
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Withdraw])
}
