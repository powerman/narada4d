package file

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/powerman/check"
)

// - initialize & new
//   - file://user@/
//   - file://localhost/
//   - file:///?a=1
//   - file:///#a
//   - file://
func TestBadLocation(t *testing.T) {

}

// - initialize
//   - file:///path/to/read-only/dir/
//   - file:///path/to/dir/with/subdir/.lock/
//   - file:///path/to/dir/with/subdir/.lock.queue/
//   - file:///path/to/dir/with/subdir/.version/
//   - file:///path/to/empty/dir/ (success)
//   - repeat initialize()
func TestInitialize(t *testing.T) {

}

// - new
//   - file:///path/to/empty/dir/
//   - after initialize() (success)
func TestNew(t *testing.T) {

}

func testLock(name string, loc *url.URL, unlockc chan struct{}, statusc chan string) {
	v, err := new(loc)
	if err != nil {
		panic(err)
	}

	cancel := make(chan struct{}, 1)
	go func() {
		select {
		case <-cancel:
		case <-time.After(100 * time.Millisecond):
			statusc <- "blocked " + name
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
}

// - EX1, UN1, EX2, UN2
func TestExSequence(t *testing.T) {

}

// - EX1, EX2 (block), UN1, (unblock EX2), UN2
func TestExParallel(tt *testing.T) {
	t := check.T(tt)
	// TODO tempdir, loc, initialize(loc)
	statusc := make(chan string)
	un1 := make(chan struct{})
	un2 := make(chan struct{})
	go testLock("EX1", nil, un1, statusc)
	t.Equal(<-statusc, "acquired EX1")
	go testLock("EX2", nil, un2, statusc)
	t.Equal(<-statusc, "blocked EX2")
	un1 <- struct{}{}
	t.Equal(<-statusc, "acquired EX2")
	un2 <- struct{}{}
}

// - EX1, SH2 (block), UN1, (unblock SH2), UN2
func TestExShParallel(t *testing.T) {

}

// - SH1, SH2, UN1, UN2
func TestShParallel(t *testing.T) {

}

// - SH1, EX2 (block), SH3 (block), UN1, (unblock EX2), UN2, (unblock SH3), UN3
func TestExPriority(t *testing.T) {

}

// - Get = "none", Get = "none"
func TestGetNone(t *testing.T) {

}

// - Set("dirty"), Get = "dirty"
// - Set(""), Get = ""
// - Set("0"), Get = "0"
// - Set("1.2.3"), Get = "1.2.3"
func TestSet(t *testing.T) {

}
