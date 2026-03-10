package internal

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lopolopen/gap/internal/broker"
	"github.com/lopolopen/gap/internal/entity"
	"github.com/lopolopen/gap/internal/enum"
	"github.com/lopolopen/gap/internal/errx"
	"github.com/lopolopen/gap/internal/storage"
)

type Handler[T any] func(ctx context.Context, msg T, headers map[string]string) error

type Sub struct {
	opts     *Options
	group    string
	storage  storage.Storage
	broker   broker.Broker
	handlers map[string]Handler[[]byte]
	pump     *Pump
}

// Subscribe implements [Subscriber].
func (s *Sub) Subscribe(topic string, handler Handler[[]byte]) error {
	_, ok := s.handlers[topic]
	if ok {
		return errx.ErrMultiHandlers(topic, s.group)
	}
	s.handlers[topic] = handler

	slog.Debug(fmt.Sprintf("subscribe: topic(%s) -> group(%s)", topic, s.group))

	ctx := context.Background()
	err := s.broker.Subscribe(ctx, topic)
	if err != nil {
		return err
	}

	return nil
}

func NewSub(opts *Options, storage storage.Storage, broker broker.Broker) *Sub {
	if broker == nil {
		panic(errx.ErrNoBroker)
	}

	sub := &Sub{
		opts:     opts,
		group:    opts.Group,
		storage:  storage,
		broker:   broker,
		handlers: make(map[string]Handler[[]byte]),
	}
	if storage != nil {
		sub.pump = NewPump(opts, storage, broker)
	} else {
		slog.Debug("working on no-persistence mode")
	}
	return sub
}

func (s *Sub) Listening() error {
	ctx := s.opts.Context
	enCh, err := s.broker.Receive(ctx)
	if err != nil {
		return err
	}

	if s.pump != nil {
		enCh = s.pump.PollingHandle(enCh)
	}

	go func() {
		for en := range enCh {
			err := s.handleAndUpdateStatues(ctx, en)
			if en.Tag == nil {
				continue
			}
			if err != nil {
				if err := s.broker.Reject(en.Tag); err != nil {
					slog.Error("failed to reject message", slog.Any("err", err))
				}
				continue
			}
			if err := s.broker.Commit(en.Tag); err != nil {
				slog.Error("failed to commit message", slog.Any("err", err))
			}
		}
	}()

	return nil
}

func (s *Sub) handle(ctx context.Context, envelope *entity.Envelope) error {
	handler, ok := s.handlers[envelope.Topic]
	if !ok {
		return errx.ErrHandlerNotFound(envelope.Topic, s.group)
	}
	payload, err := envelope.PayloadBytes()
	if err != nil {
		return err
	}
	err = handler(ctx, payload, envelope.Headers)
	if err != nil {
		return err
	}
	return nil
}

func (s *Sub) handleAndUpdateStatues(ctx context.Context, envelope *entity.Envelope) error {
	err := s.handle(ctx, envelope)
	if err != nil {
		slog.Error("failed to handle message", slog.Any("err", err))
	}
	if s.storage == nil {
		return err
	}

	if err != nil {
		if err := s.storage.UpdateStatusReceived(ctx, envelope.ID, enum.StatusFailed); err != nil {
			slog.Error("failed to set received status to Failed", slog.Any("err", err))
			return err
		}
		return err
	}
	if err := s.storage.UpdateStatusReceived(ctx, envelope.ID, enum.StatusSucceeded); err != nil {
		slog.Warn("falied to set received status to Succeeded", slog.Any("err", err))
	}
	return nil
}
