package goosepostgres //nolint:testpackage // TestMain requires same package.

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/powerman/getenv"
	"github.com/powerman/gotest/testinit"
)

func TestMain(m *testing.M) {
	if os.Getenv("PGUSER") == "" {
		fmt.Println("$PGUSER, $PGPASSWORD, $PGHOST, $PGPORT, $PGDATABASE, $PGSSLMODE must be set for PostgreSQL integration tests (skipping)")
		return
	}
	testinit.Main(m)
}

var (
	ctx            = context.Background()
	testTimeFactor = getenv.Float("GO_TEST_TIME_FACTOR", 1.0)
	testSecond     = time.Duration(float64(time.Second) * testTimeFactor)
)
