package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/lopolopen/gap/internal"
	"github.com/lopolopen/gap/internal/entity"
	"github.com/lopolopen/gap/internal/enum"
	"github.com/lopolopen/gap/internal/errx"
	"github.com/lopolopen/gap/internal/storage"
	"github.com/lopolopen/gap/internal/tx"
	mysqlopt "github.com/lopolopen/gap/options/mysql"
)

type Storage struct {
	gapOpts *internal.Options
	opts    *mysqlopt.Options
	db      *sql.DB
	tx      *sql.Tx
}

func (s *Storage) execer() Execer {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

// Bind implements [storage.Storage].
func (s *Storage) Bind(txer tx.Txer) (storage.Storage, error) {
	tx, ok := txer.Tx().(*sql.Tx)
	if !ok {
		return nil, errx.ErrInvalidSQLTx
	}
	newS := Storage{
		gapOpts: s.gapOpts,
		opts:    s.opts,
		db:      s.db,
		tx:      tx,
	}
	return &newS, nil
}

// ClaimPublishedBatch implements [storage.Storage].
func (s *Storage) ClaimPublishedBatch(ctx context.Context, batchSize int) (es []*entity.Envelope, err error) {
	o := s.gapOpts
	var pubs []*Published
	const script1 = "SELECT `id`,`created_at`,`version`,`topic`,`status`,`headers`,`payload`,`retries`,`expired_at` " +
		"FROM `%s_published` WHERE `version` = ? AND `status` IN (?, ?) AND `retries` < ? LIMIT ? FOR UPDATE SKIP LOCKED"
	const script2 = "UPDATE `%s_published` SET `status` = ? WHERE `id` IN (%s)"

	var tx *sql.Tx
	tx, err = s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	rows, err := tx.QueryContext(ctx, fmt.Sprintf(script1, s.opts.Schema),
		o.Version,
		enum.StatusPending,
		enum.StatusFailed,
		o.MaxRetries,
		batchSize,
	)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var pub Published
		err = rows.Scan(
			&pub.ID,
			&pub.CreatedAt,
			&pub.Version,
			&pub.Topic,
			&pub.Status,
			&pub.Headers,
			&pub.Payload,
			&pub.Retries,
			&pub.ExpiredAt,
		)
		if err != nil {
			return nil, err
		}
		pubs = append(pubs, &pub)
	}
	c := len(pubs)
	if c == 0 {
		return nil, nil
	}
	placeholders := make([]string, c)
	params := make([]any, c+1)
	params[0] = enum.StatusProcessing
	for i, p := range pubs {
		placeholders[i] = "?"
		params[i+1] = p.ID
		es = append(es, p.ToEntity())
	}
	_, err = tx.ExecContext(ctx, fmt.Sprintf(script2, s.opts.Schema, strings.Join(placeholders, ", ")), params...)
	if err != nil {
		return nil, err
	}
	return es, nil
}

// ClaimReceivedBatch implements [storage.Storage].
func (s *Storage) ClaimReceivedBatch(ctx context.Context, batchSize int) (es []*entity.Envelope, err error) {
	o := s.gapOpts
	var recs []*Received
	const script1 = "SELECT `id`,`created_at`,`version`,`topic`,`status`,`headers`,`payload`,`retries`,`expired_at`,`group` FROM `%s_received` " +
		"WHERE `version` = ? " +
		"AND `group` = ? " +
		"AND `status` IN (?, ?) " +
		"AND `retries` < ? " +
		"LIMIT ? FOR UPDATE SKIP LOCKED"
	const script2 = "UPDATE `%s_received` SET `status` = ? WHERE `id` IN (%s)"

	var tx *sql.Tx
	tx, err = s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	rows, err := tx.QueryContext(ctx, fmt.Sprintf(script1, s.opts.Schema),
		o.Version,
		o.Group,
		enum.StatusPending,
		enum.StatusFailed,
		o.MaxRetries,
		batchSize,
	)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var rec Received
		err = rows.Scan(
			&rec.ID,
			&rec.CreatedAt,
			&rec.Version,
			&rec.Topic,
			&rec.Status,
			&rec.Headers,
			&rec.Payload,
			&rec.Retries,
			&rec.ExpiredAt,
			&rec.Group,
		)
		if err != nil {
			return nil, err
		}
		recs = append(recs, &rec)
	}
	c := len(recs)
	if c == 0 {
		return nil, nil
	}
	placeholders := make([]string, c)
	params := make([]any, c+1)
	params[0] = enum.StatusProcessing
	for i, r := range recs {
		placeholders[i] = "?"
		params[i+1] = r.ID
		es = append(es, r.ToEntity())
	}
	_, err = tx.ExecContext(ctx, fmt.Sprintf(script2, s.opts.Schema, strings.Join(placeholders, ", ")), params...)
	if err != nil {
		return nil, err
	}
	return es, nil
}

// CreatePublished implements [storage.Storage].
func (s *Storage) CreatePublished(ctx context.Context, envelop *entity.Envelope) error {
	if envelop == nil {
		return errx.ErrParamIsNil("envelop")
	}

	if s.tx == nil {
		slog.Warn("publishing does not work in transaction")
	}

	pub := new(Published).FromEntity(envelop)
	const script = "INSERT INTO `%s_published` (" +
		"`id`,`created_at`,`version`,`topic`,`status`,`headers`,`payload`,`retries`,`expired_at`" +
		") VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)"
	xer := s.execer()
	_, err := xer.ExecContext(ctx, fmt.Sprintf(script, s.opts.Schema),
		pub.ID,
		pub.CreatedAt,
		pub.Version,
		pub.Topic,
		pub.Status,
		string(pub.Headers),
		string(pub.Payload),
		pub.Retries,
		pub.ExpiredAt,
	)
	return err
}

// CreateReceived implements [storage.Storage].
func (s *Storage) CreateReceived(ctx context.Context, envelop *entity.Envelope) error {
	if envelop == nil {
		return errx.ErrParamIsNil("envelop")
	}

	rec := new(Received).FromEntity(envelop)
	const script = "INSERT INTO `%s_received` (" +
		"`id`,`created_at`,`version`,`topic`,`status`,`headers`,`payload`,`retries`,`expired_at`,`group`" +
		") VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := s.db.ExecContext(ctx, fmt.Sprintf(script, s.opts.Schema),
		rec.ID,
		rec.CreatedAt,
		rec.Version,
		rec.Topic,
		rec.Status,
		string(rec.Headers),
		string(rec.Payload),
		rec.Retries,
		rec.ExpiredAt,
		rec.Group,
	)
	return err
}

// UpdateStatusPublished implements [storage.Storage].
func (s *Storage) UpdateStatusPublished(ctx context.Context, id uint, status enum.Status) error {
	const script = "UPDATE `%s_published` SET `status` = ? WHERE `id` = ?"
	_, err := s.db.ExecContext(ctx, fmt.Sprintf(script, s.opts.Schema), status, id)
	if err != nil {
		return err
	}
	return nil
}

// UpdateStatusReceived implements [storage.Storage].
func (s *Storage) UpdateStatusReceived(ctx context.Context, id uint, status enum.Status) error {
	const script = "UPDATE `%s_received` SET `status` = ? WHERE `id` = ?"
	_, err := s.db.ExecContext(ctx, fmt.Sprintf(script, s.opts.Schema), status, id)
	if err != nil {
		return err
	}
	return nil
}

func NewStorage(gapOpts *internal.Options) *Storage {
	opts := gapOpts.MySQL()
	db, err := sql.Open("mysql", opts.DSN)
	if err != nil {
		slog.Error("failed to connect mysql database", slog.Any("err", err))
		panic(err)
	}
	s := &Storage{
		gapOpts: gapOpts,
		opts:    opts,
		db:      db,
	}
	err = s.init()
	if err != nil {
		slog.Error("failed to init storage", slog.Any("err", err))
		panic(err)
	}
	var _ storage.Storage = s
	return s
}

func (s *Storage) init() error {
	const script1 = "CREATE TABLE IF NOT EXISTS `%s_published` (" +
		"`id` bigint(20) unsigned NOT NULL," +
		"`created_at` datetime(3) NOT NULL," +
		"`version` varchar(16) NOT NULL," +
		"`topic` varchar(256) NOT NULL," +
		"`group` varchar(128) DEFAULT NULL," +
		"`status` enum('Pending','Processing','Succeeded','Failed') NOT NULL," +
		"`headers` text DEFAULT NULL," +
		"`payload` longtext DEFAULT NULL," +
		"`retries` bigint(20) DEFAULT 0," +
		"`expired_at` datetime(3) DEFAULT NULL," +
		"PRIMARY KEY (`id`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4"
	_, err := s.db.Exec(fmt.Sprintf(script1, s.opts.Schema))
	if err != nil {
		return err
	}

	const script2 = "CREATE TABLE IF NOT EXISTS `%s_received` (" +
		"`id` bigint(20) unsigned NOT NULL," +
		"`created_at` datetime(3) NOT NULL," +
		"`version` varchar(16) NOT NULL," +
		"`topic` varchar(256) NOT NULL," +
		"`group` varchar(128) DEFAULT NULL," +
		"`status` enum('Pending','Processing','Succeeded','Failed') NOT NULL," +
		"`headers` text DEFAULT NULL," +
		"`payload` longtext DEFAULT NULL," +
		"`retries` bigint(20) DEFAULT 0," +
		"`expired_at` datetime(3) DEFAULT NULL," +
		"PRIMARY KEY (`id`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4"
	_, err = s.db.Exec(fmt.Sprintf(script2, s.opts.Schema))
	if err != nil {
		return err
	}

	return nil
}
