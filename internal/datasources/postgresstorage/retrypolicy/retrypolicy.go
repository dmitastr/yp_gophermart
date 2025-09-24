package retrypolicy

import (
	"errors"
	"time"

	"context"

	"github.com/avast/retry-go/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type RetryPolicy struct {
	Attempts      uint
	Delay         time.Duration
	LastErrorOnly bool
}

func NewRetryPolicy(attempts uint, delay time.Duration, lastErrorOnly bool) *RetryPolicy {
	return &RetryPolicy{Attempts: attempts, Delay: delay, LastErrorOnly: lastErrorOnly}
}

func (rp *RetryPolicy) GetOptions(ctx context.Context) []retry.Option {
	return []retry.Option{
		retry.Context(ctx),
		retry.Attempts(rp.Attempts),
		retry.Delay(rp.Delay),
		retry.LastErrorOnly(rp.LastErrorOnly),
		retry.RetryIf(rp.RetryIf),
	}
}
func (rp *RetryPolicy) RetryIf(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return rp.classifyPgError(pgErr)
	}

	return false
}

func (rp *RetryPolicy) classifyPgError(pgErr *pgconn.PgError) bool {
	// Коды ошибок PostgreSQL: https://www.postgresql.org/docs/current/errcodes-appendix.html

	switch pgErr.Code {
	// Класс 08 - Ошибки соединения
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure:
		return false

	// Класс 40 - Откат транзакции
	case pgerrcode.TransactionRollback, // 40000
		pgerrcode.SerializationFailure, // 40001
		pgerrcode.DeadlockDetected:     // 40P01
		return true

	// Класс 57 - Ошибка оператора
	case pgerrcode.CannotConnectNow: // 57P03
		return true

		// Класс 22 - Ошибки данных
	case pgerrcode.DataException,
		pgerrcode.NullValueNotAllowedDataException:
		return false

	// Класс 23 - Нарушение ограничений целостности
	case pgerrcode.IntegrityConstraintViolation,
		pgerrcode.RestrictViolation,
		pgerrcode.NotNullViolation,
		pgerrcode.ForeignKeyViolation,
		pgerrcode.UniqueViolation,
		pgerrcode.CheckViolation:
		return false

	// Класс 42 - Синтаксические ошибки
	case pgerrcode.SyntaxErrorOrAccessRuleViolation,
		pgerrcode.SyntaxError,
		pgerrcode.UndefinedColumn,
		pgerrcode.UndefinedTable,
		pgerrcode.UndefinedFunction:
		return false

	default:
		return false
	}

}
