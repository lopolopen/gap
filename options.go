package gap

import (
	"github.com/lopolopen/gap/options/gap"
	"github.com/lopolopen/shoot"
)

var (
	WithContext = gap.Context

	UseRabbitMQ  = gap.UseRabbitMQ
	UseKafka     = gap.UseKafka
	UseGorm      = gap.UseGorm
	UseMySQL     = gap.UseMySQL
	UseDashboard = gap.UseDashboard

	Version      = gap.Version
	ServiceName  = gap.ServiceName
	DefaultGroup = gap.DefaultGroup
	WorkerID     = gap.WorkerID

	Inject = gap.Inject
)

func HandleTopic[T any](handler gap.Handler[T], topic string) shoot.Option[Options, *Options] {
	return gap.HandleTopicWithinGroup(handler, topic, "")
}

func HandleTopicWithinGroup[T any](group, topic string, handler gap.Handler[T]) shoot.Option[Options, *Options] {
	return gap.HandleTopicWithinGroup(handler, topic, group)
}
