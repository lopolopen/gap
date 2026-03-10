package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lopolopen/gap/internal"
	"github.com/lopolopen/gap/internal/broker"
	"github.com/lopolopen/gap/internal/entity"
	"github.com/lopolopen/gap/internal/errx"
	"github.com/lopolopen/gap/options/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Broker struct {
	gapOpts *internal.Options
	opts    *rabbitmq.Options
	chPool  ChanPool
	x       string
	q       string
	ctag    string
}

// Pub implements [gap.Broker].
func (b *Broker) Send(ctx context.Context, envelope *entity.Envelope) error {
	routingKey := envelope.Topic
	body, err := envelope.PayloadBytes()
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return errx.ErrNilPayload
	}
	return b.send(ctx, routingKey, envelope.Headers, envelope.Payload)
}

func (b *Broker) send(ctx context.Context, routingKey string, headers map[string]string, body []byte) error {
	ch, err := b.chPool.Rent()
	if err != nil {
		return err
	}
	defer ch.Close()

	tbl := make(map[string]any)
	for k, v := range headers {
		tbl[k] = v
	}

	// Exchange may not be bound to any queues.
	// In this case:
	// The message will be discarded by the broker.
	// The message in the database will be marked as successfully sent if persistence is enabled.
	err = ch.PublishWithContext(ctx, b.exchange(), routingKey, false, false, amqp.Publishing{
		Headers:      tbl,
		MessageId:    "",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
	if err != nil {
		return err
	}

	return nil
}

func (b *Broker) Subscribe(_ context.Context, topic string) error {
	ch, err := b.chPool.Rent()
	if err != nil {
		return err
	}
	defer ch.Close()

	o := b.opts.QueueOpts
	q, err := ch.QueueDeclare(b.queue(), o.Durable, o.AutoDelete, o.Exclusive, false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(q.Name, topic, b.exchange(), false, nil)
	if err != nil {
		return err
	}
	slog.Debug(fmt.Sprintf("binded to routing key: %s (%s)", topic, b.gapOpts.Group))
	return nil
}

func (b *Broker) Receive(ctx context.Context) (<-chan *entity.Envelope, error) {
	enCh := make(chan *entity.Envelope)

	go func() {
		defer close(enCh)

		for {
			ch, err := b.chPool.Rent()
			var deliCh <-chan amqp.Delivery
			if err != nil {
				slog.Error("rent channel failed", slog.Any("err", err))
				goto RETRY
			}

			deliCh, err = ch.ConsumeWithContext(
				ctx,
				b.queue(),
				b.consumerTag(),
				false, false, false, false, nil,
			)
			if err != nil {
				slog.Error("consume failed", slog.Any("err", err))
				goto RETRY
			}

			for {
				select {
				case <-ctx.Done():
					return
				case deli, ok := <-deliCh:
					if !ok {
						slog.Warn("delivery channel closed, retrying with new channel")
						goto RETRY
					}
					en := entity.NewEnvelope(b.gapOpts.Version, deli.RoutingKey, nil).
						WithPayload(deli.Body).
						WithGroup(b.gapOpts.Group).
						WithTag(deli)

					for k, v := range deli.Headers {
						en.AddHeader(k, fmt.Sprintf("%v", v))
					}

					enCh <- en
				}
			}

		RETRY:
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}
	}()

	return enCh, nil
}

func (b *Broker) Commit(tag any) error {
	deli, _ := tag.(amqp.Delivery)
	return deli.Ack(false)
}

func (b *Broker) Reject(tag any) error {
	deli, _ := tag.(amqp.Delivery)
	return deli.Nack(false, true)
}

func NewBroker(gapOpts *internal.Options) *Broker {
	opts := gapOpts.RabbitMQ()
	if opts == nil {
		panic("rabbitmq options not configured")
	}

	b := &Broker{
		gapOpts: gapOpts,
		opts:    opts,
		chPool:  NewDefaultPool(opts),
	}

	err := b.init()
	if err != nil {
		slog.Error("failed to init broker", slog.Any("err", err))
		panic(err)
	}

	var _ broker.Broker = b
	return b
}

func (b *Broker) init() error {
	ch, err := b.chPool.Rent()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(b.exchange(), rabbitmq.OptionsExchangeKind, true, false, false, false, nil)
	if err != nil {
		return err
	}

	return nil
}

func (b *Broker) exchange() string {
	if b.x == "" {
		b.x = fmt.Sprintf("gap.%s.x.%s", b.gapOpts.Version, b.opts.Exchange)
	}
	return b.x
}

func (b *Broker) queue() string {
	if b.q == "" {
		b.q = fmt.Sprintf("gap.%s.q.%s", b.gapOpts.Version, b.gapOpts.Group)
	}
	return b.q
}

func (b *Broker) consumerTag() string {
	if b.ctag == "" {
		name := b.gapOpts.ServiceName
		if name == "" {
			name = os.Args[0]
		}
		b.ctag = fmt.Sprintf("gap.%s", name)
	}
	return b.ctag
}
