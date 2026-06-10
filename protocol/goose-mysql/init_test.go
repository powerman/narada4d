package goosemysql //nolint:testpackage // TestMain requires same package.

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
	if os.Getenv("NARADA4D_TEST_MYSQL") == "" {
		fmt.Println("$NARADA4D_TEST_MYSQL must be set for goose-mysql integration tests (skipping)")
		return
	}
	testinit.Main(m)
}

var (
	ctx            = context.Background()
	testTimeFactor = getenv.Float("GO_TEST_TIME_FACTOR", 1.0)
	testSecond     = time.Duration(float64(time.Second) * testTimeFactor)
)
