package internal

import (
	"context"
	"fmt"
	"reflect"

	"github.com/lopolopen/gap/internal/broker"
	"github.com/lopolopen/gap/internal/entity"
	"github.com/lopolopen/gap/internal/errx"
	"github.com/lopolopen/gap/internal/storage"
	"github.com/lopolopen/gap/internal/tx"
)

type Pub[T any] struct {
	opts    *Options
	storage storage.Storage
	broker  broker.Broker
}

func NewPub[T any](opts *Options, storage storage.Storage, broker broker.Broker) *Pub[T] {
	if broker == nil {
		panic(errx.ErrNoBroker)
	}

	pub := &Pub[T]{
		opts:    opts,
		storage: storage,
		broker:  broker,
	}
	var _ Publisher[T] = pub
	return pub
}

// Bind implements [Pub].
func (p *Pub[T]) Bind(txer tx.Txer) (Publisher[T], error) {
	return p.bind(txer)
}

func (p *Pub[T]) bind(txer tx.Txer) (*Pub[T], error) {
	if p.storage == nil {
		return nil, fmt.Errorf("cannot bind txer: %v", errx.ErrNoStorage)
	}
	stor, err := p.storage.Bind(txer)
	if err != nil {
		return nil, err
	}
	return &Pub[T]{
		opts:    p.opts,
		storage: stor,
		broker:  p.broker,
	}, nil
}

// Publish implements [Pub].
func (p *Pub[T]) Publish(ctx context.Context, topic string, msg T, args ...any) error {
	var hds Headers
	hds.Add(args...)

	e := entity.NewEnvelope(p.opts.Version, topic, msg)
	idstr := e.IDString()
	typ := reflect.TypeOf(msg)
	hds.Add(
		Pair(KeysMessageID, idstr),
		Pair(KeysCorrelationID, idstr),
		Pair(KeysMessageType, typ.String()),
	)

	headers := hds.Value()
	for k, v := range headers {
		e.AddHeader(k, v)
	}
	if err := e.Verify(); err != nil {
		return err
	}
	if p.storage != nil {
		return p.storage.CreatePublished(ctx, e)
	}

	err := p.broker.Send(ctx, e)
	if err != nil {
		return err
	}
	return nil
}

type EventPub struct {
	*Pub[Event]
}

func (e *EventPub) Bind(txer tx.Txer) (EventPublisher, error) {
	p, err := e.Pub.bind(txer)
	if err != nil {
		return nil, err
	}
	return &EventPub{Pub: p}, nil
}

func (p *EventPub) Publish(ctx context.Context, event Event, args ...any) error {
	return p.Pub.Publish(ctx, event.Topic(), event, args...)
}
