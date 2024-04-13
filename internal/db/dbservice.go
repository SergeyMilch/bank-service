package db

import (
	"context"

	"github.com/SergeyMilch/bank-service/pkg/models"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type DBService interface {
    Begin(ctx context.Context) (Transaction, error)
    GetAccount(ctx context.Context, tx Transaction, id int) (*models.Account, error)
    UpdateAccount(ctx context.Context, tx Transaction, id int, balance float64) error
}

// Transaction определяет интерфейс для транзакций базы данных
type Transaction interface {
    Commit(ctx context.Context) error
    Rollback(ctx context.Context) error
    QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
    Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}
