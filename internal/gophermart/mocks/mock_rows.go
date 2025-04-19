package mocks

import (
	"errors"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type MockRows struct {
	data  []entity.Balance
	index int
}

func (r *MockRows) Close() {
	r.index = -1
}

func (r *MockRows) Next() bool {
	if r.index+1 < len(r.data) {
		r.index++
		return true
	}
	return false
}

func (r *MockRows) Scan(dest ...interface{}) error {
	if r.index < 0 || r.index >= len(r.data) {
		return errors.New("no rows")
	}
	b := r.data[r.index]
	if len(dest) == 2 {
		dest[0] = &b.Current
		dest[1] = &b.Withdrawn
	}
	return nil
}

func (r *MockRows) Values() ([]interface{}, error) {
	if r.index < 0 || r.index >= len(r.data) {
		return nil, errors.New("no rows")
	}
	b := r.data[r.index]
	return []interface{}{b.Current, b.Withdrawn}, nil
}

func (r *MockRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (r *MockRows) RawValues() [][]byte {
	return nil
}

func (r *MockRows) Conn() *pgx.Conn {
	return nil
}
