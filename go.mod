module github.com/lopolopen/gap

go 1.24.6

tool github.com/lopolopen/shoot/cmd/shoot

require (
	github.com/bwmarrin/snowflake v0.3.0
	github.com/lopolopen/shoot v0.7.1
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/segmentio/kafka-go v0.4.50
	golang.org/x/sync v0.19.0
	golang.org/x/tools v0.41.0
)

require (
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	golang.org/x/mod v0.32.0 // indirect
)

// replace github.com/lopolopen/shoot => ../shoot
