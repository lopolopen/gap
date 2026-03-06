package gap

import (
	"github.com/lopolopen/gap/internal"
	"github.com/lopolopen/gap/internal/broker"
	"github.com/lopolopen/gap/internal/broker/rabbitmq"
	"github.com/lopolopen/gap/internal/errx"
	"github.com/lopolopen/gap/internal/storage"
	"github.com/lopolopen/gap/internal/storage/gorm"
	"github.com/lopolopen/gap/internal/tx"
	"github.com/lopolopen/shoot"
)

type Tx = tx.Tx

type Txer = tx.Txer

type Publisher[T any] = internal.Publisher[T]

type Event = internal.Event

type EventPublisher = internal.EventPublisher

type Handler[T any] = internal.Handler[T]

type Options = internal.Options

func NewPublisher[T any](opts ...shoot.Option[Options, *Options]) Publisher[T] {
	gapOpts := new(Options).With(opts...)

	var stor storage.Storage
	var brok broker.Broker
	if gapOpts.Gorm() != nil {
		stor = gorm.NewStorate(gapOpts)
	}
	if gapOpts.RabbitMQ() != nil {
		brok = rabbitmq.NewBroker(gapOpts)
	}

	if stor != nil && brok != nil {
		pump := internal.NewPump(gapOpts, stor, brok)
		pump.PollingSend()
	}

	pub := internal.NewPub[T](gapOpts, stor, brok)
	return pub
}

func NewEventPublisher(opts ...shoot.Option[Options, *Options]) EventPublisher {
	pub := &internal.EventPub{
		Pub: NewPublisher[Event](opts...).(*internal.Pub[Event]),
	}
	return pub
}

var grouped groupedSubs

func Subscribe(opts ...shoot.Option[Options, *Options]) {
	gapOpts := new(Options).With(opts...)

	ds := gapOpts.Dependencies()
	grouped.dependencyOtps = append(grouped.dependencyOtps, ds...)
	hs := gapOpts.Handlers()
	grouped.handlerOtps = append(grouped.handlerOtps, hs...)

	if internal.RegisterHandlerOnly(gapOpts) {
		return
	}

	for _, dep := range grouped.dependencyOtps {
		dep.Resolve(gapOpts.Values())
	}

	err := grouped.subscribe(gapOpts)
	if err != nil {
		panic(err)
	}

	err = grouped.listeningAll()
	if err != nil {
		panic(err)
	}
}

type groupedSubs struct {
	subMap         map[string]*internal.Sub
	handlerOtps    []internal.HandlerOptions
	dependencyOtps []internal.DIOptions
}

func (g *groupedSubs) subscribe(gapOpts *Options) error {
	for _, o := range g.handlerOtps {
		if o.Handler == nil {
			return errx.ErrNilHandler
		}
		if o.Topic == "" {
			return errx.ErrEmptyTopic
		}
		opt := *gapOpts
		group := o.Group
		if group == "" {
			group = opt.DefaultGroup
		}
		opt.Group = group

		if g.subMap == nil {
			g.subMap = make(map[string]*internal.Sub)
		}
		h, ok := g.subMap[group]
		if !ok {
			var stor storage.Storage
			var brok broker.Broker
			if opt.Gorm() != nil {
				stor = gorm.NewStorate(&opt)
			}
			if opt.RabbitMQ() != nil {
				brok = rabbitmq.NewBroker(&opt)
			}

			h = internal.NewSub(&opt, stor, brok)
			g.subMap[group] = h
		}
		err := h.Subscribe(o.Topic, o.Handler)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *groupedSubs) listeningAll() error {
	for _, sub := range g.subMap {
		err := sub.Listening()
		if err != nil {
			return err
		}
	}
	return nil
}
