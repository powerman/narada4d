package schemaver_test

import (
	"testing"
)

// - test://localhost/, error
// - test://, success
// - registered[loc.Scheme] = nil, error "unknown protocol .."
func TestLocation(tt *testing.T) {

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
func TestSharedLock(t *testing.T) {

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
