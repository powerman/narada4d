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

// TODO
func TestRegisterProtocol(tt *testing.T) {

}

// - test://localhost/, error
// - test://, success
// - registered[loc.Scheme] = nil, error "unknown protocol .."
func TestLocation(tt *testing.T) {
	t := check.T(tt)
	os.Setenv(schemaver.EnvLocation, "test://localhost/")
	t.Err(schemaver.Initialize(), errBadLocation)

	// TODO
}

// - test:///ready, error
// - test:///empty, success
func TestInitialize(t *testing.T) {

}

// - test:///empty, error
// - test:///ready, success
func TestNew(t *testing.T) {

}

// - SH (with backend, return version), UN (with backend)
// - NARADA_SKIP_LOCK=1, SH (no backend, return version), UN (no backend)
func TestSharedLock(tt *testing.T) {
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
	for _, c := range cases {
		if c.setEnv {
			os.Setenv(schemaver.EnvSkipLock, c.envValue)
		} else {
			os.Unsetenv(schemaver.EnvSkipLock)
		}
		v, err := schemaver.New()
		t.Nil(err)

		old := sh
		t.Equal(v.SharedLock(), "42")
		if c.wantBackend {
			t.Equal(sh, old+1, "set=%v val=%q", c.setEnv, c.envValue)
		} else {
			t.Equal(sh, old, "set=%v val=%q", c.setEnv, c.envValue)
		}
	}
}

// - EX (with backend, return version), UN (with backend)
// - NARADA_SKIP_LOCK=1, EX (no backend, return version), UN (no backend)
func TestExclusiveLock(t *testing.T) {

}

// - UN - panic
// - SH, UN, UN - panic
// - EX, UN, UN - panic
func TestUnlock(t *testing.T) {

}

// - Get() (lockType==unlocked), panic
// - Get() (lockType==shared), success
// - Get() (lockType==exclusive), success
func TestGet(t *testing.T) {

}

// - Set() (lockType==unlocked), panic
// - Set() (lockType==shared), panic
// - Set() (lockType==exclusive), success
func TestSet(t *testing.T) {

}

// - EX (with backend), EX (no backend), SH (no backend), UN (no backend), UN (no backend), UN (with backend)
// - SH (with backend)
//   - SH (no backend), UN (no backend)
//   - SH (no backend), SH (no backend), UN (no backend), UN (no backend)
//   - UN (with backend)
// - SH, EX - panic
func TestRecursiveLocks(t *testing.T) {

}

// - if EX not acquired than callback not called
// - AddCallback (1,2)
//   - NARADA_SKIP_LOCK=, SH/EX, callback(1,2), UN
//   - NARADA_SKIP_LOCK=1, SH/EX, callback(1,2), UN
// - SH/EX, callback - panic, UN
func TestCallback(t *testing.T) {

}

var (
	errBadLocation    = errors.New("location must not contain host")
	errInitialized    = errors.New("version already initialized")
	errNotInitialized = errors.New("version is not initialized")
	sh, ex, un        int
	ver               string
)

func reset() {
	os.Unsetenv(schemaver.EnvSkipLock)
	os.Setenv(schemaver.EnvLocation, "test://")
	ver, sh, ex, un = "42", 0, 0, 0
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
