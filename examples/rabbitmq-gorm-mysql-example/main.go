package main

import (
	"context"
	"examples/rabbitmq-gorm-mysql-example/event"
	"examples/rabbitmq-gorm-mysql-example/gormtx"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/lopolopen/gap"
	optgorm "github.com/lopolopen/gap/options/gorm"
	"github.com/lopolopen/gap/options/rabbitmq"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dsn := "root:root@tcp(127.0.0.1:3306)/example?charset=utf8mb4&parseTime=True&loc=Local"
	url := "amqp://guest:guest@localhost:5672/example"

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db := must(gorm.Open(mysql.Open(dsn)))

	pub := gap.NewPublisher[time.Time](
		gap.WithContext(ctx),
		gap.UseRabbitMQ(
			rabbitmq.URL(url),
		),
		gap.UseGorm(
			optgorm.DB(db),
		),
	)

	// pub2 := gap.NewEventPublisher(
	// 	gap.WithContext(ctx),
	// 	gap.UseRabbitMQ(
	// 		rabbitmq.URL(url),
	// 	),
	// 	gap.UseGorm(
	// 		optgorm.DB(db),
	// 	),
	// )

	gap.Subscribe(
		gap.WithContext(ctx),
		gap.ServiceName("rabbitmq-gorm-mysql-example.worker"),
		gap.UseRabbitMQ(
			rabbitmq.URL(url),
		),
		gap.UseGorm(
			optgorm.LogLevel(logger.Silent),
			optgorm.MySQL(&optgorm.MySQLConf{
				DSN: dsn,
			}),
		),
		gap.Inject(db),
	)

	go runjob(db, pub)

	// go runjob2(db, pub2)

	<-ctx.Done()
	stop()
}

func runjob(db *gorm.DB, pub gap.Publisher[time.Time]) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	ctx := context.Background()
	for range ticker.C {
		slog.Info("running job...")
		db.Transaction(func(tx *gorm.DB) error {
			pub := must(pub.Bind(gormtx.New(tx)))

			//do biz logic...

			err := pub.Publish(ctx, "topic.test", time.Now(), nil)
			if err != nil {
				return err
			}
			return nil
		})
	}
}

func runjob2(db *gorm.DB, pub gap.EventPublisher) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	ctx := context.Background()
	for range ticker.C {
		slog.Info("running job2...")
		db.Transaction(func(tx *gorm.DB) error {
			pub := must(pub.Bind(gormtx.New(tx)))

			//do biz logic...

			err := pub.Publish(ctx, event.NewOrderCreated(), nil)
			if err != nil {
				return err
			}
			return nil
		})
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

//go:generate go run github.com/lopolopen/gap/cmd/gapc -file=$GOFILE

// @subscribe: topic="topic.test"
func handle(db *gorm.DB) gap.Handler[time.Time] {
	return func(ctx context.Context, msg time.Time, headers map[string]string) error {
		if db != nil {
			slog.Info(fmt.Sprintf("dependency: %T", db))
		}
		slog.Info(fmt.Sprintf("received message: %s", msg))
		return nil
	}
}

// @subscribe: group="group.test"
func handle2() gap.Handler[*event.OrderCreated] {
	return func(ctx context.Context, msg *event.OrderCreated, headers map[string]string) error {
		slog.Info(fmt.Sprintf("received event: %s, %s", msg.Topic(), msg.SN))
		return nil
	}
}
