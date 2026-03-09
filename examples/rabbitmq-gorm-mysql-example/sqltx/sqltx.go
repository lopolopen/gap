package sqltx

import (
	"database/sql"

	"github.com/lopolopen/gap"
)

type SqlTx struct {
	tx *sql.Tx
}

func New(tx *sql.Tx) *SqlTx {
	x := &SqlTx{
		tx: tx,
	}
	var _ gap.Txer = x
	return x
}

func (tx *SqlTx) Tx() any {
	return tx.tx
}
