package gap

import (
	"context"
	"time"

	"github.com/lopolopen/gap/options/dashboard"
	"github.com/lopolopen/shoot"
)

//go:generate go tool shoot new -opt -short -type=Options

type Options struct {
	//shoot: def=context.Background()
	Context context.Context

	ServiceName string

	//shoot: def="v1"
	Version string

	//shoot: def="default"
	DefaultGroup string

	Group string

	//shoot: def=200
	ClaimBatchSize int

	//shoot: def=30
	MaxRetries int

	//shoot: def=180
	LookbackSeconds int

	//shoot: def=1
	PumpingIntervalSeconds int

	//shoot: def=1
	MaxPublishConcurrency int32

	//shoot: def=-1
	WorkerID int64

	StoragePlugin PluginOptions

	BrokerPlugin PluginOptions

	// _brokerExtType       enum.PluginType
	// _brokerOpts          any
	// _storageExtType      enum.PluginType
	// _storageOpts         any
	_dashboard           *dashboard.Options
	_handlers            []HandlerOptions
	_dependencies        []HandlerDepsOptions
	_values              []any
	_registerHandlerOnly bool
}

type HandlerOptions struct {
	Group   string
	Topic   string
	Handler Handler[[]byte]
}

func (o *Options) Lookback() time.Duration {
	return time.Second * (time.Duration(o.LookbackSeconds))
}

func (o *Options) PumpingInterval() time.Duration {
	return time.Second * (time.Duration(o.PumpingIntervalSeconds))
}

// func (o *Options) BrokerExt() enum.ExtType {
// 	return o._brokerExt
// }

// func (o *Options) RabbitMQ() *rabbitmq.Options {
// 	return o._rabitmq
// }

// func (o *Options) Kafka() *kafka.Options {
// 	return o._kafa
// }

// func (o *Options) StorageExt() enum.ExtType {
// 	return o._storageExtType
// }

// func (o *Options) Gorm() *gorm.Options {
// 	return o._gorm
// }

// func (o *Options) MySQL() *mysql.Options {
// 	return o._mysql
// }

// func (o *Options) SetBrokerOptions(typ enum.PluginType, opts any) {
// 	o._brokerExtType = typ
// 	o._brokerOpts = opts
// }

// func (o *Options) BrokenOptions() (enum.PluginType, any) {
// 	return o._brokerExtType, o._brokerOpts
// }

// func (o *Options) SetStorageOptions(typ enum.PluginType, opts any) {
// 	o._storageExtType = typ
// 	o._storageOpts = opts
// }

// func (o *Options) StorageOptions() (enum.PluginType, any) {
// 	return o._storageExtType, o._storageOpts
// }

func (o *Options) Dashboard() *dashboard.Options {
	return o._dashboard
}

func (o *Options) Handlers() []HandlerOptions {
	return o._handlers
}

func (o *Options) Dependencies() []HandlerDepsOptions {
	return o._dependencies
}

func (o *Options) Values() []any {
	return o._values
}

func UseDashboard(opts ...shoot.Option[dashboard.Options, *dashboard.Options]) shoot.Option[Options, *Options] {
	return func(o *Options) {
		options := new(dashboard.Options).With(opts...)
		o._dashboard = options
	}
}

func Inject(values ...any) shoot.Option[Options, *Options] {
	return func(o *Options) {
		o._values = append(o._values, values...)
	}
}

func GoGenerated() shoot.Option[Options, *Options] {
	return func(o *Options) {
		o._registerHandlerOnly = true
	}
}
