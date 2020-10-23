// Package goosemysql registers schemaver.Backend implemented using goose.
package goosemysql

import (
	"database/sql"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-sql-driver/mysql"
	"github.com/powerman/goose"
	"github.com/powerman/must"

	"github.com/powerman/narada4d/internal"
	"github.com/powerman/narada4d/schemaver"
)

const (
	sqlCreateTable = `
CREATE TABLE Narada4D (
	 var VARCHAR(191) PRIMARY KEY
	,val VARCHAR(255) NOT NULL
)
SELECT "version_from" as var, "goose" as val
`
	sqlInitialized   = `SELECT COUNT(*) FROM Narada4D`
	sqlSharedLock    = `LOCK TABLES Narada4D READ`
	sqlExclusiveLock = `LOCK TABLES Narada4D WRITE`
	sqlUnlock        = `UNLOCK TABLES`
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

func init() {
	schemaver.RegisterProtocol("goose-mysql", schemaver.Backend{
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
		if err := s.init(); err != nil {
			_ = s.Close()
			return nil, err
		}
	}
	return s, nil
}

func dsn(loc *url.URL) string {
	dsn := &url.URL{}
	*dsn = *loc
	dsn.Host = "tcp(" + dsn.Host + ")"
	return strings.TrimPrefix(dsn.String(), "goose-mysql://")
}

func newStorage(loc *url.URL) (*storage, error) {
	db, err := sql.Open("mysql", dsn(loc))
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	s := &storage{
		db:    db,
		goose: goose.NewInstance(),
	}
	err = s.goose.SetDialect("mysql")
	if err != nil {
		_ = s.Close()
		return nil, err
	}
	return s, nil
}

func (s *storage) initialized() bool {
	var count int
	_ = s.db.QueryRow(sqlInitialized).Scan(&count)
	return count > 0
}

func (s *storage) init() error {
	_, err := s.goose.EnsureDBVersion(s.db)
	if err == nil {
		_, err = s.db.Exec(sqlCreateTable)
	}
	return err
}

func (s *storage) SharedLock() {
	if s.tx != nil {
		panic("already locked")
	}
	op := func() (err error) {
		s.tx, err = s.db.Begin()
		if err == nil {
			_, err = s.tx.Exec(sqlSharedLock)
		}
		if errors.As(err, new(*mysql.MySQLError)) { // Retry on network errors.
			err = backoff.Permanent(err)
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
		s.tx, err = s.db.Begin()
		if err == nil {
			_, err = s.tx.Exec(sqlExclusiveLock)
		}
		if errors.As(err, new(*mysql.MySQLError)) { // Retry on network errors.
			err = backoff.Permanent(err)
		}
		return err
	}
	must.PanicIf(backoff.Retry(op, internal.NewBackOff()))
}

func (s *storage) Unlock() {
	if s.tx == nil {
		panic("not locked")
	}
	_, err := s.tx.Exec(sqlUnlock)
	if err == nil {
		err = s.tx.Commit()
	}
	s.tx = nil
	if err != nil && !errors.As(err, new(*mysql.MySQLError)) { // Ignore network errors.
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

func (s *storage) Set(string) {
	panic("not supported")
}

func (s *storage) Close() error {
	if s.tx != nil {
		return errLocked
	}
	return s.db.Close()
}
