package broker

import (
	"context"

	"github.com/lopolopen/gap/internal/entity"
)

type Broker interface {
	Send(ctx context.Context, envelope *entity.Envelope) error

	Subscribe(topic string) error

	Receive(ctx context.Context) (<-chan *entity.Envelope, error)

	Commit(tag any) error

	Reject(tag any) error
}
