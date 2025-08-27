package postgres_storage

import (
	"context"
	"fmt"
	"time"

	"github.com/dmitastr/yp_gophermart/internal/config"
	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/tracelog"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, cfg *config.Config) (*PostgresStorage, error) {
	dbConfig, err := pgxpool.ParseConfig(cfg.DatabaseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}
	// dbConfig.ConnConfig.Tracer = &tracelog.TraceLog{
	// 	Logger:   logger.GetLogger(),
	// 	LogLevel: tracelog.LogLevelInfo,
	// }
	pool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db with url=%s: %v", cfg.DatabaseURI, err)
	}
	fmt.Println("Database connection established successfully")
	return &PostgresStorage{pool: pool}, nil
}

func (p *PostgresStorage) InsertUser(ctx context.Context, user models.User) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}

	query := `INSERT INTO users (username, password, created_at) VALUES (@name, @pass, @created_at)`
	args := pgx.NamedArgs{"name": user.Name, "pass": user.Hash, "created_at": time.Now()}
	if _, err := tx.Exec(ctx, query, args); err != nil {
		tx.Rollback(ctx)
		return err
	}
	defer tx.Commit(ctx)
	return nil
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
		tx.Rollback(ctx)
		return err
	}
	defer tx.Commit(ctx)

	return err
}

func (p *PostgresStorage) GetOrders(ctx context.Context, username string) ([]models.Order, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, "SELECT order_id, status, accrual, uploaded_at FROM orders WHERE username = $1", username)
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Order])
}

func (p *PostgresStorage) toNamedArgs(user models.User) pgx.NamedArgs {
	return pgx.NamedArgs{"name": user.Name, "pass": user.Hash, "created_at": time.Now()}
}
