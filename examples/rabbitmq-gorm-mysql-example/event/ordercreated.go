package event

import "github.com/google/uuid"

type OrderCreated struct {
	SN string
}

func NewOrderCreated() *OrderCreated {
	return &OrderCreated{
		SN: uuid.New().String(),
	}
}

func (e OrderCreated) Topic() string {
	return "order.created"
}
