package file

import (
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/powerman/check"
)

func TestBadLocation(tt *testing.T) {
	t := check.T(tt)

	cases := []struct {
		path    string
		wanterr error
	}{
		{"file://user@/", errors.New("location must contain only path")},
		{"file://localhost/", errors.New("location must contain only path")},
		{"file:///?a=1", errors.New("location must contain only path")},
		{"file:///#a", errors.New("location must contain only path")},
	}

	for _, v := range cases {
		loc, err := url.Parse(v.path)
		t.Nil(err)
		t.Err(initialize(loc), v.wanterr, v.path)
		_, err = new(loc)
		t.Err(err, v.wanterr, v.path)
	}
}

func TestInitialize(tt *testing.T) {
	t := check.T(tt)

	// - file:///path/to/read-only/dir/
	tempdir, err := ioutil.TempDir("", "gotest")
	t.Nil(err)
	defer func() { t.Nil(os.Remove(tempdir)) }()
	t.Nil(os.Chmod(tempdir, 0555))
	loc, err := url.Parse(tempdir)
	t.Nil(err)

	t.Err(initialize(loc), syscall.EACCES)

	// - file:///path/to/dir/with/subdir/.lock/
	t.Nil(os.Chmod(tempdir, 0755))
	lpath := tempdir + "/.lock"
	t.Nil(os.Mkdir(lpath, 0755))

	t.Err(initialize(loc), syscall.EISDIR)
	t.Nil(os.Remove(lpath))

	// - file:///path/to/dir;/with/subdir/.lock.queue/
	lqpath := tempdir + "/.lock.queue"
	t.Nil(os.Mkdir(lqpath, 0755))

	t.Err(initialize(loc), syscall.EISDIR)
	t.Nil(os.Remove(lqpath))
	t.Nil(os.Remove(tempdir + "/.lock"))

	// - file:///path/to/dir/with/subdir/.version/
	vpath := tempdir + "/.version"
	t.Nil(os.Mkdir(vpath, 0755))

	t.Err(initialize(loc), syscall.EEXIST)
	t.Nil(os.Remove(lpath))
	t.Nil(os.Remove(lqpath))
	t.Nil(os.Remove(vpath))

	// - file:///path/to/empty/dir/  (success)
	t.Err(initialize(loc), nil)

	// - repeat initialize()
	t.Err(initialize(loc), errors.New("version already initialized at "+tempdir+"/.version"))
	cleanup(t, tempdir)
}

func TestNew(tt *testing.T) {
	t := check.T(tt)

	// - file:///path/to/empty/dir/
	tempdir, err := ioutil.TempDir("", "gotest")
	t.Nil(err)
	defer func() { t.Nil(os.Remove(tempdir)) }()
	loc, err := url.Parse(tempdir)
	t.Nil(err)

	_, err = new(loc)
	t.Err(err, errors.New("version is not initialized at "+tempdir+"/.version"))

	// - after initialize() (success)
	t.Nil(initialize(loc))
	defer cleanup(t, tempdir)
	_, err = new(loc)
	t.Err(err, nil)
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
func TestExSequence(tt *testing.T) {
	t := check.T(tt)

	tempdir, err := ioutil.TempDir("", "gotest")
	t.Nil(err)
	defer func() { t.Nil(os.Remove(tempdir)) }()
	loc, err := url.Parse("file://" + tempdir)
	t.Nil(err)
	t.Nil(initialize(loc))
	defer cleanup(t, tempdir)

	statusc := make(chan string)
	un1 := make(chan struct{})
	un2 := make(chan struct{})
	go testLock("EX1", loc, un1, statusc)
	t.Equal(<-statusc, "acquired EX1")
	un1 <- struct{}{}
	go testLock("EX2", loc, un2, statusc)
	t.Equal(<-statusc, "acquired EX2")
	un2 <- struct{}{}
}

// - EX1, EX2 (block), UN1, (unblock EX2), UN2
func TestExParallel(tt *testing.T) {
	t := check.T(tt)

	tempdir, err := ioutil.TempDir("", "gotest")
	t.Nil(err)
	defer func() { t.Nil(os.Remove(tempdir)) }()
	loc, err := url.Parse("file://" + tempdir)
	t.Nil(err)
	t.Nil(initialize(loc))
	defer cleanup(t, tempdir)

	statusc := make(chan string)
	un1 := make(chan struct{})
	un2 := make(chan struct{})
	go testLock("EX1", loc, un1, statusc)
	t.Equal(<-statusc, "acquired EX1")
	go testLock("EX2", loc, un2, statusc)
	t.Equal(<-statusc, "blocked EX2")
	un1 <- struct{}{}
	t.Equal(<-statusc, "acquired EX2")
	un2 <- struct{}{}
}

// - EX1, SH2 (block), UN1, (unblock SH2), UN2
func TestExShParallel(tt *testing.T) {
	t := check.T(tt)

	tempdir, err := ioutil.TempDir("", "gotest")
	t.Nil(err)
	defer func() { t.Nil(os.Remove(tempdir)) }()
	loc, err := url.Parse("file://" + tempdir)
	t.Nil(err)
	t.Nil(initialize(loc))
	defer cleanup(t, tempdir)

	statusc := make(chan string)
	un1 := make(chan struct{})
	un2 := make(chan struct{})
	go testLock("EX1", loc, un1, statusc)
	t.Equal(<-statusc, "acquired EX1")
	go testLock("SH2", loc, un2, statusc)
	t.Equal(<-statusc, "blocked SH2")
	un1 <- struct{}{}
	t.Equal(<-statusc, "acquired SH2")
	un2 <- struct{}{}
}

// - SH1, SH2, UN1, UN2
func TestShParallel(tt *testing.T) {
	t := check.T(tt)

	tempdir, err := ioutil.TempDir("", "gotest")
	t.Nil(err)
	defer func() { t.Nil(os.Remove(tempdir)) }()
	loc, err := url.Parse("file://" + tempdir)
	t.Nil(err)
	t.Nil(initialize(loc))
	defer cleanup(t, tempdir)

	statusc := make(chan string)
	un1 := make(chan struct{})
	un2 := make(chan struct{})
	go testLock("SH1", loc, un1, statusc)
	t.Equal(<-statusc, "acquired SH1")
	go testLock("SH2", loc, un2, statusc)
	t.Equal(<-statusc, "acquired SH2")
	un1 <- struct{}{}
	un2 <- struct{}{}
}

// - SH1, EX2 (block), SH3 (block), UN1, (unblock EX2), UN2, (unblock SH3), UN3
func TestExPriority(tt *testing.T) {
	t := check.T(tt)

	tempdir, err := ioutil.TempDir("", "gotest")
	t.Nil(err)
	defer func() { t.Nil(os.Remove(tempdir)) }()
	loc, err := url.Parse("file://" + tempdir)
	t.Nil(err)
	t.Nil(initialize(loc))
	defer cleanup(t, tempdir)

	statusc := make(chan string)
	un1 := make(chan struct{})
	un2 := make(chan struct{})
	un3 := make(chan struct{})
	go testLock("SH1", loc, un1, statusc)
	t.Equal(<-statusc, "acquired SH1")
	go testLock("EX2", loc, un2, statusc)
	t.Equal(<-statusc, "blocked EX2")
	go testLock("SH3", loc, un3, statusc)
	t.Equal(<-statusc, "blocked SH3")
	un1 <- struct{}{}
	t.Equal(<-statusc, "acquired EX2")
	un2 <- struct{}{}
	t.Equal(<-statusc, "acquired SH3")
	un3 <- struct{}{}
}

// - Get = "none", Get = "none"
func TestGetNone(tt *testing.T) {
	t := check.T(tt)

	tempdir, err := ioutil.TempDir("", "gotest")
	t.Nil(err)
	defer func() { t.Nil(os.Remove(tempdir)) }()
	loc, err := url.Parse("file://" + tempdir)
	t.Nil(err)
	t.Nil(initialize(loc))
	defer cleanup(t, tempdir)
	p, err := new(loc)
	t.Nil(err)

	t.Equal(p.Get(), "none")
	t.Equal(p.Get(), "none")
}

func TestSet(tt *testing.T) {
	t := check.T(tt)

	tempdir, err := ioutil.TempDir("", "gotest")
	t.Nil(err)
	defer func() { t.Nil(os.Remove(tempdir)) }()
	loc, err := url.Parse("file://" + tempdir)
	t.Nil(err)
	t.Nil(initialize(loc))
	defer cleanup(t, tempdir)
	p, err := new(loc)
	t.Nil(err)

	// - Set("") panics
	t.PanicMatch(func() { p.Set("") }, `no such file or directory`)

	cases := []struct {
		val  string
		want string
	}{
		{"dirty", "dirty"},
		{" ", " "},
		{"0", "0"},
		{"1.2.3", "1.2.3"},
	}

	for _, v := range cases {
		p.Set(v.val)
		t.Match(p.Get(), v.want)
	}
}

func cleanup(t *check.C, tempdir string) {
	t.Nil(os.Remove(tempdir + "/.lock"))
	t.Nil(os.Remove(tempdir + "/.lock.queue"))
	t.Nil(os.Remove(tempdir + "/.version"))
}
