package schemaver

import "testing"

// - EnvLocation = "", error
// - EnvLocation = " ", error
// - EnvLocation = "NARADA4D" success
// - registered[loc.Scheme] = nil, error "unknown protocol .."
func TestLocation(t *testing.T) {

}

// - already Initialized, error
func TestInitialize(t *testing.T) {

}

// - version is not initialized, error
// - EnvSkipLock != "", locktype = exclusive
func TestNew(t *testing.T) {

}

// - EX1, SH1, SH1, UN1
// - SH1, SH2, UN1, UN2
func TestSharedLock(t *testing.T) {

}

// - EX1, EX1, SH1, UN1
// - SH1, EX1, panic
func TestExclusiveLock(t *testing.T) {

}

// - Get() (lockType==unlocked), panic
func TestGet(t *testing.T) {

}

// - Set() (lockType!= exclusive), panic
func TestSet(t *testing.T) {

}
