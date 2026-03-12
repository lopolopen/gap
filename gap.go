package gap

import (
	"github.com/lopolopen/gap/internal"
	"github.com/lopolopen/gap/internal/tx"
	"github.com/lopolopen/gap/options/gap"
)

const (
	KeysMessageID     = internal.KeysMessageID
	KeysMessageType   = internal.KeysMessageType
	KeysGroup         = internal.KeysGroup
	KeysCorrelationID = internal.KeysCorrelationID
)

var Pair = internal.Pair

type Tx = tx.Tx

type Txer = tx.Txer

type Publisher[T any] = internal.Publisher[T]

type Event = internal.Event

type EventPublisher = internal.EventPublisher

type Handler[T any] = gap.Handler[T]

type Options = gap.Options
