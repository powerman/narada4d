// wait-mysql waits for MySQL to become available
// by periodically trying to connect and ping it
// using NARADA4D_TEST_MYSQL environment variable as connection URL.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const defaultTimeout = 60 * time.Second

func mysqlDSN(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	dsn := &url.URL{}
	*dsn = *u
	dsn.Host = "tcp(" + dsn.Host + ")"
	return strings.TrimPrefix(dsn.String(), "mysql://")
}

func run() error {
	dsnURL := os.Getenv("NARADA4D_TEST_MYSQL")
	if dsnURL == "" {
		return nil
	}

	dsnStr := mysqlDSN(dsnURL)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	for {
		db, err := sql.Open("mysql", dsnStr)
		if err == nil {
			err = db.PingContext(ctx)
			_ = db.Close()
		}
		if err == nil {
			fmt.Println("MySQL is ready")
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("MySQL not ready within %v: %w", defaultTimeout, err)
		case <-time.After(1 * time.Second):
		}
	}
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
