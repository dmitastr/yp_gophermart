package postgresstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dmitastr/yp_gophermart/internal/config"
	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"github.com/dmitastr/yp_gophermart/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/tracelog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
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

	return &PostgresStorage{pool: pool}, nil
}

func (p *PostgresStorage) InsertUser(ctx context.Context, user models.User) (err error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}

	query := `INSERT INTO users (username, password, created_at) VALUES (@name, @pass, @created_at)`
	args := pgx.NamedArgs{"name": user.Name, "pass": user.Hash, "created_at": time.Now()}
	if _, err := tx.Exec(ctx, query, args); err != nil {
		err = errors.Join(err, tx.Rollback(ctx))
		return err
	}
	defer func() {
		err = errors.Join(err, tx.Commit(ctx))
	}()

	return
}

func (p *PostgresStorage) GetUser(ctx context.Context, username string) (models.User, error) {
	tx, err := p.pool.Begin(ctx)
	defer func() {
		_ = tx.Commit(ctx)
	}()

	if err != nil {
		return models.User{}, err
	}

	var user models.User
	err = tx.QueryRow(ctx, "SELECT username, password FROM users WHERE username = $1", username).Scan(&user.Name, &user.Password)
	return user, err
}

func (p *PostgresStorage) UpdateUser(ctx context.Context, user models.User) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}

	args := p.toNamedArgs(user)
	if _, err := tx.Exec(ctx, "UPDATE users SET password=@pass WHERE username = @name", args); err != nil {
		err = errors.Join(err, tx.Rollback(ctx))
		return err
	}
	defer func() {
		err = errors.Join(err, tx.Commit(ctx))
	}()

	return err
}

func (p *PostgresStorage) GetOrders(ctx context.Context, username string) ([]models.Order, error) {
	tx, err := p.pool.Begin(ctx)
	defer func() {
		_ = tx.Commit(ctx)
	}()
	if err != nil {
		return nil, err
	}

	query := `SELECT order_id, status, accrual, uploaded_at, username
				FROM orders 
                WHERE username = $1 
                ORDER BY uploaded_at DESC`

	rows, err := tx.Query(ctx, query, username)
	if err != nil {
		logger.Errorf("error getting orders: %v\n", err)
		err = errors.Join(err, tx.Rollback(ctx))
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Order])
}

func (p *PostgresStorage) GetOrder(ctx context.Context, orderID models.OrderID) (*models.Order, error) {
	tx, err := p.pool.Begin(ctx)
	defer func() {
		_ = tx.Commit(ctx)
	}()

	if err != nil {
		return nil, err
	}

	query := `SELECT order_id, status, accrual, uploaded_at, username
				FROM orders 
                WHERE order_id = $1
                ORDER BY uploaded_at DESC`

	var order models.Order
	err = tx.QueryRow(ctx, query, orderID).Scan(&order.OrderID, &order.Status, &order.Accrual, &order.UploadedAt, &order.Username)
	if err != nil {
		err = errors.Join(err, tx.Rollback(ctx))
		return nil, err
	}

	return &order, nil
}

func (p *PostgresStorage) PostOrder(ctx context.Context, order *models.Order) (*models.Order, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO orders (order_id, status, accrual, uploaded_at, username) 
	VALUES (@order_id, @status, @accrual, @uploaded_at, @username) 
	ON CONFLICT (order_id, username)  DO UPDATE SET 
	status = @status, 
    accrual = @accrual`

	if _, err := tx.Exec(ctx, query, order.ToNamedArgs()); err != nil {
		err = errors.Join(err, tx.Rollback(ctx))
		return nil, err
	}
	defer func() {
		err = errors.Join(err, tx.Commit(ctx))
	}()
	return order, nil

}

func (p *PostgresStorage) GetBalance(ctx context.Context, username string) (*models.Balance, error) {
	tx, err := p.pool.Begin(ctx)
	defer func() {
		_ = tx.Commit(ctx)
	}()

	if err != nil {
		return nil, err
	}

	query := `SELECT username, current, withdrawn
				FROM balance
                WHERE username = $1`

	var balance models.Balance
	var withdrawn sql.NullFloat64
	var current sql.NullFloat64
	err = tx.QueryRow(ctx, query, username).Scan(&balance.Username, &current, &withdrawn)
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

	query := `INSERT INTO withdrawals (username, order_id, sum, processed_at) 
	VALUES (@username, @order_id, @sum, @processed_at) 
	ON CONFLICT (order_id, username)  DO NOTHING`

	if _, err := tx.Exec(ctx, query, withdraw.ToNamedArgs()); err != nil {
		err = errors.Join(err, tx.Rollback(ctx))
		return err
	}
	defer func() {
		err = errors.Join(err, tx.Commit(ctx))
	}()

	return nil
}

func (p *PostgresStorage) GetWithdrawals(ctx context.Context, username string) ([]models.Withdraw, error) {
	tx, err := p.pool.Begin(ctx)
	defer func() {
		_ = tx.Commit(ctx)
	}()
	if err != nil {
		return nil, err
	}

	query := `SELECT order_id, sum, processed_at, username
				FROM withdrawals 
                WHERE username = $1 
                ORDER BY processed_at DESC`

	rows, err := tx.Query(ctx, query, username)
	if err != nil {
		logger.Errorf("error getting witdrawals: %v\n", err)
		err = errors.Join(err, tx.Rollback(ctx))
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Withdraw])
}

func (p *PostgresStorage) toNamedArgs(user models.User) pgx.NamedArgs {
	return pgx.NamedArgs{"name": user.Name, "pass": user.Hash, "created_at": time.Now()}
}
