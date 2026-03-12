package rabbitmq

import (
	"log/slog"

	"github.com/lopolopen/gap/options/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ChanPool interface {
	Rent() (*amqp.Channel, error)
}

type DefaultPool struct {
	opts *rabbitmq.Options
	conn *amqp.Connection
}

func NewDefaultPool(opts *rabbitmq.Options) *DefaultPool {
	return &DefaultPool{
		opts: opts,
	}
}

func (p *DefaultPool) connection() (*amqp.Connection, error) {
	if p.conn != nil && !p.conn.IsClosed() {
		return p.conn, nil
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
	conn, err := amqp.Dial(p.opts.AmqpURL())
	if err != nil {
		slog.Error("failed to connect rabbitmq server", slog.Any("err", err))
		return nil, err
	}
	p.conn = conn
	return conn, nil
}

func (p *DefaultPool) Rent() (*amqp.Channel, error) {
	conn, err := p.connection()
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if p.opts.ConfirmMode {
		err := ch.Confirm(false)
		if err != nil {
			return nil, err
		}
	}
	return ch, err
}
