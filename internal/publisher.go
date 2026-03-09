package internal

import (
	"context"

	"github.com/lopolopen/gap/internal/tx"
)

type Publisher[T any] interface {
	Bind(txer tx.Txer) (Publisher[T], error)

	Publish(ctx context.Context, topic string, msg T, args ...any) error
}

type EventPublisher interface {
	Bind(txer tx.Txer) (EventPublisher, error)

	Publish(ctx context.Context, event Event, args ...any) error
}
