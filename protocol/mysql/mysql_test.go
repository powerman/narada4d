package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/powerman/check"
)

const (
	dbName           = "gotest"
	dbHost           = "127.0.0.1"
	dbUser           = "gotestuser"
	dbPass           = "gotestpass"
	dropTableVersion = "DROP TABLE Narada4D"
)

var dbPort string
var locUser, locRoot *url.URL

var dockerIDs []string

func TestMain(m *testing.M) {
	_, err := docker("info")
	if err != nil {
		fmt.Println("SKIP:", err)
		os.Exit(0)
	}

	dbPort, err = runMySQL()
	if err != nil {
		fmt.Println(err)
		dockerCleanup()
		os.Exit(1)
	}

	locUser, err = url.Parse(fmt.Sprintf("mysql://%s:%s@%s:%s/%s", dbUser, dbPass, dbHost, dbPort, dbName))
	if err != nil {
		panic(err)
	}
	locRoot, err = url.Parse(fmt.Sprintf("mysql://root@%s:%s/%s", dbHost, dbPort, dbName))
	if err != nil {
		panic(err)
	}

	code := m.Run()
	check.Report()
	dockerCleanup()
	os.Exit(code)
}

func TestConnect(tt *testing.T) {
	t := check.T(tt)

	require := "require mysql://username[:password]@host[:port]/database"
	cases := []struct {
		url     string
		wanterr error
	}{
		{fmt.Sprintf("mysql://%s:%s@%s:%s/%s", dbUser, dbPass, dbHost, dbPort, dbName), nil},
		{fmt.Sprintf("mysql://root@%s:%s/%s", dbHost, dbPort, dbName), nil},
		{fmt.Sprintf("mysql://%s:%s@%s:%s/", dbUser, dbPass, dbHost, dbPort), errors.New("database absent, " + require)},
		{fmt.Sprintf("mysql://%s:%s@%s:%s", dbUser, dbPass, dbHost, dbPort), errors.New("database absent, " + require)},
		{fmt.Sprintf("mysql://%s:%s@/%s", dbUser, dbPass, dbName), errors.New("host absent, " + require)},
		{fmt.Sprintf("mysql://:%s@%s:%s/%s", dbPass, dbHost, dbPort, dbName), errors.New("username absent, " + require)},
		{fmt.Sprintf("mysql://%s:%s@%s:%s/%s?a=3", dbUser, dbPass, dbHost, dbPort, dbName), errors.New("unexpected query params or fragment, " + require)},
		{fmt.Sprintf("mysql://%s:%s@%s:%s/%s#a", dbUser, dbPass, dbHost, dbPort, dbName), errors.New("unexpected query params or fragment, " + require)},
		{fmt.Sprintf("mysql://"), errors.New("username absent, " + require)},
	}

	for _, v := range cases {
		p, err := url.Parse(v.url)
		t.Nil(err)
		c, err := connect(p)
		t.Err(err, v.wanterr)
		if v.wanterr == nil {
			c.db.Close()
		}
	}

	p, err := url.Parse(fmt.Sprintf("mysql://incUserName:%s@%s:%s/%s", dbPass, dbHost, dbPort, dbName))
	t.Nil(err)
	_, err = connect(p)
	t.Match(err, `Access denied for user 'incUserName'@.* \(using password: YES\)`)

	p, err = url.Parse(fmt.Sprintf("mysql://%s:incPass@%s:%s/%s", dbUser, dbHost, dbPort, dbName))
	t.Nil(err)
	_, err = connect(p)
	t.Match(err, `Access denied for user 'gotestuser'@.* \(using password: YES\)`)
}

func TestInitialize(tt *testing.T) {
	t := check.T(tt)
	t.Nil(initialize(locUser))
	dropTable(t)
}

func TestInitialized(tt *testing.T) {
	t := check.T(tt)

	v, err := connect(locUser)
	t.Nil(err)

	//- Not initialized()
	t.False(v.initialized())

	//- Initialized()
	t.Nil(initialize(locUser))
	t.True(v.initialized())
	dropTable(t)
}

func testLock(name string, loc *url.URL, unlockc chan struct{}, statusc chan string) {
	v, err := new(loc)
	if err != nil {
		panic(err)
	}

	cancel := make(chan struct{}, 1)
	go func() {
		select {
		case <-cancel:
		case <-time.After(100 * time.Millisecond):
			statusc <- "block " + name
		}
	}()

	switch {
	case strings.HasPrefix(name, "EX"):
		v.ExclusiveLock()
	case strings.HasPrefix(name, "SH"):
		v.SharedLock()
	default:
		panic("name must begins with EX or SH")
	}
	cancel <- struct{}{}
	statusc <- "acquired " + name

	<-unlockc
	v.Unlock()
}

// - EX1, UN1, EX2, UN2
func TestExSequence(tt *testing.T) {
	t := check.T(tt)

	t.Nil(initialize(locUser))
	defer dropTable(t)

	statusc := make(chan string)
	un1 := make(chan struct{})
	un2 := make(chan struct{})
	go testLock("EX1", locUser, un1, statusc)
	t.Equal(<-statusc, "acquired EX1")
	un1 <- struct{}{}
	go testLock("EX2", locUser, un2, statusc)
	t.Equal(<-statusc, "acquired EX2")
	un2 <- struct{}{}
}

// - EX1, EX2(block), UN1, (unblockEX2), UN2
func TestExParallel(tt *testing.T) {
	t := check.T(tt)

	t.Nil(initialize(locUser))
	defer dropTable(t)

	statusc := make(chan string)
	un1 := make(chan struct{})
	un2 := make(chan struct{})
	go testLock("EX1", locUser, un1, statusc)
	t.Equal(<-statusc, "acquired EX1")
	go testLock("EX2", locUser, un2, statusc)
	t.Equal(<-statusc, "block EX2")
	un1 <- struct{}{}
	t.Equal(<-statusc, "acquired EX2")
	un2 <- struct{}{}

}

// - EX1, SH2(block), UN1, (unblock)SH2, UN2
func TestExShParallel(tt *testing.T) {
	t := check.T(tt)

	t.Nil(initialize(locUser))
	defer dropTable(t)

	statusc := make(chan string)
	un1 := make(chan struct{})
	un2 := make(chan struct{})
	go testLock("EX1", locUser, un1, statusc)
	t.Equal(<-statusc, "acquired EX1")
	go testLock("SH2", locUser, un2, statusc)
	t.Equal(<-statusc, "block SH2")
	un1 <- struct{}{}
	t.Equal(<-statusc, "acquired SH2")
	un2 <- struct{}{}
}

// - SH1, SH2, UN1, UN2
func TestShParallel(tt *testing.T) {
	t := check.T(tt)

	t.Nil(initialize(locUser))
	defer dropTable(t)

	statusc := make(chan string)
	un1 := make(chan struct{})
	un2 := make(chan struct{})
	go testLock("SH1", locUser, un1, statusc)
	t.Equal(<-statusc, "acquired SH1")
	go testLock("SH2", locUser, un2, statusc)
	t.Equal(<-statusc, "acquired SH2")
	un1 <- struct{}{}
	un2 <- struct{}{}
}

// - SH1, EX2(block), SH3(block), UN1, (unblock)EX2, UN2, (unblock)SH3, UN3
func TestExPriority(tt *testing.T) {
	t := check.T(tt)

	t.Nil(initialize(locUser))
	defer dropTable(t)

	statusc := make(chan string)
	un1 := make(chan struct{})
	un2 := make(chan struct{})
	un3 := make(chan struct{})
	go testLock("SH1", locUser, un1, statusc)
	t.Equal(<-statusc, "acquired SH1")
	go testLock("EX2", locUser, un2, statusc)
	t.Equal(<-statusc, "block EX2")
	go testLock("SH3", locUser, un3, statusc)
	t.Equal(<-statusc, "block SH3")
	un1 <- struct{}{}
	t.Equal(<-statusc, "acquired EX2")
	un2 <- struct{}{}
	t.Equal(<-statusc, "acquired SH3")
	un3 <- struct{}{}
}

func TestGet(tt *testing.T) {
	t := check.T(tt)

	v, err := connect(locUser)
	t.Nil(err)

	// - Not initialized
	t.Panic(func() { v.Get() }, `Table 'gotest.Narada4D' dosen't exist`)

	// - Initialized
	t.Nil(initialize(locUser))
	defer dropTable(t)
	t.Equal(v.Get(), "none")
}

func TestSet(tt *testing.T) {
	t := check.T(tt)

	c, err := connect(locUser)
	t.Nil(err)

	t.Nil(initialize(locUser))
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

	for _, v := range cases {
		if v.wantpanic {
			t.PanicMatch(func() { c.Set(v.val) }, `invalid version value, require 'none' or 'dirty' or one or more digits separated with single dots`)
		} else {
			t.NotPanic(func() { c.Set(v.val) })
			t.Equal(c.Get(), v.val)
		}
	}
}

func dropTable(t *check.C) {
	t.Helper()
	v, err := connect(locUser)
	t.Nil(err)
	_, err = v.db.Exec(dropTableVersion)
	t.Nil(err)
	t.Nil(v.db.Close())
}

func dockerCleanup() {
	for _, id := range dockerIDs {
		if _, err := docker("kill", id); err != nil {
			fmt.Println(err)
		}
	}
}

func runMySQL() (port string, err error) {
	id, err := docker("run", "-d", "--rm", "-P",
		"-e", "MYSQL_ALLOW_EMPTY_PASSWORD=yes",
		"-e", "MYSQL_DATABASE="+dbName,
		"-e", "MYSQL_USER="+dbUser,
		"-e", "MYSQL_PASSWORD="+dbPass,
		"mysql")
	if err != nil {
		return "", err
	}
	id = strings.TrimSpace(id)
	dockerIDs = append(dockerIDs, id)

	port, err = dockerPort(id, "3306/tcp")
	if err != nil {
		return "", err
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=1s&readTimeout=1s&writeTimeout=1s",
		dbUser, dbPass, dbHost, port, dbName))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	stdout := os.Stdout
	os.Stdout = nil
	mysql.SetLogger(log.New(ioutil.Discard, "", 0))
	defer func() {
		os.Stdout = stdout
		mysql.SetLogger(log.New(os.Stderr, "[mysql] ", log.Ldate|log.Ltime|log.Lshortfile))
	}()

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	for {
		if err = db.PingContext(ctx); err == nil {
			return port, nil
		}
		time.Sleep(time.Second)
	}
	return "", errors.New("failed to connect to mysql")
}

func dockerPort(id, internalPort string) (string, error) {
	out, err := docker("inspect", id)
	if err != nil {
		return "", err
	}
	var inspect []struct {
		NetworkSettings struct {
			Ports map[string][]struct {
				HostPort string
			}
		}
	}
	err = json.Unmarshal([]byte(out), &inspect)
	if err != nil {
		return "", err
	}
	port := inspect[0].NetworkSettings.Ports[internalPort][0].HostPort
	if port == "" {
		return "", errors.New("failed to detect port")
	}
	return port, nil
}

func docker(args ...string) (string, error) {
	out, err := exec.Command("docker", args...).Output()
	if exitErr, ok := err.(*exec.ExitError); ok {
		fmt.Println("docker", strings.Join(args, " "))
		fmt.Println(string(exitErr.Stderr))
	}
	return string(out), err
}
