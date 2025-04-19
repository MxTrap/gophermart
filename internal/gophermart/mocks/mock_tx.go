package mocks

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type MockTx struct {
	CommitFn   func(ctx context.Context) error
	RollbackFn func(ctx context.Context) error
	ExecFn     func(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

func (m *MockTx) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

func (m *MockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

func (m *MockTx) Commit(ctx context.Context) error {
	if m.CommitFn != nil {
		return m.CommitFn(ctx)
	}
	return nil
}

func (m *MockTx) Rollback(ctx context.Context) error {
	if m.RollbackFn != nil {
		return m.RollbackFn(ctx)
	}
	return nil
}

// Реализуем остальные методы pgx.Tx как заглушки, если они не нужны
func (m *MockTx) Begin(ctx context.Context) (pgx.Tx, error) {
	return nil, nil
}
func (m *MockTx) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (m *MockTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}
func (m *MockTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return nil
}
func (m *MockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return nil
}
func (m *MockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (m *MockTx) Conn() *pgx.Conn {
	return nil
}
