// +build integration

package mysql

import (
	"context"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	proxypkg "github.com/docker/go-connections/proxy"
	"github.com/go-sql-driver/mysql"
	"github.com/powerman/check"
	"github.com/powerman/gotest/testinit"
	"github.com/powerman/mysqlx"

	"github.com/powerman/narada4d/internal"
)

const (
	testDBSuffix = "github.com/powerman/narada4d/protocol/mysql"
	sqlDropTable = "DROP TABLE Narada4D"
)

var (
	loc   *url.URL
	proxy *proxypkg.TCPProxy
)

func init() { testinit.Setup(2, setupIntegration) }

func setupIntegration() {
	logger := log.New(os.Stderr, "", log.LstdFlags)
	var err error

	loc, err = url.Parse(os.Getenv("NARADA4D_TEST_MYSQL"))
	if err != nil {
		testinit.Fatal("failed to parse $NARADA4D_TEST_MYSQL as URL: ", err)
	}

	dbCfg, err := mysql.ParseDSN(dsn(loc))
	if err != nil {
		testinit.Fatal("failed to parse $NARADA4D_TEST_MYSQL as DSN: ", err)
	}
	dbCfg.Timeout = 3 * testSecond

	ctx, cancel := context.WithTimeout(ctx, 7*testSecond)
	defer cancel()
	proxy, err = internal.NewTCPProxy(ctx, "127.0.0.1:0", dbCfg.Addr)
	if err != nil {
		testinit.Fatal("failed to NewTCPProxy: ", err)
	}
	testinit.Teardown(func() { proxy.Close() })
	dbCfg.Addr = proxy.FrontendAddr().String()

	dbCfg, cleanup, err := mysqlx.EnsureTempDB(logger, testDBSuffix, dbCfg)
	if err != nil {
		testinit.Fatal(err)
	}
	testinit.Teardown(cleanup)

	loc.Host = dbCfg.Addr
	loc.Path = "/" + dbCfg.DBName
}

func dsn(loc *url.URL) string {
	dsn := &url.URL{}
	*dsn = *loc
	dsn.Host = "tcp(" + dsn.Host + ")"
	return strings.TrimPrefix(dsn.String(), "mysql://")
}

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
