package mysql

import (
	"database/sql"
	"errors"
	"net/url"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/powerman/narada4d/schemaver"
)

const (
	createTable = `
			CREATE TABLE Narada4D(
			var VARCHAR(255) PRIMARY KEY,
			val VARCHAR(255) NOT NULL
			)
			SELECT "version" as var, "none" as val
			`

	isInitialized = `SELECT COUNT(*) FROM Narada4D`
	sharedLock    = `LOCK TABLE Narada4D READ`
	exclusiveLock = `LOCK TABLE Narada4D WRITE`
	unlock        = `UNLOCK TABLES`
	getVersion    = `SELECT val FROM Narada4D WHERE var='version'`
	setVersion    = `UPDATE Narada4D SET val=? WHERE var='version'`
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
	s := &storage{}
	err := parse(loc)
	if err != nil {
		return nil, err
	}
	pass, _ := loc.User.Password()

	cfg := mysql.NewConfig()
	cfg.User = loc.User.Username()
	cfg.Passwd = pass
	cfg.Net = "tcp"
	cfg.Addr = loc.Host
	cfg.DBName = loc.Path
	cfg.Collation = "utf8mb4_general_ci"
	cfg.ReadTimeout = 5 * time.Second
	cfg.WriteTimeout = 5 * time.Second
	cfg.ParseTime = true

	s.db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	return s, err
}

func parse(loc *url.URL) error {
	if loc.User.Username == nil || loc.User.Password == nil || loc.Host == "" || loc.Path == "" || loc.RawQuery != "" || loc.Fragment != "" {
		return errors.New("location muct contain only User, Host, Path")
	}
	return nil
}

func initialize(loc *url.URL) error {
	s, err := connect(loc)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(createTable)
	if err != nil {
		return err
	}
	defer func() {
		err := s.db.Close()
		if err != nil {
			panic(err)
		}
	}()
	return err
}

func new(loc *url.URL) (v schemaver.Manage, err error) {
	s, err := connect(loc)
	if err != nil {
		panic(err)
	}
	return s, nil
}

func (s *storage) initialized() bool {
	v, err := s.db.Query(isInitialized)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := v.Close()
		if err != nil {
			panic(err)
		}
	}()
	return true
}

func (s *storage) SharedLock() {
	_, err := s.db.Exec(sharedLock)
	if err != nil {
		panic(err)
	}
}

func (s *storage) ExclusiveLock() {
	_, err := s.db.Exec(exclusiveLock)
	if err != nil {
		panic(err)
	}
}

func (s *storage) Unlock() {
	_, err := s.db.Exec(unlock)
	if err != nil {
		panic(err)
	}
}

func (s *storage) Get() string {
	var version string
	_ = s.db.QueryRow(getVersion).Scan(&version)
	return version
}

func (s *storage) Set(ver string) {
	_, err := s.db.Exec(setVersion, ver)
	if err != nil {
		panic(err)
	}
}
