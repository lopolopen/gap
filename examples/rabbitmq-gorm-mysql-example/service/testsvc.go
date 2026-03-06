package service

import (
	"context"
	"examples/rabbitmq-gorm-mysql-example/event"
	"examples/rabbitmq-gorm-mysql-example/repo"

	"github.com/lopolopen/gap"
)

//go:generate go tool gapc

type TestSvc struct {
	tx        gap.Tx
	pub       gap.EventPublisher
	orderRepo repo.OrderRepo
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func (s *TestSvc) Run() {
	ctx := context.Background()
	s.tx.DoInTx(ctx, func(ctx context.Context, txer gap.Txer) error {
		pub := must(s.pub.Bind(txer))
		orderRepo := must(s.orderRepo.Bind(txer))

		err := orderRepo.Create("...")
		if err != nil {
			return err
		}

		err = pub.Publish(ctx, event.OrderCreated{
			SN: "x123456",
		}, nil)
		if err != nil {
			return err
		}

		return nil
	})
}

// @subscribe
func (s *TestSvc) SubscribeEvent(event event.OrderCreated) error {
	return nil
}
