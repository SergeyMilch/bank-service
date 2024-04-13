package db

import (
	"context"

	"github.com/SergeyMilch/bank-service/pkg/models"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DB struct {
    Pool *pgxpool.Pool
}

func NewDB(pool *pgxpool.Pool) *DB {
    return &DB{Pool: pool}
}

// Begin возвращает интерфейс Transaction
func (db *DB) Begin(ctx context.Context) (Transaction, error) {
    tx, err := db.Pool.Begin(ctx)
    if err != nil {
        return nil, err
    }
    return &TransactionAdapter{Tx: tx}, nil
}

// TransactionAdapter адаптирует pgx.Tx к интерфейсу Transaction
type TransactionAdapter struct {
    Tx pgx.Tx
}

func (t *TransactionAdapter) Commit(ctx context.Context) error {
    return t.Tx.Commit(ctx)
}

func (t *TransactionAdapter) Rollback(ctx context.Context) error {
    return t.Tx.Rollback(ctx)
}

func (t *TransactionAdapter) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
    return t.Tx.QueryRow(ctx, sql, args...)
}

func (t *TransactionAdapter) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
    return t.Tx.Exec(ctx, sql, arguments...)
}

func (db *DB) GetAccount(ctx context.Context, tx Transaction, id int) (*models.Account, error) {
    account := &models.Account{}
    row := tx.QueryRow(ctx, "SELECT id, balance FROM bank_account WHERE id = $1", id)
    err := row.Scan(&account.ID, &account.Balance)
    if err != nil {
        return nil, err
    }
    return account, nil
}

func (db *DB) UpdateAccount(ctx context.Context, tx Transaction, id int, balance float64) error {
    _, err := tx.Exec(ctx, "UPDATE bank_account SET balance = $1 WHERE id = $2", balance, id)
    return err
}
