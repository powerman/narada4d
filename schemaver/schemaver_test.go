package schemaver_test

import (
	"context"
	"errors"
	urlpkg "net/url"
	"os"
	"sync"
	"testing"
	"testing/synctest"
	"time"

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
	tt.Parallel()
	t := check.Must(tt)

	// - register already registered protocol, panic
	t.PanicMatch(func() {
		schemaver.RegisterProtocol("test", schemaver.Backend{
			New:        mockNew,
			Initialize: mockInitialize,
		})
	}, `protocol "test" already registered`)

	// - RegisterProtocol(Backend{New == nil}), panic
	t.PanicMatch(func() {
		schemaver.RegisterProtocol("new", schemaver.Backend{Initialize: mockInitialize})
	}, `can't register protocol "new" with nil implementation`)
	// - RegisterProtocol(Backend{Initialize == nil}), panic
	t.PanicMatch(func() {
		schemaver.RegisterProtocol("new", schemaver.Backend{New: mockNew})
	}, `can't register protocol "new" with nil implementation`)
}

func TestLocation(tt *testing.T) {
	t := check.Must(tt)
	reset()

	// - test://localhost/, error
	tt.Setenv(schemaver.EnvLocation, "test://localhost/")
	t.Err(schemaver.Initialize(), errBadLocation)
	_, err := schemaver.New()
	t.Err(err, errBadLocation)

	// - test://, success
	tt.Setenv(schemaver.EnvLocation, "test://")
	t.Equal(schemaver.Initialize(), nil)
	_, err = schemaver.New()
	t.Nil(err)

	// - case registered[loc.Scheme] = nil, error
	tt.Setenv(schemaver.EnvLocation, "new://")
	t.Match(schemaver.Initialize(), `unknown protocol: "new"`)
	_, err = schemaver.New()
	t.Match(err, `unknown protocol: "new"`)
}

func TestInitialize(tt *testing.T) {
	t := check.Must(tt)
	reset()

	// - test:///ready, error
	tt.Setenv(schemaver.EnvLocation, "test:///ready")
	t.Err(schemaver.Initialize(), errInitialized)

	// - test:///empty, success
	tt.Setenv(schemaver.EnvLocation, "test:///empty")
	t.Err(schemaver.Initialize(), nil)
}

func TestNew(tt *testing.T) {
	t := check.Must(tt)
	reset()

	// - test:///invalid, error
	tt.Setenv(schemaver.EnvLocation, "test:///invalid")
	_, err := schemaver.New()
	t.Err(err, errInvalid)

	// - test:///ready, success
	tt.Setenv(schemaver.EnvLocation, "test:///ready")
	_, err = schemaver.New()
	t.Nil(err)
}

// - SH/EX (with backend, return version), UN (with backend).
// - NARADA_SKIP_LOCK=1, SH/EX (no backend, return version), UN (no backend).
func TestShExLock(tt *testing.T) { //nolint:paralleltest // Uses global state.
	t := check.Must(tt)
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

	for _, c := range cases {
		if c.setEnv {
			tt.Setenv(schemaver.EnvSkipLock, c.envValue)
		} else {
			os.Unsetenv(schemaver.EnvSkipLock)
		}
		v, err := schemaver.New()
		t.Nil(err)

		for _, counter := range []*int{&sh, &ex} {
			old := *counter
			oldun := un
			if counter == &sh {
				t.Equal(v.SharedLock(), "42")
			} else {
				t.Equal(v.ExclusiveLock(), "42")
			}
			v.Unlock()
			if c.wantBackend {
				t.Equal(*counter, old+1, "set=%v val=%q", c.setEnv, c.envValue)
				t.Equal(un, oldun+1)
			} else {
				t.Equal(*counter, old, "set=%v val=%q", c.setEnv, c.envValue)
				t.Equal(un, oldun)
			}
		}
	}
}

func TestUnlock(tt *testing.T) {
	t := check.Must(tt)
	reset()

	tt.Setenv(schemaver.EnvLocation, "test://")
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
	t := check.Must(tt)
	reset()

	tt.Setenv(schemaver.EnvLocation, "test://")
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
	t := check.Must(tt)
	reset()

	tt.Setenv(schemaver.EnvLocation, "test://")
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
	v.Unlock()
}

func TestRecursiveLocks(tt *testing.T) {
	t := check.Must(tt)
	reset()

	tt.Setenv(schemaver.EnvLocation, "test://")
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

func TestHoldSharedLock(tt *testing.T) {
	synctest.Test(tt, func(tt *testing.T) {
		t := check.Must(tt)
		reset()

		tt.Setenv(schemaver.EnvLocation, "test://")
		v, err := schemaver.New()
		t.Nil(err)

		ctx, cancel := context.WithCancel(t.Context())
		v.HoldSharedLock(ctx, time.Second/10)
		time.Sleep(time.Second / 2)
		mu.Lock()
		t.Between(sh, 4, 7)
		mu.Unlock()
		cancel()
		time.Sleep(time.Second / 2)
		mu.Lock()
		t.Equal(sh, 6)
		mu.Unlock()
		t.Nil(v.Close())
	})
}

func TestHoldSharedLock2(tt *testing.T) {
	synctest.Test(tt, func(tt *testing.T) {
		t := check.Must(tt)
		reset()

		tt.Setenv(schemaver.EnvLocation, "test://")
		v, err := schemaver.New()
		t.Nil(err)

		ctx, cancel := context.WithCancel(t.Context())
		v.HoldSharedLock(ctx, time.Second/10)
		v.HoldSharedLock(t.Context(), time.Second/10)
		time.Sleep(time.Second / 2)
		mu.Lock()
		t.Equal(sh, 1)
		mu.Unlock()
		cancel()
		time.Sleep(time.Second / 2)
		mu.Lock()
		t.Between(sh, 4, 8)
		mu.Unlock()
		t.Nil(v.Close())
	})
}

func TestAddCallback(tt *testing.T) {
	t := check.Must(tt)
	reset()

	tt.Setenv(schemaver.EnvLocation, "test://")
	v, err := schemaver.New()
	t.Nil(err)

	// - AddCallback(nil) - panic
	t.PanicMatch(func() { v.AddCallback(nil) }, `require callback`)

	// - AddCallback, SH (callback), EX (panic, no callback), UN
	call1, call2 := 0, 0
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
	for _, cb2 := range []func(string){nil, func(string) { call2++ }} {
		for _, skipLock := range []bool{false, true} {
			v, err = schemaver.New()
			t.Nil(err)
			reset()
			call1, call2 = 0, 0
			v.AddCallback(cb1)
			if cb2 != nil {
				v.AddCallback(cb2)
			}
			if skipLock {
				tt.Setenv(schemaver.EnvSkipLock, "1")
			} else {
				os.Unsetenv(schemaver.EnvSkipLock)
			}
			v.SharedLock()
			t.Equal(call1, 1)
			if cb2 == nil {
				t.Equal(call2, 0)
			} else {
				t.Equal(call2, 1)
			}
			v.Unlock()
			v.ExclusiveLock()
			t.Equal(call1, 2)
			if cb2 == nil {
				t.Equal(call2, 0)
			} else {
				t.Equal(call2, 2)
			}
			v.Unlock()
		}
	}

	// - SH/EX, callback - panic, UN
	reset()
	tt.Setenv(schemaver.EnvLocation, "test://")
	v, err = schemaver.New()
	t.Nil(err)

	cb := func(string) { panic(`stopped`) }
	v.AddCallback(cb)
	t.PanicMatch(func() { v.SharedLock() }, `stopped`)
	v.Unlock()
	t.Equal(un, 1)
	t.PanicMatch(func() { v.ExclusiveLock() }, `stopped`)
	v.Unlock()
	t.Equal(un, 2)
}

var (
	errBadLocation = errors.New("location must not contain host")
	errInitialized = errors.New("version already initialized")
	errInvalid     = errors.New("version is invalid")
	mu             sync.Mutex
	sh, ex, un     int
	ver            string
)

func reset() {
	os.Unsetenv(schemaver.EnvSkipLock)
	os.Setenv(schemaver.EnvLocation, "test://")
	ver, sh, ex, un = "42", 0, 0, 0
}

func mockInitialize(loc *urlpkg.URL) error {
	if loc.Host != "" {
		return errBadLocation
	}
	if loc.Path == "/ready" {
		return errInitialized
	}
	return nil
}

func mockNew(loc *urlpkg.URL) (schemaver.Manage, error) {
	if loc.Host != "" {
		return nil, errBadLocation
	}
	if loc.Path == "/invalid" {
		return nil, errInvalid
	}
	return &mockManage{}, nil
}

type mockManage struct{}

func (*mockManage) SharedLock()    { mu.Lock(); sh++; mu.Unlock() }
func (*mockManage) ExclusiveLock() { mu.Lock(); ex++; mu.Unlock() }
func (*mockManage) Unlock()        { mu.Lock(); un++; mu.Unlock() }
func (*mockManage) Get() string    { return ver }
func (*mockManage) Set(v string)   { ver = v }
func (*mockManage) Close() error   { return nil }
