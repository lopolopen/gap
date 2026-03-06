package kafka

import (
	"context"
	"fmt"
	"log"

	"github.com/lopolopen/gap/internal"
	"github.com/lopolopen/gap/internal/broker"
	"github.com/lopolopen/gap/internal/entity"
	"github.com/lopolopen/gap/internal/errx"
	"github.com/segmentio/kafka-go"
)

type Broker struct {
	gapOpts *internal.Options
	writer  *kafka.Writer
	reader  *kafka.Reader
	topic   string
	groupID string
}

// Send implements [gap.Broker].
func (b *Broker) Send(ctx context.Context, envelope *entity.Envelope) error {
	body, err := envelope.PayloadBytes()
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return errx.ErrNilPayload
	}

	msg := kafka.Message{
		Topic: b.topic,
		Value: body,
		Headers: func() []kafka.Header {
			hs := []kafka.Header{}
			for k, v := range envelope.Headers {
				hs = append(hs, kafka.Header{Key: k, Value: []byte(fmt.Sprint(v))})
			}
			return hs
		}(),
	}

	return b.writer.WriteMessages(ctx, msg)
}

// receive 消费消息，转成 Envelope
func (b *Broker) receive(ctx context.Context) (<-chan *entity.Envelope, error) {
	enCh := make(chan *entity.Envelope)

	go func() {
		defer close(enCh)
		for {
			m, err := b.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("read error: %v", err)
				return
			}
			e := entity.NewEnvelope(m.Topic, nil).
				WithPayload(m.Value).
				WithTag(m.Offset)
			for _, h := range m.Headers {
				e.AddHeader(h.Key, string(h.Value))
			}
			enCh <- e
		}
	}()
	return enCh, nil
}

func NewBroker(gapOpts *internal.Options) *Broker {
	// opts := gapOpts.Kafka()
	b := &Broker{
		// gapOpts: gapOpts,
		// topic:   opts.Topic,
		// groupID: gapOpts.Group,
		// writer: &kafka.Writer{
		// 	Addr:     kafka.TCP(opts.Brokers...),
		// 	Topic:    opts.Topic,
		// 	Balancer: &kafka.LeastBytes{},
		// },
		// reader: kafka.NewReader(kafka.ReaderConfig{
		// 	Brokers: opts.Brokers,
		// 	GroupID: gapOpts.Group,
		// 	Topic:   opts.Topic,
		// }),
	}
	var _ broker.Broker = b
	return b
}
