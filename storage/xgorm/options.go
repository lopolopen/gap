package xgorm

import (
	"github.com/lopolopen/gap/internal/enum"
	"github.com/lopolopen/gap/options/gap"
	"github.com/lopolopen/shoot"
)

//go:generate go tool shoot new -opt -short -type=Options

type Options struct {
	//shoot: def="gap"
	Schema string

	LogLevel int

	GormDB any

	MySQL *MySQLConf

	PostgreSQL *PostgreSQLConf
}

func (o *Options) PluginType() enum.PluginType {
	return enum.GORM
}

type MySQLConf struct {
	DSN string
}

type PostgreSQLConf struct{}

func UseGorm(opts ...shoot.Option[Options, *Options]) shoot.Option[gap.Options, *gap.Options] {
	return func(o *gap.Options) {
		options := new(Options).With(opts...)
		o.StoragePlugin = options
	}
}
