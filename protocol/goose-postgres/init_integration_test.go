// +build integration

package goosepostgres

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/powerman/check"
	"github.com/powerman/gotest/testinit"
	"github.com/powerman/pqx"
)

const (
	testDBSuffix = "github.com/powerman/narada4d/protocol/goose_postgres"
	sqlDropTable = "DROP TABLE goose_db_version"
)

var loc *url.URL

func init() { testinit.Setup(2, setupIntegration) }

func setupIntegration() {
	logger := log.New(os.Stderr, "", log.LstdFlags)
	_, cleanup, err := pqx.EnsureTempDB(logger, testDBSuffix, pqx.Config{
		ConnectTimeout: 3 * testSecond,
		DBName:         os.Getenv("PGDATABASE"),
	})
	if err != nil {
		if err, ok := err.(*pq.Error); !ok || err.Code.Class().Name() == "invalid_authorization_specification" {
			logger.Print("set environment variables to allow connection to postgresql:\nhttps://www.postgresql.org/docs/current/libpq-envars.html")
		}
		testinit.Fatal(err)
	}
	testinit.Teardown(cleanup)

	loc, err = url.Parse(fmt.Sprintf("goose-postgres://%s:%s@%s:%s/%s?sslmode=%s",
		env2path("PGUSER"), env2path("PGPASSWORD"),
		env2path("PGHOST"), env2path("PGPORT"),
		env2path("PGDATABASE")+url.PathEscape("_"+testDBSuffix),
		env2query("PGSSLMODE")))
	if err != nil {
		testinit.Fatal(err)
	}
}

func env2path(env string) string  { return url.PathEscape(os.Getenv(env)) }
func env2query(env string) string { return url.QueryEscape(os.Getenv(env)) }

func dropTable(t *check.C) {
	t.Helper()
	v, err := newStorage(loc)
	t.Nil(err)
	s := v.(*storage)
	_, err = s.db.Exec(sqlDropTable)
	t.Nil(err)
	t.Nil(v.Close())
}

func testLock(name string, loc *url.URL, unlockc chan struct{}, statusc chan string) {
	v, err := newStorage(loc)
	if err != nil {
		panic(err)
	}

	cancel := make(chan struct{}, 1)
	go func() {
		select {
		case <-cancel:
		case <-time.After(testSecond / 10):
			statusc <- "block " + name
		}
	}()

	switch {
	case strings.HasPrefix(name, "EX"):
		v.ExclusiveLock()
	case strings.HasPrefix(name, "SH"):
		v.SharedLock()
	default:
		panic("name must begins with EX or SH")
	}
	cancel <- struct{}{}
	statusc <- "acquired " + name

	<-unlockc
	v.Unlock()
	_ = v.(*storage).db.Close()
}
