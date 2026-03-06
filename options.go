package gap

import (
	"github.com/lopolopen/gap/internal"
	"github.com/lopolopen/shoot"
)

var (
	WithContext = internal.Context

	UseRabbitMQ = internal.UseRabbitMQ
	UseGorm     = internal.UseGorm

	Version      = internal.Version
	ServiceName  = internal.ServiceName
	DefaultGroup = internal.DefaultGroup

	Inject = internal.Inject
)

// func EnableInbox() shoot.Option[Options, *Options] {
// 	return internal.EnableInbox(true)
// }

func HandleTopic[T any](handler Handler[T], topic string) shoot.Option[Options, *Options] {
	return internal.HandleTopicWithinGroup(handler, topic, "")
}

func HandleTopicWithinGroup[T any](group, topic string, handler Handler[T]) shoot.Option[Options, *Options] {
	return internal.HandleTopicWithinGroup(handler, topic, group)
}
