package postgresstorage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dmitastr/yp_gophermart/internal/config"
	"github.com/dmitastr/yp_gophermart/internal/domain/models"
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
	fmt.Println("Database connection established successfully")

	m, err := migrate.New(
		"file://database/migrations",
		cfg.DatabaseURI)
	if err != nil {
		return nil, err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}
	fmt.Println("Database migration succeeded")

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
	if err != nil {
		return nil, err
	}

	query := `SELECT order_id, status, accrual, uploaded_at, username
				FROM orders 
                WHERE username = $1 
                ORDER BY uploaded_at DESC`

	rows, err := tx.Query(ctx, query, username)
	if err != nil {
		fmt.Printf("error getting orders: %v\n", err)
		err = errors.Join(err, tx.Rollback(ctx))
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Order])
}

func (p *PostgresStorage) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	tx, err := p.pool.Begin(ctx)
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

func (p *PostgresStorage) toNamedArgs(user models.User) pgx.NamedArgs {
	return pgx.NamedArgs{"name": user.Name, "pass": user.Hash, "created_at": time.Now()}
}
