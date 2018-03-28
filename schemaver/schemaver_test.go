package schemaver

import "testing"

// - file://user@/
// - file://localhost/
// - file://?a=1
// - file://#a
// - file://path/to/empty/dir/, success
// - registered[loc.Scheme] = nil, error "unknown protocol .."
func TestLocation(t *testing.T) {

}

// - already Initialized, error
// - file:///path/to/empty/dir/, success
func TestInitialize(t *testing.T) {

}

// - version is not initialized, error
// - after initialize, success
func TestNew(t *testing.T) {

}

// - lockType != unlocked, !SharedLock
// - recursive lock, unlock one time
// - Get(version)
// - callback(version)
// - SH, EX - panic
func TestSharedLock(t *testing.T) {

}

//- lockType = exclusive, !ExclusiveLock
// - recursive lock, unlock one time
// - Get(version)
// - callback(version)
// - "NARADA_SKIP_LOCK", !SharedLock, !ExclusiveLock, "NARADA_SKIP_LOCK" = ""
func TestExclusiveLock(t *testing.T) {

}

// - Get() (lockType==unlocked), panic
// - Get() (lockType==shared), success
// - Get() (lockType==exclusive), success
func TestGet(t *testing.T) {

}

// - Set() (lockType!= exclusive), panic
// - Set() (lockType==shared), success
// - Set() (lockType==exclusive), success
func TestSet(t *testing.T) {

}
