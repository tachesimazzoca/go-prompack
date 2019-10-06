package prompack

import (
	"context"
	"database/sql"
)

type Querier interface {
	Query(s string) ([][]string, error)
}

type mockQuerier struct {
	f func(s string) ([][]string, error)
}

func NewMockQuerier(f func(s string) ([][]string, error)) *mockQuerier {
	return &mockQuerier{f: f}
}

func (q *mockQuerier) Query(s string) ([][]string, error) {
	return q.f(s)
}

type dbQuerier struct {
	db *sql.DB
}

func NewDBQuerier(db *sql.DB) *dbQuerier {
	return &dbQuerier{db: db}
}

func (q *dbQuerier) Query(s string) ([][]string, error) {
	ctx := context.TODO()
	conn, err := q.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rs, err := conn.QueryContext(ctx, s)
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	xs := make([][]string, 0)
	numCols := 0
	for rs.Next() {
		if numCols == 0 {
			if cols, err := rs.Columns(); err != nil {
				return nil, err
			} else {
				numCols = len(cols)
			}
		}
		ys := make([]string, numCols)
		ysp := make([]interface{}, len(ys))
		for i, _ := range ys {
			ysp[i] = &ys[i]
		}
		if err := rs.Scan(ysp...); err != nil {
			return nil, err
		}
		xs = append(xs, ys)
	}
	return xs, nil
}
