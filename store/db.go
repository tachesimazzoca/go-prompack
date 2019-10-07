package store

import (
	"context"
	"database/sql"
	"errors"
	"sync"
)

type dbStore struct {
	db     *sql.DB
	closed bool
	mux    sync.Mutex
}

func NewDBStore(db *sql.DB) *dbStore {
	return &dbStore{db: db, closed: false}
}

func (st *dbStore) Query(ctx context.Context, s string) ([][]string, error) {
	if st.closed {
		return nil, errors.New("already closed")
	}
	conn, err := st.db.Conn(ctx)
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

func (st *dbStore) Close() error {
	defer st.mux.Unlock()
	st.mux.Lock()
	st.closed = true
	return st.db.Close()
}
