// Package goosemysql registers schemaver.Backend implemented using goose.
package goosemysql

import (
	"database/sql"
	"errors"
	"net/url"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql" // Driver.
	"github.com/powerman/goose"
	"github.com/powerman/must"
	"github.com/powerman/narada4d/schemaver"
)

const (
	sqlInitialized   = `SELECT COUNT(*) FROM goose_db_version`
	sqlSharedLock    = `LOCK TABLE goose_db_version READ`
	sqlExclusiveLock = `LOCK TABLE goose_db_version WRITE`
	sqlUnlock        = `UNLOCK TABLES`
)

type storage struct {
	db    *sql.DB
	tx    *sql.Tx
	goose *goose.Instance
}

func init() {
	schemaver.RegisterProtocol("goose-mysql", schemaver.Backend{
		Initialize: initialize,
		New:        newStorage,
	})
}

func initialized(db *sql.DB) bool {
	var count int
	_ = db.QueryRow(sqlInitialized).Scan(&count)
	return count > 0
}

func initialize(loc *url.URL) error {
	manage, err := newStorage(loc)
	if err != nil {
		return err
	}
	s := manage.(*storage)
	defer s.db.Close() //nolint:errcheck // Defer.

	if initialized(s.db) {
		return errors.New("already initialized")
	}

	_, err = s.goose.EnsureDBVersion(s.db)
	return err
}

func dsn(loc *url.URL) string {
	dsn := &url.URL{}
	*dsn = *loc
	dsn.Host = "tcp(" + dsn.Host + ")"
	return strings.TrimPrefix(dsn.String(), "goose-mysql://")
}

func newStorage(loc *url.URL) (schemaver.Manage, error) {
	db, err := sql.Open("mysql", dsn(loc))
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	s := &storage{
		db:    db,
		goose: goose.NewInstance(),
	}
	err = s.goose.SetDialect("mysql")
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *storage) SharedLock() {
	if s.tx != nil {
		panic("already locked")
	}
	var err error
	s.tx, err = s.db.Begin()
	must.PanicIf(err)
	_, err = s.tx.Exec(sqlSharedLock)
	must.PanicIf(err)
}

func (s *storage) ExclusiveLock() {
	if s.tx != nil {
		panic("already locked")
	}
	var err error
	s.tx, err = s.db.Begin()
	must.PanicIf(err)
	_, err = s.tx.Exec(sqlExclusiveLock)
	must.PanicIf(err)
}

func (s *storage) Unlock() {
	if s.tx == nil {
		panic("not locked")
	}
	_, err := s.tx.Exec(sqlUnlock)
	must.PanicIf(err)
	must.PanicIf(s.tx.Commit())
	s.tx = nil
}

func (s *storage) Get() string {
	if s.tx == nil {
		panic("not locked")
	}
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
		return errors.New("locked")
	}
	return s.db.Close()
}
