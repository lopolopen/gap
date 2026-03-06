package tx

import "context"

type Txer interface {
	Tx() any
}

type Tx interface {
	Txer

	DoInTx(ctx context.Context, action func(ctx context.Context, txer Txer) error) error
}
