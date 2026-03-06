package storage

import (
	"context"

	"github.com/lopolopen/gap/internal/entity"
	"github.com/lopolopen/gap/internal/enum"
	"github.com/lopolopen/gap/internal/tx"
)

type Storage interface {
	Bind(txer tx.Txer) (Storage, error)

	CreatePublished(ctx context.Context, envelop *entity.Envelope) error

	UpdateStatusPublished(ctx context.Context, id uint, status enum.Status) error

	ClaimPublishedBatch(ctx context.Context, batchSize int) ([]*entity.Envelope, error)

	CreateReceived(ctx context.Context, envelop *entity.Envelope) error

	ClaimReceivedBatch(ctx context.Context, batchSize int) ([]*entity.Envelope, error)

	UpdateStatusReceived(ctx context.Context, id uint, status enum.Status) error
}
