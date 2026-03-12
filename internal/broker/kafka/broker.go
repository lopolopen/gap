package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/lopolopen/gap/internal/broker"
	"github.com/lopolopen/gap/internal/entity"
	"github.com/lopolopen/gap/internal/errx"
	"github.com/lopolopen/gap/options/gap"
	optkfk "github.com/lopolopen/gap/options/kafka"
	"github.com/segmentio/kafka-go"
)

const (
	// Topic creation default timeouts
	topicCreateTimeout = 30 * time.Second
	topicCreateRetries = 5

	// Topic validation retry backoff
	topicValidateDelay = 500 * time.Millisecond
)

type Broker struct {
	gapOpts        *gap.Options
	opts           *optkfk.Options
	connFactory    *ConnFactory
	writer         *kafka.Writer
	reader         *kafka.Reader
	groupID        string
	topics         []string
	topicMu        sync.RWMutex
	createdTopics  map[string]bool
	createdTopicMu sync.Mutex
}

// Send implements [gap.Broker].
func (b *Broker) Send(ctx context.Context, envelope *entity.Envelope) error {
	if envelope == nil {
		return errx.ErrParamIsNil("envelope")
	}
	if envelope.Topic == "" {
		return errx.ErrEmptyTopic
	}
	body, err := envelope.PayloadBytes()
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return errx.ErrNilPayload
	}

	if err := b.ensureTopic(ctx, envelope.Topic); err != nil {
		return err
	}

	return b.send(ctx, envelope.Topic, envelope.Headers, body)
}

func (b *Broker) send(ctx context.Context, topic string, headers map[string]string, body []byte) error {
	hds := make([]kafka.Header, 0, len(headers))
	for k, v := range headers {
		hds = append(hds, kafka.Header{Key: k, Value: []byte(v)})
	}

	msg := kafka.Message{
		Topic:   topic,
		Value:   body,
		Headers: hds,
	}

	return b.writer.WriteMessages(ctx, msg)
}

// Subscribe implements [gap.Broker].
func (b *Broker) Subscribe(ctx context.Context, topic string) error {
	ctx, cancel := context.WithTimeout(ctx, topicCreateTimeout)
	defer cancel()

	if err := b.ensureTopic(ctx, topic); err != nil {
		return err
	}

	b.topicMu.Lock()
	defer b.topicMu.Unlock()

	if slices.Contains(b.topics, topic) {
		return nil
	}

	b.topics = append(b.topics, topic)
	slog.Debug(fmt.Sprintf("subscribed to topic: %s (%s)", topic, b.gapOpts.Group))
	return nil
}

// Receive implements [gap.Broker].
func (b *Broker) Receive(ctx context.Context) (<-chan *entity.Envelope, error) {
	b.topicMu.RLock()
	topics := make([]string, len(b.topics))
	copy(topics, b.topics)
	b.topicMu.RUnlock()

	if len(topics) == 0 {
		return nil, fmt.Errorf("no topics subscribed")
	}

	// If reader was created without topics, reconfigure it with subscribed topics
	if b.reader == nil {
		brokers := b.opts.Brokers
		b.reader = kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			GroupID:        b.mkGroupID(),
			GroupTopics:    topics,
			StartOffset:    kafka.LastOffset,
			CommitInterval: 0,
		})
	}

	enCh := make(chan *entity.Envelope)

	go func() {
		defer close(enCh)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// b.reader.FetchMessage(ctx)
			m, err := b.reader.ReadMessage(ctx)
			if err != nil {
				slog.Error("read message failed", slog.Any("err", err))
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			en := entity.NewEnvelope(b.gapOpts.Version, m.Topic, nil).
				WithPayload(m.Value).
				WithGroup(b.gapOpts.Group).
				WithTag(m.Offset)

			for _, h := range m.Headers {
				en.AddHeader(h.Key, string(h.Value))
			}

			enCh <- en
		}
	}()

	return enCh, nil
}

// Commit implements [gap.Broker].
func (b *Broker) Commit(tag any) error {
	// Kafka offset is managed automatically by the consumer group
	// The reader handles offset management internally with CommitInterval
	// This is a no-op since kafka-go handles auto-commit
	return nil
}

// Reject implements [gap.Broker].
func (b *Broker) Reject(tag any) error {
	// Kafka doesn't have a built-in reject mechanism like RabbitMQ
	// In practice, this would require storing the offset and potentially re-reading
	// For now, we just log it - actual handling depends on your use case
	slog.Warn("reject called for kafka client, no built-in reject available")
	offset, ok := tag.(int64)
	if !ok {
		return fmt.Errorf("invalid tag type for kafka: %T", tag)
	}
	slog.Info("rejected message at offset", slog.Int64("offset", offset))
	return nil
}

// ensureTopic creates a topic if it doesn't exist, with production-grade error handling and retries.
// It uses a cache to avoid redundant create attempts within the same broker instance.
func (b *Broker) ensureTopic(ctx context.Context, topic string) error {
	b.createdTopicMu.Lock()
	if b.createdTopics[topic] {
		b.createdTopicMu.Unlock()
		return nil
	}
	defer b.createdTopicMu.Unlock()

	topicOpts := b.opts.TopicOpts

	for attempt := 0; attempt < topicCreateRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * 100 * time.Millisecond):
			}
		}

		// Try each broker until one succeeds
		conn, err := b.connFactory.CreateConn(true)
		if err != nil {
			continue
		}

		err = conn.CreateTopics(kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     topicOpts.NumPartitions,
			ReplicationFactor: topicOpts.ReplicationFactor,
		})
		conn.Close()
		if err != nil {
			continue
		}

		err = b.waitTopicReady(ctx, topic)
		if err != nil {
			continue
		}

		// Topic created successfully, mark in cache
		b.createdTopics[topic] = true

		slog.Debug("successfully created kafka topic",
			slog.String("topic", topic),
			slog.Int("partitions", topicOpts.NumPartitions),
			slog.Int("replication_factor", int(topicOpts.ReplicationFactor)),
		)
		return nil
	}

	err := errors.New("failed to ensure topic exists")
	slog.Error(err.Error(), slog.String("topic", topic))
	return err
}

// waitTopicReady waits for a topic to be fully replicated and ready for use.
func (b *Broker) waitTopicReady(ctx context.Context, topic string) error {
	const maxWaitTime = 30 * time.Second
	ctx, cancel := context.WithTimeout(ctx, maxWaitTime)
	defer cancel()

	ticker := time.NewTicker(topicValidateDelay)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for topic to be ready: %w", ctx.Err())
		case <-ticker.C:
			conn, err := b.connFactory.CreateConn(false)
			if err != nil {
				continue
			}

			partitions, err := conn.ReadPartitions(topic)
			conn.Close()

			if err == nil && len(partitions) > 0 {
				// Verify all replicas are available
				allReady := true
				for _, partition := range partitions {
					// Check if leader is valid (Leader.ID should be >= 0)
					if len(partition.Replicas) == 0 || partition.Leader.ID < 0 {
						allReady = false
						break
					}
				}
				if allReady {
					return nil
				}
			}
		}
	}
}

// NewBroker creates a new Kafka broker instance.
func NewBroker(gapOpts *gap.Options) *Broker {
	opts := gapOpts.Kafka()
	if opts == nil {
		panic("kafka options not configured")
	}

	b := &Broker{
		gapOpts:       gapOpts,
		opts:          opts,
		connFactory:   NewConnFactory(gapOpts, opts),
		createdTopics: make(map[string]bool),
	}
	b.writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  opts.Brokers,
		Dialer:   b.connFactory.CreaterDialer(),
		Balancer: &kafka.LeastBytes{},
	})

	var _ broker.Broker = b
	return b
}

func (b *Broker) mkGroupID() string {
	if b.groupID == "" {
		b.groupID = fmt.Sprintf("gap.%s.g.%s", b.gapOpts.Version, b.gapOpts.Group)
	}
	return b.groupID
}
