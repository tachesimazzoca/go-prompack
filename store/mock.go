package store

import (
	"context"
	"errors"
)

type mockStore struct {
	f      func(s string) ([][]string, error)
	closed bool
}

func NewMockStore(f func(s string) ([][]string, error)) *mockStore {
	return &mockStore{f: f, closed: false}
}

func (st *mockStore) Query(ctx context.Context, s string) ([][]string, error) {
	if st.closed {
		return nil, errors.New("already closed")
	}
	return st.f(s)
}

func (st *mockStore) Close() error {
	st.closed = true
	return nil
}
