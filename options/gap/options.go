package gap

import (
	"context"
	"time"

	"github.com/lopolopen/gap/internal/enum"
	"github.com/lopolopen/gap/options/dashboard"
	"github.com/lopolopen/gap/options/gorm"
	"github.com/lopolopen/gap/options/kafka"
	"github.com/lopolopen/gap/options/mysql"
	"github.com/lopolopen/gap/options/rabbitmq"
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

	_brokerExt           enum.ExtType
	_rabitmq             *rabbitmq.Options
	_kafa                *kafka.Options
	_storageExt          enum.ExtType
	_gorm                *gorm.Options
	_mysql               *mysql.Options
	_dashboard           *dashboard.Options
	_handlers            []HandlerOptions
	_dependencies        []DIOptions
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

func (o *Options) BrokerExt() enum.ExtType {
	return o._brokerExt
}

func (o *Options) RabbitMQ() *rabbitmq.Options {
	return o._rabitmq
}

func (o *Options) Kafka() *kafka.Options {
	return o._kafa
}

func (o *Options) StorageExt() enum.ExtType {
	return o._storageExt
}

func (o *Options) Gorm() *gorm.Options {
	return o._gorm
}

func (o *Options) MySQL() *mysql.Options {
	return o._mysql
}

func (o *Options) Dashboard() *dashboard.Options {
	return o._dashboard
}

func (o *Options) Handlers() []HandlerOptions {
	return o._handlers
}

func (o *Options) Dependencies() []DIOptions {
	return o._dependencies
}

func (o *Options) Values() []any {
	return o._values
}

func UseRabbitMQ(opts ...shoot.Option[rabbitmq.Options, *rabbitmq.Options]) shoot.Option[Options, *Options] {
	return func(o *Options) {
		options := new(rabbitmq.Options).With(opts...)
		o._rabitmq = options
		o._brokerExt = enum.RabbitMQ
	}
}

func UseKafka(opts ...shoot.Option[kafka.Options, *kafka.Options]) shoot.Option[Options, *Options] {
	return func(o *Options) {
		options := new(kafka.Options).With(opts...)
		o._kafa = options
		o._brokerExt = enum.Kafka
	}
}

func UseGorm(opts ...shoot.Option[gorm.Options, *gorm.Options]) shoot.Option[Options, *Options] {
	return func(o *Options) {
		options := new(gorm.Options).With(opts...)
		o._gorm = options
		o._storageExt = enum.GORM
	}
}

func UseMySQL(opts ...shoot.Option[mysql.Options, *mysql.Options]) shoot.Option[Options, *Options] {
	return func(o *Options) {
		options := new(mysql.Options).With(opts...)
		o._mysql = options
		o._storageExt = enum.MySQL
	}
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
