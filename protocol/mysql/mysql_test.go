// +build integration

package mysql

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/powerman/check"
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

	require := "require mysql://username[:password]@host[:port]/database"
	cases := []struct {
		url     string
		wanterr error
	}{
		{fmt.Sprintf("mysql://%s:%s@%s:%s/%s", dbUser, dbPass, dbHost, dbPort, dbName), nil},
		{fmt.Sprintf("mysql://%s:%s@%s:%s/", dbUser, dbPass, dbHost, dbPort), errors.New("database absent, " + require)},
		{fmt.Sprintf("mysql://%s:%s@%s:%s", dbUser, dbPass, dbHost, dbPort), errors.New("database absent, " + require)},
		{fmt.Sprintf("mysql://%s:%s@/%s", dbUser, dbPass, dbName), errors.New("host absent, " + require)},
		{fmt.Sprintf("mysql://:%s@%s:%s/%s", dbPass, dbHost, dbPort, dbName), errors.New("username absent, " + require)},
		{fmt.Sprintf("mysql://%s:%s@%s:%s/%s?a=3", dbUser, dbPass, dbHost, dbPort, dbName), errors.New("unexpected query params or fragment, " + require)},
		{fmt.Sprintf("mysql://%s:%s@%s:%s/%s#a", dbUser, dbPass, dbHost, dbPort, dbName), errors.New("unexpected query params or fragment, " + require)},
		{"mysql://", errors.New("username absent, " + require)},
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

	//- Not initialized()
	t.False(s.initialized())

	//- Initialized()
	t.Nil(initialize(loc))
	t.True(s.initialized())
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

	s, err := newStorage(loc)
	t.Nil(err)

	t.PanicMatch(func() { s.SharedLock() }, `doesn't exist`)
}

func TestGet(tt *testing.T) {
	t := check.T(tt)

	v, err := newInitializedStorage(loc)
	t.Nil(err)
	defer dropTable(t)

	v.SharedLock()
	t.Equal(v.Get(), "none")
	v.Unlock()
}

func TestSet(tt *testing.T) {
	t := check.T(tt)

	c, err := newInitializedStorage(loc)
	t.Nil(err)
	defer dropTable(t)

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

	c.ExclusiveLock()
	defer c.Unlock()
	for _, v := range cases {
		v := v
		if v.wantpanic {
			t.PanicMatch(func() { c.Set(v.val) }, `invalid version value, require 'none' or 'dirty' or one or more digits separated with single dots`)
		} else {
			t.NotPanic(func() { c.Set(v.val) })
			t.Equal(c.Get(), v.val)
		}
	}
}
