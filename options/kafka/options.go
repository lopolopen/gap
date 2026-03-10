package kafka

import "github.com/lopolopen/shoot"

//go:generate go tool shoot new -opt -short -type=Options,TopicOptions

type Options struct {
	//shoot: def=""
	Password string

	//shoot: def=""
	UserName string

	//shoot: def=[]string{"localhost:9092"}
	Brokers []string

	//shoot: def=new(TopicOptions).With()
	TopicOpts *TopicOptions
}

type TopicOptions struct {
	//shoot: def=-1
	NumPartitions int

	//shoot: def=-1
	ReplicationFactor int
}

func ConfigTopic(opts ...shoot.Option[TopicOptions, *TopicOptions]) shoot.Option[Options, *Options] {
	return func(o *Options) {
		options := new(TopicOptions).With(opts...)
		o.TopicOpts = options
	}
}
