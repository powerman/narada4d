// +build integration

package mysql

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/powerman/check"

	"github.com/powerman/narada4d/internal"
)

func TestConnect(tt *testing.T) {
	t := check.T(tt)

	var (
		dbUser    = loc.User.Username()
		dbPass, _ = loc.User.Password()
		dbHost    = loc.Hostname()
		dbPort    = loc.Port()
		dbName    = strings.TrimPrefix(loc.Path, "/")
	)

	cases := []struct {
		url     string
		wanterr error
	}{
		{fmt.Sprintf("mysql://%s:%s@%s:%s/%s", dbUser, dbPass, dbHost, dbPort, dbName), nil},
		{fmt.Sprintf("mysql://%s:%s@%s:%s/", dbUser, dbPass, dbHost, dbPort), errLocationRequireDB},
		{fmt.Sprintf("mysql://%s:%s@%s:%s", dbUser, dbPass, dbHost, dbPort), errLocationRequireDB},
		{fmt.Sprintf("mysql://%s:%s@/%s", dbUser, dbPass, dbName), errLocationRequireHost},
		{fmt.Sprintf("mysql://:%s@%s:%s/%s", dbPass, dbHost, dbPort, dbName), errLocationRequireUsername},
		{fmt.Sprintf("mysql://%s:%s@%s:%s/%s?a=3", dbUser, dbPass, dbHost, dbPort, dbName), errLocationInvalid},
		{fmt.Sprintf("mysql://%s:%s@%s:%s/%s#a", dbUser, dbPass, dbHost, dbPort, dbName), errLocationInvalid},
		{"mysql://", errLocationRequireUsername},
	}

	for _, v := range cases {
		p, err := url.Parse(v.url)
		t.Nil(err)
		s, err := newStorage(p)
		t.Err(err, v.wanterr)
		if v.wanterr == nil {
			s.Close()
		}
	}

	p, err := url.Parse(fmt.Sprintf("mysql://incUserName:%s@%s:%s/%s", dbPass, dbHost, dbPort, dbName))
	t.Nil(err)
	_, err = newStorage(p)
	t.Match(err, `Access denied`)

	p, err = url.Parse(fmt.Sprintf("mysql://%s:incPass@%s:%s/%s", dbUser, dbHost, dbPort, dbName))
	t.Nil(err)
	_, err = newStorage(p)
	t.Match(err, `Access denied`)
}

func TestInitialize(tt *testing.T) {
	t := check.T(tt)
	t.Nil(initialize(loc))
	dropTable(t)
}

func TestInitialized(tt *testing.T) {
	t := check.T(tt)

	s, err := newStorage(loc)
	t.Nil(err)
	defer s.Close()

	//- Not initialized()
	t.False(s.initialized())

	//- Initialized()
	t.Nil(initialize(loc))
	t.True(s.initialized())
	dropTable(t)
}

// - EX1, UN1, EX2, UN2.
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

// - EX1, EX2(block), UN1, (unblockEX2), UN2.
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

// - EX1, SH2(block), UN1, (unblock)SH2, UN2.
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

// - SH1, SH2, UN1, UN2.
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

// - SH1, EX2(block), SH3(block), UN1, (unblock)EX2, UN2, (unblock)SH3, UN3.
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

	s, err := newStorage(loc)
	t.Nil(err)
	defer s.Close()

	t.PanicMatch(func() { s.SharedLock() }, `doesn't exist`)
	s.tx.Rollback()
}

func TestGet(tt *testing.T) {
	t := check.T(tt)

	v, err := newInitializedStorage(loc)
	t.Nil(err)
	defer dropTable(t)
	defer v.Close()

	v.SharedLock()
	t.Equal(v.Get(), "none")
	v.Unlock()
}

func TestSet(tt *testing.T) {
	t := check.T(tt)

	v, err := newInitializedStorage(loc)
	t.Nil(err)
	defer dropTable(t)
	defer v.Close()

	cases := []struct {
		val       string
		wantpanic bool
	}{
		{"42.", true},
		{"42..", true},
		{".42", true},
		{"-42", true},
		{"", true},
		{"rat", true},
		{"v1.2.3", true},
		{"None", true},
		{"none", false},
		{"dirty", false},
		{"43", false},
		{"0", false},
		{"43.0.1", false},
	}

	v.ExclusiveLock()
	defer v.Unlock()
	for _, tc := range cases {
		tc := tc
		if tc.wantpanic {
			t.PanicMatch(func() { v.Set(tc.val) }, `invalid version value, require 'none' or 'dirty' or one or more digits separated with single dots`)
		} else {
			t.NotPanic(func() { v.Set(tc.val) })
			t.Equal(v.Get(), tc.val)
		}
	}
}

func TestReconnect(tt *testing.T) {
	t := check.T(tt)

	v, err := newInitializedStorage(loc)
	t.Nil(err)
	defer dropTable(t)
	defer v.Close()

	restartProxy := func() {
		proxy.Close()
		t.Nil(internal.WaitTCPPortClosed(ctx, proxy.FrontendAddr()))
		go func() {
			var err error
			time.Sleep(time.Second)
			proxy, err = internal.NewTCPProxy(ctx, proxy.FrontendAddr().String(), proxy.BackendAddr().String())
			t.Nil(err)
		}()
	}

	v.SharedLock()
	restartProxy()
	t.NotPanic(v.Unlock)

	t.NotPanic(v.SharedLock)
	v.Unlock()

	restartProxy()
	t.NotPanic(v.ExclusiveLock)
	v.Unlock()
}
