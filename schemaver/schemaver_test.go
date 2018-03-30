package schemaver

import (
	"testing"
)

// - test://localhost/, error
// - test://path/to/empty/dir/, success
// - registered[loc.Scheme] = nil, error "unknown protocol .."
func TestLocation(tt *testing.T) {

}

// - already Initialized, error
// - file:///path/to/empty/dir/, success
func TestInitialize(t *testing.T) {

}

// - version is not initialized, error
// - after initialize, success
func TestNew(t *testing.T) {

}

// - SH(with backend), Get(version), UN(with backend)
func TestSharedLock(t *testing.T) {

}

// - EX(with backend), Get(version), Set(version), UN(with backend)
// - "NARADA_SKIP_LOCK"=1, EX(no backend), UN(no backend), "NARADA_SKIP_LOCK"=""
func TestExclusiveLock(t *testing.T) {

}

// - Get() (lockType==unlocked), panic
// - Get() (lockType==shared), success
// - Get() (lockType==exclusive), success
func TestGet(t *testing.T) {

}

// - Set() (lockType!= exclusive), panic
// - Set() (lockType==shared), panic
// - Set() (lockType==exclusive), success
func TestSet(t *testing.T) {

}

// - "NARADA_SKIP_LOCK"=1, SH(no backend), UN(no backend)
// - SH(with backend), SH(no backend), SH(no backend), UN(with backend)
// - EX(with backend), SH(no backend), UN(no backend), UN(with backend)
// - SH(with backend)
//   - SH(no backend), UN(no backend)
//   - SH(no backend), SH(no backend), UN(no backend), UN(no backend)
//   - UN(with backend)
// - SH, EX - panic
// - UN(with backend) - panic
// - SH(with backend), UN(with backend), UN(with backend) - panic
// - EX(with backend), UN(with backend), UN(with backend) - panic
func TestShEx(t *testing.T) {

}

// - Unlock, callback(version), os.Exit()
// - SH(with backend), Get(callback), UN(with backend)
// - SH(with backend), Get(callback), SH(no backend), Get(callback), UN(no backend), UN(with backend)
// - AddCallback(0,1,2)
//   - SH/EX(with backend), callback(0,1,2) UN(with backend)
// - callback - panic, unlock
func TestCallback(t *testing.T) {

}
