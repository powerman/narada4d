// Package goosepostgres registers schemaver.Backend implemented using goose.
package goosepostgres

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"strconv"

	"github.com/cenkalti/backoff/v4"
	"github.com/lib/pq"
	"github.com/powerman/goose"
	"github.com/powerman/must"
	_ "github.com/powerman/pqx" // Registers pqx driver.

	"github.com/powerman/narada4d/internal"
	"github.com/powerman/narada4d/schemaver"
)

const (
	sqlInitialized   = `SELECT COUNT(*) FROM goose_db_version`
	sqlSharedLock    = `LOCK TABLE goose_db_version IN SHARE MODE`
	sqlExclusiveLock = `LOCK TABLE goose_db_version IN SHARE UPDATE EXCLUSIVE MODE`
)

var (
	errAlreadyInitialized = errors.New("already initialized")
	errLocked             = errors.New("locked")
)

type storage struct {
	db    *sql.DB
	tx    *sql.Tx
	goose *goose.Instance
}

func init() { //nolint:gochecknoinits // Registration pattern.
	schemaver.RegisterProtocol("goose-postgres", schemaver.Backend{
		Initialize: initialize,
		New:        newInitializedStorage,
	})
}

func initialize(loc *url.URL) error {
	s, err := newStorage(loc)
	if err != nil {
		return err
	}
	defer s.Close() //nolint:errcheck // Defer.

	if s.initialized() {
		return errAlreadyInitialized
	}
	return s.init()
}

func newInitializedStorage(loc *url.URL) (schemaver.Manage, error) {
	s, err := newStorage(loc)
	if err != nil {
		return nil, err
	}
	if !s.initialized() {
		err := s.init()
		if err != nil {
			_ = s.Close()
			return nil, err
		}
	}
	return s, nil
}

func newStorage(loc *url.URL) (*storage, error) {
	loc.Scheme = "postgres"
	db, err := sql.Open("pqx", loc.String())
	if err != nil {
		return nil, err
	}
	err = db.PingContext(context.Background())
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	s := &storage{
		db:    db,
		goose: goose.NewInstance(),
	}
	err = s.goose.SetDialect("postgres")
	if err != nil {
		_ = s.Close()
		return nil, err
	}
	return s, nil
}

func (s *storage) initialized() bool {
	var count int
	_ = s.db.QueryRowContext(context.Background(), sqlInitialized).Scan(&count)
	return count > 0
}

func (s *storage) init() error {
	_, err := goose.EnsureDBVersion(s.db)
	return err
}

func (s *storage) SharedLock() {
	if s.tx != nil {
		panic("already locked")
	}
	op := func() (err error) {
		s.tx, err = s.db.BeginTx(context.Background(), nil)
		if err == nil {
			_, err = s.tx.ExecContext(context.Background(), "SET LOCAL statement_timeout = 0")
		}
		if err == nil {
			_, err = s.tx.ExecContext(context.Background(), "SET LOCAL lock_timeout = 0")
		}
		if err == nil {
			_, err = s.tx.ExecContext(context.Background(), "SET LOCAL idle_in_transaction_session_timeout = 0")
		}
		if err == nil {
			_, err = s.tx.ExecContext(context.Background(), sqlSharedLock)
		}
		if err != nil {
			if s.tx != nil {
				_ = s.tx.Rollback()
				s.tx = nil
			}
			if errors.As(err, new(*pq.Error)) { // Retry on network errors.
				err = backoff.Permanent(err)
			}
		}
		return err
	}
	must.PanicIf(backoff.Retry(op, internal.NewBackOff()))
}

func (s *storage) ExclusiveLock() {
	if s.tx != nil {
		panic("already locked")
	}
	op := func() (err error) {
		s.tx, err = s.db.BeginTx(context.Background(), nil)
		if err == nil {
			_, err = s.tx.ExecContext(context.Background(), "SET LOCAL statement_timeout = 0")
		}
		if err == nil {
			_, err = s.tx.ExecContext(context.Background(), "SET LOCAL lock_timeout = 0")
		}
		if err == nil {
			_, err = s.tx.ExecContext(context.Background(), "SET LOCAL idle_in_transaction_session_timeout = 0")
		}
		if err == nil {
			_, err = s.tx.ExecContext(context.Background(), sqlExclusiveLock)
		}
		if err != nil {
			if s.tx != nil {
				_ = s.tx.Rollback()
				s.tx = nil
			}
			if errors.As(err, new(*pq.Error)) { // Retry on network errors.
				err = backoff.Permanent(err)
			}
		}
		return err
	}
	must.PanicIf(backoff.Retry(op, internal.NewBackOff()))
}

func (s *storage) Unlock() {
	if s.tx == nil {
		panic("not locked")
	}
	err := s.tx.Commit()
	s.tx = nil
	if err != nil && !errors.As(err, new(*pq.Error)) { // Ignore network errors.
		err = nil
	}
	must.PanicIf(err)
}

func (s *storage) Get() string {
	v, err := goose.EnsureDBVersion(s.db)
	must.PanicIf(err)
	if v == 0 {
		return "none"
	}
	return strconv.Itoa(int(v))
}

func (*storage) Set(string) {
	panic("not supported")
}

func (s *storage) Close() error {
	if s.tx != nil {
		return errLocked
	}
	return s.db.Close()
}
