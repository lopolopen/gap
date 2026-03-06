package repo

import "github.com/lopolopen/gap"

type OrderRepo interface {
	Bind(txer gap.Txer) (OrderRepo, error)

	Create(order any) error
}
