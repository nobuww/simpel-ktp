package store

import (
	"github.com/nobuww/simpel-ktp/internal/store/pg_store"
)

type Repository interface {
	pg_store.Querier
}
