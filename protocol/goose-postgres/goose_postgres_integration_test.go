// +build integration

package goosepostgres

import (
	"testing"

	"github.com/powerman/check"
)

func TestInitialize(tt *testing.T) {
	t := check.T(tt)
	t.Nil(initialize(loc))
	dropTable(t)
}

func TestInitialized(tt *testing.T) {
	t := check.T(tt)

	v, err := newStorage(loc)
	t.Nil(err)
	s := v.(*storage)
	defer s.db.Close()

	//- Not initialized()
	t.False(initialized(s.db))

	//- Initialized()
	t.Nil(initialize(loc))
	t.True(initialized(s.db))
	dropTable(t)
}

// - EX1, UN1, EX2, UN2
func TestExSequence(tt *testing.T) {
	t := check.T(tt)

	t.Nil(initialize(loc))
	defer dropTable(t)

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

// - EX1, EX2(block), UN1, (unblockEX2), UN2
func TestExParallel(tt *testing.T) {
	t := check.T(tt)

	t.Nil(initialize(loc))
	defer dropTable(t)

	statusc := make(chan string)
	un1 := make(chan struct{})
	un2 := make(chan struct{})
	go testLock("EX1", loc, un1, statusc)
	t.Equal(<-statusc, "acquired EX1")
	go testLock("EX2", loc, un2, statusc)
	t.Equal(<-statusc, "block EX2")
	un1 <- struct{}{}
	t.Equal(<-statusc, "acquired EX2")
	un2 <- struct{}{}
}

// - EX1, SH2(block), UN1, (unblock)SH2, UN2
func TestExShParallel(tt *testing.T) {
	t := check.T(tt)

	t.Nil(initialize(loc))
	defer dropTable(t)

	statusc := make(chan string)
	un1 := make(chan struct{})
	un2 := make(chan struct{})
	go testLock("EX1", loc, un1, statusc)
	t.Equal(<-statusc, "acquired EX1")
	go testLock("SH2", loc, un2, statusc)
	t.Equal(<-statusc, "block SH2")
	un1 <- struct{}{}
	t.Equal(<-statusc, "acquired SH2")
	un2 <- struct{}{}
}

// - SH1, SH2, UN1, UN2
func TestShParallel(tt *testing.T) {
	t := check.T(tt)

	t.Nil(initialize(loc))
	defer dropTable(t)

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

// - SH1, EX2(block), SH3(block), UN1, (unblock)EX2, UN2, (unblock)SH3, UN3
func TestExPriority(tt *testing.T) {
	t := check.T(tt)

	t.Nil(initialize(loc))
	defer dropTable(t)

	statusc := make(chan string)
	un1 := make(chan struct{})
	un2 := make(chan struct{})
	un3 := make(chan struct{})
	go testLock("SH1", loc, un1, statusc)
	t.Equal(<-statusc, "acquired SH1")
	go testLock("EX2", loc, un2, statusc)
	t.Equal(<-statusc, "block EX2")
	go testLock("SH3", loc, un3, statusc)
	t.Equal(<-statusc, "block SH3")
	un1 <- struct{}{}
	t.Equal(<-statusc, "acquired EX2")
	un2 <- struct{}{}
	t.Equal(<-statusc, "acquired SH3")
	un3 <- struct{}{}
}

func TestNotInitialized(tt *testing.T) {
	t := check.T(tt)

	v, err := newStorage(loc)
	t.Nil(err)
	defer v.(*storage).db.Close()

	t.PanicMatch(func() { v.SharedLock() }, `"goose_db_version" does not exist`)
	defer v.(*storage).tx.Rollback()
}

func TestGet(tt *testing.T) {
	t := check.T(tt)

	v, err := newStorage(loc)
	t.Nil(err)
	t.Nil(initialize(loc))
	defer dropTable(t)
	defer v.(*storage).db.Close()

	v.SharedLock()
	t.Equal(v.Get(), "none")
	v.Unlock()
}

func TestSet(tt *testing.T) {
	t := check.T(tt)

	v, err := newStorage(loc)
	t.Nil(err)
	t.Nil(initialize(loc))
	defer dropTable(t)
	defer v.(*storage).db.Close()

	v.ExclusiveLock()
	t.PanicMatch(func() { v.Set("42") }, `not supported`)
	v.Unlock()
}
