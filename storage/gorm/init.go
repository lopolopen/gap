package gorm

import (
	"errors"
	"log/slog"

	"github.com/lopolopen/gap/internal"
	"github.com/lopolopen/gap/internal/dashboard"
	"github.com/lopolopen/gap/internal/enum"
	"github.com/lopolopen/gap/options/gap"
	gormopt "github.com/lopolopen/gap/options/gorm"
	"github.com/lopolopen/gap/storage"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type factory struct{}

func (f *factory) CreateStorage(gapOpts *gap.Options) (storage.Storage, error) {
	opts := gapOpts.Gorm()
	db, err := makeDB(opts)
	if err != nil {
		return nil, err
	}

	stor := NewStorage(gapOpts, db)
	err = stor.init()
	if err != nil {
		slog.Error("failed to init storage", slog.Any("err", err))
		return nil, err
	}
	return stor, nil
}

func makeDB(opts *gormopt.Options) (*gorm.DB, error) {
	if opts.GormDB != nil {
		gormDB, ok := opts.GormDB.(*gorm.DB)
		if !ok {
			return nil, errors.New("option GormDB must be *gorm.DB instance")
		}
		db := *gormDB
		return &db, nil
	}
	var dial gorm.Dialector
	if opts.MySQL != nil {
		dial = mysql.Open(opts.MySQL.DSN)
	}

	db, err := gorm.Open(dial)
	if err != nil {
		slog.Error("failed to connect database", slog.Any("err", err))
		panic(err)
	}

	return db, nil
}

func init() {
	internal.Register[storage.Factory](enum.GORM, &factory{})
	dashboard.AddMeta(enum.Storage, enum.GORM.String(), "")
}
