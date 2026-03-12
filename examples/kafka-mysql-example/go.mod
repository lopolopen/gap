module examples/kafka-mysql-example

go 1.24.6

require (
	github.com/go-sql-driver/mysql v1.8.1
	github.com/lopolopen/gap v0.0.1-alpha.5
	github.com/lopolopen/gap/storage/mysql v0.0.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/bwmarrin/snowflake v0.3.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/lopolopen/shoot v0.7.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
	github.com/segmentio/kafka-go v0.4.50 // indirect
	golang.org/x/sync v0.19.0 // indirect
	gorm.io/gorm v1.31.1 // indirect
)

replace github.com/lopolopen/gap => ../../../gap

replace github.com/lopolopen/gap/storage/mysql => ../../../gap/storage/mysql
