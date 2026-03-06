package gorm

import (
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//go:generate go tool shoot new -opt -short -type=Options

type Options struct {
	//shoot: def="gap"
	Schema string

	LogLevel logger.LogLevel

	DB *gorm.DB

	MySQL *MySQLConf

	PostgreSQL *PostgreSQLConf
}

type MySQLConf struct {
	DSN string
}

type PostgreSQLConf struct{}
