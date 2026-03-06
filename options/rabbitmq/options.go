package rabbitmq

import (
	"fmt"

	"github.com/lopolopen/shoot"
)

//go:generate go tool shoot new -opt -short -type=Options,QueueOptions

const OptionsExchangeKind = "topic"

type Options struct {
	//shoot: def="guest"
	Password string `yaml:"password"`

	//shoot: def="guest"
	UserName string `yaml:"username"`

	//shoot: def="/"
	VirtualHost string `yaml:"virtual_host"`

	//shoot: def="default"
	Exchange string `yaml:"exchange"`

	//shoot: def="localhost:5672"
	Endpoint string `yaml:"endpoint"`

	URL string `yaml:"url"`

	ConfirmMode bool `yaml:"confirm_mode"`

	//shoot: def=new(QueueOptions).With()
	QueueOpts *QueueOptions
}

type QueueOptions struct {
	//shoot: def=true
	Durable bool

	Exclusive bool

	AutoDelete bool
}

func (o *Options) Url() string {
	if o.URL != "" {
		return o.URL
	}
	return fmt.Sprintf("amqp://%s:%s@%s%s", o.UserName, o.Password, o.Endpoint, o.VirtualHost)
}

// func (o *Options) QueueOptions() *QueueOptions {
// 	return o._queueOpts
// }

func ConfigQueue(opts ...shoot.Option[QueueOptions, *QueueOptions]) shoot.Option[Options, *Options] {
	return func(o *Options) {
		options := new(QueueOptions).With(opts...)
		o.QueueOpts = options
	}
}
