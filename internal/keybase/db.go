package keybase

import (
	"database/sql"

	"github.com/keybase/managed-bots/base"
)

type DB struct {
	*base.BaseOAuthDB
}

func NewDB(db *sql.DB) *DB {
	return &DB{
		BaseOAuthDB: base.NewBaseOAuthDB(db),
	}
}
