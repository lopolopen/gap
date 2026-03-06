package internal

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/lopolopen/gap/internal/broker"
	"github.com/lopolopen/gap/internal/entity"
	"github.com/lopolopen/gap/internal/enum"
	"github.com/lopolopen/gap/internal/errx"
	"github.com/lopolopen/gap/internal/storage"
	"golang.org/x/sync/errgroup"
)

var stopHandlingOnce sync.Once
var stopSendingOnce sync.Once

type Pump struct {
	gapOpts *Options
	storage storage.Storage
	broker  broker.Broker
}

func NewPump(gapOpts *Options, storage storage.Storage, broker broker.Broker) *Pump {
	if storage == nil {
		panic(errx.ErrNoStorage)
	}
	if broker == nil {
		panic(errx.ErrNoBroker)
	}
	return &Pump{
		gapOpts: gapOpts,
		storage: storage,
		broker:  broker,
	}
}

func (p *Pump) PollingSend() {
	ctx := p.gapOpts.Context
	go p.send(ctx)
}

func (p *Pump) send(ctx context.Context) {
	ticker := time.NewTicker(p.gapOpts.PumpingInterval())
	defer ticker.Stop()

	defer func() {
		if ctx.Err() != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				stopSendingOnce.Do(func() {
					slog.Info("gap pump stopped sending")
				})
			} else {
				slog.Error("gap pump stopped sending with error", slog.Any("err", ctx.Err()))
			}
		}
	}()

	c := p.gapOpts.MaxPublishConcurrency
	for {
		es, err := p.storage.ClaimPublishedBatch(ctx, p.gapOpts.ClaimBatchSize)
		if err != nil {
			slog.Error("failed to claim published messages", slog.Any("err", err))
			select {
			case <-time.After(30 * time.Second):
				continue
			case <-ctx.Done():
				return
			}
		}

		if c <= 1 {
			for _, e := range es {
				if err := p.sendAndUpdateStatus(ctx, e); err != nil {
					select {
					case <-time.After(5 * time.Second):
						continue
					case <-ctx.Done():
						return
					}
				}
			}
		} else {
			eg, ctx := errgroup.WithContext(ctx)
			sem := make(chan struct{}, c)

			for _, e := range es {
				ev := e
				sem <- struct{}{}
				eg.Go(func() error {
					defer func() { <-sem }()
					return p.sendAndUpdateStatus(ctx, ev)
				})
			}
			if err := eg.Wait(); err != nil {
				select {
				case <-time.After(5 * time.Second):
					continue
				case <-ctx.Done():
					return
				}
			}
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

func (p *Pump) sendAndUpdateStatus(ctx context.Context, envelope *entity.Envelope) error {
	err := p.broker.Send(ctx, envelope)
	if err != nil {
		slog.Error("failed to send message", slog.Any("err", err))
		if err := p.storage.UpdateStatusPublished(ctx, envelope.ID, enum.StatusFailed); err != nil {
			slog.Error("falied to set published status to Failed", slog.Any("err", err))
			return err
		}
		return err
	}
	if err := p.storage.UpdateStatusPublished(ctx, envelope.ID, enum.StatusSucceeded); err != nil {
		slog.Warn("falied to set published status to Succeeded", slog.Any("err", err))
	}
	return nil
}

func (p *Pump) PollingHandle(enCh <-chan *entity.Envelope) <-chan *entity.Envelope {
	ctx := p.gapOpts.Context
	go func() {
		for en := range enCh {
			err := p.storage.CreateReceived(ctx, en)
			if err != nil {
				slog.Error("failed to create received", slog.Any("err", err))
				if err := p.broker.Reject(en.Tag); err != nil {
					slog.Error("failed to reject message", slog.Any("err", err))
				}
				continue
			}
			if err := p.broker.Commit(en.Tag); err != nil {
				slog.Error("failed to commit message", slog.Any("err", err))
			}
		}
	}()

	outCh := make(chan *entity.Envelope)
	go p.handle(ctx, outCh)
	return outCh
}

func (p *Pump) handle(ctx context.Context, outCh chan *entity.Envelope) {
	ticker := time.NewTicker(p.gapOpts.PumpingInterval())
	defer ticker.Stop()
	defer close(outCh)

	defer func() {
		if errors.Is(ctx.Err(), context.Canceled) {
			stopHandlingOnce.Do(func() {
				slog.Info("gap pump stopped handling")
			})
		} else {
			slog.Error("gap pump stopped handling with error", slog.Any("err", ctx.Err()))
		}
	}()

	for {
		es, err := p.storage.ClaimReceivedBatch(ctx, p.gapOpts.ClaimBatchSize)
		if err != nil {
			slog.Error("failed to claim received messages", slog.Any("err", err))
			select {
			case <-time.After(30 * time.Second):
			case <-ctx.Done():
				return
			}
			continue
		}

		if len(es) == 0 {
			select {
			case <-time.After(10 * time.Second):
			case <-ctx.Done():
				return
			}
			continue
		}

		for _, e := range es {
			select {
			case outCh <- e:
			case <-ctx.Done():
				return
			}
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}
