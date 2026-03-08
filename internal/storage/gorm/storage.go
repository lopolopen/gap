package gorm

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/lopolopen/gap/internal"
	"github.com/lopolopen/gap/internal/entity"
	"github.com/lopolopen/gap/internal/enum"
	"github.com/lopolopen/gap/internal/errx"
	"github.com/lopolopen/gap/internal/storage"
	"github.com/lopolopen/gap/internal/tx"
	gormopt "github.com/lopolopen/gap/options/gorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type TableCommentter interface {
	TableComment() string
}

type Storage struct {
	gapOpts *internal.Options
	opts    *gormopt.Options
	db      *gorm.DB
}

// CreatePublished implements [storage.Storage].
func (s *Storage) CreatePublished(ctx context.Context, envelop *entity.Envelope) error {
	if envelop == nil {
		return errx.ErrParamIsNil("envelop")
	}
	if envelop.Topic == "" {
		return errx.ErrEmptyTopic
	}
	payload, err := envelop.PayloadBytes()
	if err != nil {
		return err
	}
	if len(payload) == 0 {
		return errx.ErrNilPayload
	}

	headers, err := envelop.HeadersBytes()
	if err != nil {
		return err
	}

	pub := &Published{
		Model: Model{
			ID:        envelop.ID,
			CreatedAt: time.Now(),
			Version:   s.gapOpts.Version,
			Topic:     envelop.Topic,
			Status:    enum.StatusPending,
			Payload:   string(payload),
			Headers:   string(headers),
		},
	}
	if _, ok := s.db.Statement.ConnPool.(gorm.TxCommitter); !ok {
		slog.Warn("publishing does not work in transaction")
	}
	err = s.db.Create(pub).Error
	if err != nil {
		return err
	}
	envelop.ID = pub.ID
	return nil
}

func (s *Storage) CreateReceived(ctx context.Context, envelop *entity.Envelope) error {
	if envelop == nil {
		return errx.ErrParamIsNil("envelop")
	}
	if envelop.Topic == "" {
		return errx.ErrEmptyTopic
	}
	headers, err := envelop.HeadersBytes()
	if err != nil {
		return err
	}
	payload, err := envelop.PayloadBytes()
	if err != nil {
		return err
	}
	if len(payload) == 0 {
		return errx.ErrNilPayload
	}

	rec := &Received{
		Model: Model{
			ID:        envelop.ID,
			CreatedAt: time.Now(),
			Version:   envelop.Version,
			Topic:     envelop.Topic,
			Status:    enum.StatusPending,
			Headers:   string(headers),
			Payload:   string(payload),
		},
		Group: envelop.Group,
	}
	err = s.db.Create(rec).Error
	if err != nil {
		return err
	}
	envelop.ID = rec.ID
	return nil
}

func NewStorate(gapOpts *internal.Options) *Storage {
	opts := gapOpts.Gorm()
	s := &Storage{
		gapOpts: gapOpts,
		opts:    opts,
	}
	db := makeGormDB(opts)
	s.setDB(db)
	err := s.init()
	if err != nil {
		log.Fatalf("failed to init storage: %v", err)
	}
	var _ storage.Storage = s
	return s
}

func makeGormDB(c *gormopt.Options) *gorm.DB {
	if c.DB != nil {
		db := *c.DB
		return &db
	}
	var dial gorm.Dialector
	if c.MySQL != nil {
		dial = mysql.Open(c.MySQL.DSN)
	}

	db, err := gorm.Open(dial)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	return db
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
		conf.Logger = logger.Default.LogMode(s.opts.LogLevel)
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
	newS := Storage{
		gapOpts: s.gapOpts,
		opts:    s.opts,
	}
	newS.setDB(db)
	return &newS, nil
}

func (s *Storage) UpdateStatusPublished(ctx context.Context, id uint, status enum.Status) error {
	return s.db.WithContext(ctx).
		Model(&Published{}).
		Where("id = ?", id).
		Update("status", status).Error
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
			Where("version = ?", o.Version).
			Where("status IN ?", []enum.Status{enum.StatusPending, enum.StatusFailed}).
			Where("retries < ?", o.MaxRetries).
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
			Where("id IN ?", ids).
			Update("status", enum.StatusProcessing).Error
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
			Where("version = ?", o.Version).
			Where("`group` = ?", o.Group).
			Where("status IN ?", []enum.Status{enum.StatusPending, enum.StatusFailed}).
			Where("retries < ?", o.MaxRetries).
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
			Where("id IN ?", ids).
			Update("status", enum.StatusProcessing).Error
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
		Where("id = ?", id).
		Update("status", status).Error
}
