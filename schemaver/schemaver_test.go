package schemaver_test

import (
	"errors"
	"net/url"
	"os"
	"testing"

	"github.com/powerman/check"
	"github.com/powerman/narada4d/schemaver"
)

func init() {
	schemaver.RegisterProtocol("test", schemaver.Backend{
		New:        mockNew,
		Initialize: mockInitialize,
	})
}

func TestRegisterProtocol(tt *testing.T) {
	t := check.T(tt)

	// - registered[file://], panic ("protocol already registered")
	t.PanicMatch(func() {
		schemaver.RegisterProtocol("test", schemaver.Backend{
			New:        mockNew,
			Initialize: mockInitialize,
		})
	}, `protocol \"test\" already registered`)

	// - registered[new://], beckend.New == nil, panic ("can't register protocol with nil implementation")
	t.PanicMatch(func() {
		schemaver.RegisterProtocol("new", schemaver.Backend{
			Initialize: mockInitialize})
	}, `can't register protocol \"new\" with nil implementation`)
	// - registered[new://], backend.Initialize == nil, panic ("can't register protocol with nil implementation")
	t.PanicMatch(func() {
		schemaver.RegisterProtocol("new", schemaver.Backend{
			New: mockNew})
	}, `can't register protocol \"new\" with nil implementation`)
}

func TestLocation(tt *testing.T) {
	t := check.T(tt)
	reset()

	// - test://localhost/, error
	os.Setenv(schemaver.EnvLocation, "test://localhost/")
	t.Err(schemaver.Initialize(), errBadLocation)
	_, err := schemaver.New()
	t.Err(err, errors.New("location must not contain host"))

	// - test://, success
	os.Setenv(schemaver.EnvLocation, "test://")
	t.Equal(schemaver.Initialize(), nil)
	_, err = schemaver.New()
	t.Equal(err, nil)

	// - registered[loc.Scheme] = nil, error "unknown protocol .."
	os.Setenv(schemaver.EnvLocation, "new://")
	t.Err(schemaver.Initialize(), errors.New("unknown protocol in $NARADA4D: \"new\""))
	_, err = schemaver.New()
	t.Err(err, errors.New("unknown protocol in $NARADA4D: \"new\""))
}

func TestInitialize(tt *testing.T) {
	t := check.T(tt)
	reset()

	// - test:///ready, error
	os.Setenv(schemaver.EnvLocation, "test:///ready")
	t.Err(schemaver.Initialize(), errInitialized)

	// - test:///empty, success
	os.Setenv(schemaver.EnvLocation, "test:///empty")
	t.Err(schemaver.Initialize(), nil)
}

func TestNew(tt *testing.T) {
	t := check.T(tt)
	reset()

	// - test:///empty, error
	os.Setenv(schemaver.EnvLocation, "test:///empty")
	_, err := schemaver.New()
	t.Err(err, errNotInitialized)

	// - test:///ready, success
	os.Setenv(schemaver.EnvLocation, "test:///ready")
	_, err = schemaver.New()
	t.Equal(err, nil)
}

func TestShExLock(tt *testing.T) {
	t := check.T(tt)
	reset()

	cases := []struct {
		setEnv      bool
		envValue    string
		wantBackend bool
	}{
		{false, "", true},
		{true, "", true},
		{true, "1", false},
		{true, "anything", false},
	}

	// - SH (with backend, return version), UN (with backend)
	// - NARADA_SKIP_LOCK=1, SH (no backend, return version), UN (no backend)
	for _, c := range cases {
		if c.setEnv {
			os.Setenv(schemaver.EnvSkipLock, c.envValue)
		} else {
			os.Unsetenv(schemaver.EnvSkipLock)
		}
		v, err := schemaver.New()
		t.Nil(err)

		old := sh
		oldun := un
		t.Equal(v.SharedLock(), "42")
		v.Unlock()
		if c.wantBackend {
			t.Equal(sh, old+1, "set=%v val=%q", c.setEnv, c.envValue)
			t.Equal(un, oldun+1)
		} else {
			t.Equal(sh, old, "set=%v val=%q", c.setEnv, c.envValue)
			t.Equal(un, oldun)
		}
	}

	// - EX (with backend, return version), UN (with backend)
	// - NARADA_SKIP_LOCK=1, EX (no backend, return version), UN (no backend)
	for _, c := range cases {
		if c.setEnv {
			os.Setenv(schemaver.EnvSkipLock, c.envValue)
		} else {
			os.Unsetenv(schemaver.EnvSkipLock)
		}
		v, err := schemaver.New()
		t.Nil(err)

		old := ex
		oldun := un
		t.Equal(v.ExclusiveLock(), "42")
		v.Unlock()
		if c.wantBackend {
			t.Equal(ex, old+1, "set=%v val=%q", c.setEnv, c.envValue)
			t.Equal(un, oldun+1)
		} else {
			t.Equal(ex, old, "set=%v val=%q", c.setEnv, c.envValue)
			t.Equal(un, oldun)
		}
	}
}

func TestUnlock(tt *testing.T) {
	t := check.T(tt)
	reset()

	os.Setenv(schemaver.EnvLocation, "test://")
	v, err := schemaver.New()
	t.Nil(err)

	// - UN - panic
	t.PanicMatch(func() { v.Unlock() }, `can't unlock, no lock acquired`)

	// - SH, UN, UN - panic
	v.SharedLock()
	v.Unlock()
	t.PanicMatch(func() { v.Unlock() }, `can't unlock, no lock acquired`)

	// - EX, UN, UN - panic
	v.ExclusiveLock()
	v.Unlock()
	t.PanicMatch(func() { v.Unlock() }, `can't unlock, no lock acquired`)
}

func TestGet(tt *testing.T) {
	t := check.T(tt)
	reset()

	os.Setenv(schemaver.EnvLocation, "test://")
	v, err := schemaver.New()
	t.Nil(err)

	// - Get() (lockType==unlocked), panic
	t.PanicMatch(func() { v.Get() }, `require SharedLock or ExclusiveLock`)

	// - Get() (lockType==shared), success
	v.SharedLock()
	t.Equal(v.Get(), "42")
	v.Unlock()

	// - Get() (lockType==exclusive), success
	v.ExclusiveLock()
	t.Equal(v.Get(), "42")
	v.Unlock()
}

func TestSet(tt *testing.T) {
	t := check.T(tt)
	reset()

	os.Setenv(schemaver.EnvLocation, "test://")
	v, err := schemaver.New()
	t.Nil(err)

	// - Set() (lockType==unlocked), panic
	t.PanicMatch(func() { v.Set("13") }, `require ExclusiveLock`)

	// - Set() (lockType==shared), panic
	v.SharedLock()
	t.PanicMatch(func() { v.Set("13") }, `require ExclusiveLock`)
	v.Unlock()

	// - Set() (lockType==exclusive), success
	v.ExclusiveLock()
	v.Set("13")
	t.Match(v.Get(), "13")
	v.Unlock()
}

func TestRecursiveLocks(tt *testing.T) {
	t := check.T(tt)
	reset()

	os.Setenv(schemaver.EnvLocation, "test://")
	v, err := schemaver.New()
	t.Nil(err)

	// - EX (with backend), EX (no backend), SH (no backend), UN (no backend), UN (no backend), UN (with backend)
	v.ExclusiveLock()
	t.Equal(ex, 1)
	v.ExclusiveLock()
	t.Equal(ex, 1)
	v.SharedLock()
	t.Equal(sh, 0)
	v.Unlock()
	t.Equal(un, 0)
	v.Unlock()
	t.Equal(un, 0)
	v.Unlock()
	t.Equal(un, 1)

	// - SH (with backend)
	//   - SH (no backend), UN (no backend)
	//   - SH (no backend), SH (no backend), UN (no backend), UN (no backend)
	//   - UN (with backend)
	reset()
	v.SharedLock()
	t.Equal(sh, 1)
	v.SharedLock()
	t.Equal(sh, 1)
	v.Unlock()
	t.Equal(un, 0)
	v.SharedLock()
	t.Equal(sh, 1)
	v.SharedLock()
	t.Equal(sh, 1)
	v.Unlock()
	t.Equal(un, 0)
	v.Unlock()
	t.Equal(un, 0)
	v.Unlock()
	t.Equal(un, 1)

	// - SH, EX - panic
	v.SharedLock()
	t.PanicMatch(func() { v.ExclusiveLock() }, `unable to acquire exclusive lock under shared lock`)
}

func TestAddCallback(tt *testing.T) {
	t := check.T(tt)
	reset()

	os.Setenv(schemaver.EnvLocation, "test://")
	v, err := schemaver.New()
	t.Nil(err)

	// - AddCallback(nil) - panic
	t.PanicMatch(func() { v.AddCallback(nil) }, `require callback`)

	// - AddCallback, SH (callback), EX (panic, no callback), UN
	cb1 := func(string) { call1++ }
	v.AddCallback(cb1)
	v.SharedLock()
	t.Equal(call1, 1)
	t.PanicMatch(func() { v.ExclusiveLock() }, `unable to acquire exclusive lock under shared lock`)
	t.Equal(call1, 1)
	v.Unlock()

	// - AddCallback (1,2)
	//   - NARADA_SKIP_LOCK=, SH, callback(1,2), UN
	//   - NARADA_SKIP_LOCK=, EX, callback(1,2), UN
	//   - NARADA_SKIP_LOCK=1, SH, callback(1,2), UN
	//   - NARADA_SKIP_LOCK=1, EX, callback(1,2), UN
	reset()
	cb2 := func(string) { call2++ }
	v.AddCallback(cb2)
	v.SharedLock()
	t.Equal(call1, 1)
	t.Equal(call2, 1)
	v.Unlock()
	v.ExclusiveLock()
	t.Equal(call1, 2)
	t.Equal(call2, 2)
	v.Unlock()
	os.Setenv(schemaver.EnvSkipLock, "1")
	v.SharedLock()
	t.Equal(call1, 3)
	t.Equal(call2, 3)
	v.Unlock()
	v.ExclusiveLock()
	t.Equal(call1, 4)
	t.Equal(call2, 4)
	v.Unlock()

	// - SH/EX, callback - panic, UN   !!! TODO
	reset()
	v.SharedLock()
	t.Equal(call1, 1)
	t.PanicMatch(func() { v.ExclusiveLock() }, `unable to acquire exclusive lock under shared lock`)
	t.Equal(call1, 1)
	v.Unlock()

}

var (
	errBadLocation           = errors.New("location must not contain host")
	errInitialized           = errors.New("version already initialized")
	errNotInitialized        = errors.New("version is not initialized")
	sh, ex, un, call1, call2 int
	ver                      string
)

func reset() {
	os.Unsetenv(schemaver.EnvSkipLock)
	os.Setenv(schemaver.EnvLocation, "test://")
	ver, sh, ex, un, call1, call2 = "42", 0, 0, 0, 0, 0
}

func mockInitialize(loc *url.URL) error {
	if loc.Host != "" {
		return errBadLocation
	}
	if loc.Path == "/ready" {
		return errInitialized
	}
	return nil
}

func mockNew(loc *url.URL) (schemaver.Manage, error) {
	if loc.Host != "" {
		return nil, errBadLocation
	}
	if loc.Path == "/empty" {
		return nil, errNotInitialized
	}
	return &mockManage{}, nil
}

type mockManage struct{}

func (m *mockManage) SharedLock()    { sh++ }
func (m *mockManage) ExclusiveLock() { ex++ }
func (m *mockManage) Unlock()        { un++ }
func (m *mockManage) Get() string    { return ver }
func (m *mockManage) Set(v string)   { ver = v }
