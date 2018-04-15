package mysql

import (
	"database/sql"
	"log"
	"net/url"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/powerman/narada4d/schemaver"
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

func initialize(loc *url.URL) error {
	s := &storage{}
	var err error
	s.db, err = sql.Open("mysql", loc.String())
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`CREATE TABLE Narada4D(var VARCHAR (255) PRIMARY KEY, val VARCHAR(255) NOT NULL) SELECT "version" as var, "none" as val`)

	if err != nil {
		return (err)
	}
	s.db.Close()
	return err
}

func new(loc *url.URL) (v schemaver.Manage, err error) {
	s := &storage{}
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
	return s, nil
}

func initialized() bool {
	s := &storage{}
	_, err := s.db.Query(`SELECT COUNT(*) FROM Narada4D`)
	if err != nil {
		return false
	}
	return true
}

func (s *storage) SharedLock() {
	_, err := s.db.Exec(`LOCK TABLE Narada4D READ`)
	if err != nil {
		panic(err)
	}
}

func (s *storage) ExclusiveLock() {
	_, err := s.db.Exec(`LOCK TABLE Narada4D WRITE`)
	if err != nil {
		panic(err)
	}
}

func (s *storage) Unlock() {
	_, err := s.db.Exec(`UNLOCK TABLES`)
	if err != nil {
		panic(err)
	}
}

func (s *storage) Get() string {
	v, err := s.db.Query(`SELECT val FROM Narada4D WHERE var='version'`)
	if err != nil {
		log.Fatal(err)
	}
	defer v.Close()
	var version string
	for v.Next() {
		if err := v.Scan(&version); err != nil {
			log.Fatal(err)
		}
	}

	return version
}

func (s *storage) Set(ver string) {
	_, err := s.db.Exec(`UPDATE Narada4D SET val=? WHERE var='version'`, ver)
	if err != nil {
		panic(err)
	}
}
