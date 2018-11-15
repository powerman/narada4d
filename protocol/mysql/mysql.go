package mysql

import (
	"database/sql"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/powerman/must"
	"github.com/powerman/narada4d/schemaver"
)

const (
	sqlCreateTable = `
CREATE TABLE Narada4D (
	 var VARCHAR(191) PRIMARY KEY
	,val VARCHAR(255) NOT NULL
)
SELECT "version" as var, "none" as val
`
	sqlInitialized   = `SELECT COUNT(*) FROM Narada4D`
	sqlSharedLock    = `LOCK TABLE Narada4D READ`
	sqlExclusiveLock = `LOCK TABLE Narada4D WRITE`
	sqlUnlock        = `UNLOCK TABLES`
	sqlGetVersion    = `SELECT val FROM Narada4D WHERE var='version'`
	sqlSetVersion    = `UPDATE Narada4D SET val=? WHERE var='version'`
)

type storage struct {
	db *sql.DB
	tx *sql.Tx
}

func init() {
	schemaver.RegisterProtocol("mysql", schemaver.Backend{
		Initialize: initialize,
		New:        new,
	})
}

func connect(loc *url.URL) (*storage, error) {
	err := validate(loc)
	if err != nil {
		return nil, err
	}

	cfg := mysql.NewConfig()
	cfg.User = loc.User.Username()
	cfg.Passwd, _ = loc.User.Password()
	cfg.Net = "tcp"
	cfg.Addr = loc.Host
	cfg.DBName = strings.TrimPrefix(loc.Path, "/")
	cfg.Collation = "utf8mb4_unicode_ci"
	cfg.ReadTimeout = 5 * time.Second
	cfg.WriteTimeout = 5 * time.Second
	cfg.ParseTime = true

	s := &storage{}
	s.db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	err = s.db.Ping()
	return s, err
}

func validate(loc *url.URL) error {
	switch {
	case loc.User == nil || loc.User.Username() == "":
		return errors.New("username absent, require mysql://username[:password]@host[:port]/database")
	case loc.Host == "":
		return errors.New("host absent, require mysql://username[:password]@host[:port]/database")
	case loc.Path == "" || loc.Path == "/":
		return errors.New("database absent, require mysql://username[:password]@host[:port]/database")
	case loc.RawQuery != "" || loc.Fragment != "":
		return errors.New("unexpected query params or fragment, require mysql://username[:password]@host[:port]/database")
	default:
		return nil
	}
}

func initialize(loc *url.URL) error {
	s, err := connect(loc)
	if err != nil {
		return err
	}
	defer s.db.Close()
	_, err = s.db.Exec(sqlCreateTable)
	return err
}

func new(loc *url.URL) (schemaver.Manage, error) {
	return connect(loc)
}

func (s *storage) initialized() bool {
	v, err := s.db.Query(sqlInitialized)
	if err != nil {
		return false
	}
	v.Close()
	return true
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
	var version string
	err := s.tx.QueryRow(sqlGetVersion).Scan(&version)
	must.PanicIf(err)
	return version
}

var reVersion = regexp.MustCompile(`\A(?:none|dirty|\d+(?:[.]\d+)*)\z`)

func (s *storage) Set(ver string) {
	if s.tx == nil {
		panic("not locked")
	}
	if reVersion.MatchString(ver) {
		_, err := s.tx.Exec(sqlSetVersion, ver)
		must.PanicIf(err)
	} else {
		panic("invalid version value, require 'none' or 'dirty' or one or more digits separated with single dots")
	}
}
