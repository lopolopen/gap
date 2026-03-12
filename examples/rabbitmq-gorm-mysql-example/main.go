package main

import (
	"context"
	"examples/rabbitmq-gorm-mysql-example/gormtx"
	"examples/rabbitmq-gorm-mysql-example/service"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/lopolopen/gap"
	optgorm "github.com/lopolopen/gap/options/gorm"
	"github.com/lopolopen/gap/options/rabbitmq"
	_ "github.com/lopolopen/gap/storage/gorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dsn := "root:root@tcp(127.0.0.1:3306)/example?charset=utf8mb4&parseTime=True&loc=Local"
	url := "amqp://guest:guest@localhost:5672/example"

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db := must(gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}))

	pub := gap.NewEventPublisher(
		gap.WithContext(ctx),
		gap.UseRabbitMQ(
			rabbitmq.URL(url),
		),
		gap.UseGorm(
			optgorm.GormDB(db),
		),
	)

	mySvc := service.NewMySvc(gormtx.New(db), pub)

	gap.Subscribe(
		gap.WithContext(ctx),
		gap.UseRabbitMQ(
			rabbitmq.URL(url),
		),
		gap.UseGorm(
			optgorm.MySQL(&optgorm.MySQLConf{
				DSN: dsn,
			}),
		),
		gap.Inject(mySvc),
	)

	go runjob(mySvc)

	<-ctx.Done()
	stop()
}

func runjob(mySvc *service.MySvc) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		slog.Info("running job...")

		mySvc.CreateOrder()
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
