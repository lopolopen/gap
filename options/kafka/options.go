package kafka

import "github.com/lopolopen/shoot"

//go:generate go tool shoot new -opt -short -type=Options,TopicOptions

type Options struct {
	Password string

	UserName string

	TopicOpts *TopicOptions
}

type TopicOptions struct {
	NumPartitions int

	ReplicationFactor int16
}

func ConfigTopic(opts ...shoot.Option[TopicOptions, *TopicOptions]) shoot.Option[Options, *Options] {
	return func(o *Options) {
		options := new(TopicOptions).With(opts...)
		o.TopicOpts = options
	}
}
