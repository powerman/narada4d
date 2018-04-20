package mysql

import (
	"database/sql"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/powerman/narada4d/schemaver"
)

const (
	sqlCreateTable = `
			CREATE TABLE Narada4D(
			var VARCHAR(255) PRIMARY KEY,
			val VARCHAR(255) NOT NULL
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
	cfg.Collation = "utf8mb4_general_ci"
	cfg.ReadTimeout = 5 * time.Second
	cfg.WriteTimeout = 5 * time.Second
	cfg.ParseTime = true

	s := &storage{}
	s.db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	return s, err
}

func validate(loc *url.URL) error {
	if loc.User == nil || loc.User.Username() == "" {
		return errors.New("Username absent, require mysql://username[:password]@host[:port]/database")
	} else if loc.Host == "" {
		return errors.New("incorrect Host, require mysql://username[:password]@host[:port]/database")
	} else if loc.Path == "" || loc.Path == "/" {
		return errors.New("require database, mysql://username[:password]@host[:port]/database")
	} else if loc.RawQuery != "" || loc.Fragment != "" {
		return errors.New("unexpected query params or fragment, require mysql://username[:password]@host[:port]/database")
	}
	return nil
}

func initialize(loc *url.URL) error {
	s, err := connect(loc)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(sqlCreateTable)
	if err != nil {
		return err
	}
	return s.db.Close()
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
	_, err := s.db.Exec(sqlSharedLock)
	if err != nil {
		panic(err)
	}
}

func (s *storage) ExclusiveLock() {
	_, err := s.db.Exec(sqlExclusiveLock)
	if err != nil {
		panic(err)
	}
}

func (s *storage) Unlock() {
	_, err := s.db.Exec(sqlUnlock)
	if err != nil {
		panic(err)
	}
}

func (s *storage) Get() string {
	var version string
	err := s.db.QueryRow(sqlGetVersion).Scan(&version)
	if err != nil {
		panic(err)
	}
	return version
}

func (s *storage) Set(ver string) {
	_, err := s.db.Exec(sqlSetVersion, ver)
	if err != nil {
		panic(err)
	}
}
