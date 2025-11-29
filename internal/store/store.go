package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nobuww/simpel-ktp/internal/store/pg_store"
)

type Store struct {
	*pg_store.Queries
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{
		Queries: pg_store.New(db),
		db:      db,
	}
}

func (s *Store) ExecTx(ctx context.Context, fn func(*pg_store.Queries) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	q := pg_store.New(tx)
	if err := fn(q); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return errors.Join(err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}
