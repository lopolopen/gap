package xgorm

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lopolopen/gap/internal/entity"
	"github.com/lopolopen/gap/internal/enum"
	"github.com/lopolopen/gap/internal/errx"
	"github.com/lopolopen/gap/internal/tx"
	"github.com/lopolopen/gap/options/gap"
	"github.com/lopolopen/gap/storage"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type TableCommentter interface {
	TableComment() string
}

type Storage struct {
	gapOpts *gap.Options
	opts    *Options
	db      *gorm.DB
}

// CreatePublished implements [storage.Storage].
func (s *Storage) CreatePublished(ctx context.Context, envelope *entity.Envelope) error {
	if envelope == nil {
		return errx.ErrParamIsNil("envelope")
	}

	if _, ok := s.db.Statement.ConnPool.(gorm.TxCommitter); !ok {
		slog.Warn("publishing does not work in transaction")
	}

	pub := new(Published).FromEntity(envelope)
	err := s.db.Create(pub).Error
	return err
}

func (s *Storage) CreateReceived(ctx context.Context, envelope *entity.Envelope) error {
	if envelope == nil {
		return errx.ErrParamIsNil("envelope")
	}

	rec := new(Received).FromEntity(envelope)
	err := s.db.Create(rec).Error
	return err
}

func NewStorage(gapOpts *gap.Options, db *gorm.DB) *Storage {
	sp := gapOpts.StoragePlugin
	stor := &Storage{
		gapOpts: gapOpts,
		opts:    sp.(*Options),
	}
	stor.setDB(db)

	var _ storage.Storage = stor
	return stor
}

func (s *Storage) setDB(db *gorm.DB) {
	conf := *db.Config
	namer := schema.NamingStrategy{
		SingularTable: true,
	}
	if db.Dialector.Name() == "mysql" {
		namer.TablePrefix = s.opts.Schema + "_"
	}
	conf.NamingStrategy = namer
	if s.opts.LogLevel != 0 {
		conf.Logger = logger.Default.LogMode(logger.LogLevel(s.opts.LogLevel))
	}
	db.Config = &conf
	s.db = db
}

func (s *Storage) init() error {
	return migrateWithComment(s.db,
		&Published{},
		&Received{},
	)
}

func (s *Storage) Bind(txer tx.Txer) (storage.Storage, error) {
	db, ok := txer.Tx().(*gorm.DB)
	if !ok {
		return nil, errx.ErrInvalidGormTx
	}
	newStor := Storage{
		gapOpts: s.gapOpts,
		opts:    s.opts,
	}
	newStor.setDB(db)
	return &newStor, nil
}

func (s *Storage) UpdateStatusPublished(ctx context.Context, id uint, status enum.Status) error {
	return s.db.WithContext(ctx).
		Model(&Published{}).
		Where("`id` = ?", id).
		Update("`status`", status).Error
}

func migrateWithComment(db *gorm.DB, models ...any) error {
	for _, model := range models {
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(model); err != nil {
			return err
		}
		tableName := stmt.Schema.Table

		if err := db.AutoMigrate(model); err != nil {
			return err
		}

		var tableComment string
		if tc, ok := model.(TableCommentter); ok {
			tableComment = tc.TableComment()
		}

		if tableComment != "" {
			if db.Dialector.Name() == "mysql" {
				sql := fmt.Sprintf("ALTER TABLE `%s` COMMENT = '%s'", tableName, tableComment)
				if err := db.Exec(sql).Error; err != nil {
					return err
				}
			} else {
				panic("unimplemented")
			}
		}
	}
	return nil
}

func (s *Storage) ClaimPublishedBatch(ctx context.Context, batchSize int) ([]*entity.Envelope, error) {
	o := s.gapOpts
	var pubs []*Published
	// ago := time.Now().Add(-time.Second * o.Lookback())
	s.db.Transaction(func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).
			Model(&Published{}).
			Where("`version` = ?", o.Version).
			Where("`status` IN ?", []enum.Status{enum.StatusPending, enum.StatusFailed}).
			Where("`retries` < ?", o.MaxRetries).
			// Where("created_at > ?", ago).
			Limit(batchSize).
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Scan(&pubs).Error
		if err != nil {
			return err
		}
		if len(pubs) == 0 {
			return nil
		}
		var ids []uint
		for _, p := range pubs {
			ids = append(ids, p.ID)
		}
		err = tx.WithContext(ctx).
			Model(&Published{}).
			Where("`id` IN ?", ids).
			Update("`status`", enum.StatusProcessing).Error
		if err != nil {
			return err
		}
		return nil
	})
	var es []*entity.Envelope
	for _, p := range pubs {
		es = append(es, p.ToEntity())
	}
	return es, nil
}

func (s *Storage) ClaimReceivedBatch(ctx context.Context, batchSize int) ([]*entity.Envelope, error) {
	o := s.gapOpts
	var recs []*Received
	// ago := time.Now().Add(-time.Second * o.Lookback())
	s.db.Transaction(func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).
			Model(&Received{}).
			Where("`version` = ?", o.Version).
			Where("`group` = ?", o.Group).
			Where("`status` IN ?", []enum.Status{enum.StatusPending, enum.StatusFailed}).
			Where("`retries` < ?", o.MaxRetries).
			// Where("created_at > ?", ago).
			Limit(batchSize).
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Scan(&recs).Error
		if err != nil {
			return err
		}
		if len(recs) == 0 {
			return nil
		}
		var ids []uint
		for _, p := range recs {
			ids = append(ids, p.ID)
		}
		err = tx.WithContext(ctx).
			Model(&Received{}).
			Where("`id` IN ?", ids).
			Update("`status`", enum.StatusProcessing).Error
		if err != nil {
			return err
		}
		return nil
	})
	var es []*entity.Envelope
	for _, p := range recs {
		es = append(es, p.ToEntity())
	}
	return es, nil
}

func (s *Storage) UpdateStatusReceived(ctx context.Context, id uint, status enum.Status) error {
	return s.db.WithContext(ctx).
		Model(&Received{}).
		Where("`id` = ?", id).
		Update("`status`", status).Error
}
