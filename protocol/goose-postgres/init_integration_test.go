// +build integration

package goosepostgres

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	proxypkg "github.com/docker/go-connections/proxy"
	"github.com/lib/pq"
	"github.com/powerman/check"
	"github.com/powerman/gotest/testinit"
	"github.com/powerman/pqx"

	"github.com/powerman/narada4d/internal"
)

const (
	testDBSuffix = "github.com/powerman/narada4d/protocol/goose_postgres"
	sqlDropTable = "DROP TABLE goose_db_version"
)

var (
	loc   *url.URL
	proxy *proxypkg.TCPProxy
)

func init() { testinit.Setup(2, setupIntegration) }

func setupIntegration() {
	logger := log.New(os.Stderr, "", log.LstdFlags)
	var err error

	loc, err = url.Parse(fmt.Sprintf("goose-postgres://%s:%s@%s:%s/%s?sslmode=%s",
		env2path("PGUSER"), env2path("PGPASSWORD"),
		env2path("PGHOST"), env2path("PGPORT"),
		env2path("PGDATABASE")+url.PathEscape("_"+testDBSuffix),
		env2query("PGSSLMODE")))
	if err != nil {
		testinit.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(ctx, 7*testSecond)
	defer cancel()
	proxy, err = internal.NewTCPProxy(ctx, "127.0.0.1:0", loc.Host)
	if err != nil {
		testinit.Fatal("failed to NewTCPProxy: ", err)
	}
	testinit.Teardown(func() { proxy.Close() })
	loc.Host = proxy.FrontendAddr().String()

	dbCfg := pqx.Config{
		ConnectTimeout: 3 * testSecond,
		DBName:         os.Getenv("PGDATABASE"),
		Host:           "127.0.0.1",
		Port:           proxy.FrontendAddr().(*net.TCPAddr).Port,
	}

	_, cleanup, err := pqx.EnsureTempDB(logger, testDBSuffix, dbCfg)
	if err != nil {
		if err, ok := err.(*pq.Error); !ok || err.Code.Class().Name() == "invalid_authorization_specification" {
			logger.Print("set environment variables to allow connection to postgresql:\nhttps://www.postgresql.org/docs/current/libpq-envars.html")
		}
		testinit.Fatal(err)
	}
	testinit.Teardown(cleanup)
}

func env2path(env string) string  { return url.PathEscape(os.Getenv(env)) }
func env2query(env string) string { return url.QueryEscape(os.Getenv(env)) }

func dropTable(t *check.C) {
	t.Helper()
	s, err := newStorage(loc)
	t.Nil(err)
	_, err = s.db.Exec(sqlDropTable)
	t.Nil(err)
	t.Nil(s.Close())
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
	_ = v.Close()
}
