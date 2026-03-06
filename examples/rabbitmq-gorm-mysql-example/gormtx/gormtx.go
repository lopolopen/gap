package gormtx

import (
	"github.com/lopolopen/gap"
	"gorm.io/gorm"
)

type GormTx struct {
	db *gorm.DB
}

func New(db *gorm.DB) *GormTx {
	x := &GormTx{
		db: db,
	}
	var _ gap.Txer = x
	// var _ gap.Tx = x
	return x
}

func (tx *GormTx) Tx() any {
	return tx.db
}

// func (tx *GormTx) DoInTx(ctx context.Context, action func(context.Context, gap.Txer) error) error {
// 	return tx.db.WithContext(ctx).
// 		Transaction(func(tx *gorm.DB) error {
// 			return action(ctx, &GormTx{db: tx})
// 		})
// }
