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
	dbName = "gotest"
	dbHost = "127.0.0.1"
	dbUser = "gotestuser"
	dbPass = "gotestpass"
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

// - mysql://tanya:010203@localhost:3306/PROJECT (success)
// - mysql://tanya:010203@localhost:3306/
// - mysql://tanya:010203@/PROJECT
// - mysql://tanya@localhost:3306/PROJECT(success)
// - mysql://tanya@local/PROJECT
// - mysql://010203@localhost:3306/PROJECT
// - mysql://localhost:3306/PROJECT
// - mysql://tanya:010203@localhost:3306/PROJECT/?a=3
// - mysql://tanya:010203@localhost:3306/PROJECT/#a
// - mysql://
// - test://
func TestConnect(t *testing.T) {

}

// - mysql://tanya@localhost:3306/PROJECT, TABLE created
// - mysql://tanya@localhost:3306/PROJECT, Connection drop, TABLE not created, error
func TestInitialize(tt *testing.T) {
	t := check.T(tt)
	t.Nil(initialize(locUser))
}

//- Protocol registered, 'SELECT COUNT (*) FROM Narada4D', true
//- Protocol not registered, 'SELECT COUNT (*) FROM Narada4D', false
//- Protocol registered, 'SELECT COUNT (*) FROM Narada4D', connection
//drop, false (reconnected automaticaly - true)???
func TestInitialized(t *testing.T) {

}

// - SH, SH, UN, UN
// - SH, EX(block), UN(SH), EX, UN(EX)
// - EX, SH(block), UN(EX), SH, UN(SH)
// - SH, EX(block), SH(block), UN(SH), EX, UN(EX), SH, UN(SH)
func TestShExLock(t *testing.T) {

}

// - UN, error
func TestUnlock(t *testing.T) {

}

// - Protocol registered, 'SELECT val FROM Narada4D WHERE var=`version`' (success)
// - Protocol not registered, 'SELECT val FROM Narada4D WHERE var=`version`', panic
func TestGet(t *testing.T) {

}

// - Protocol registered, sqlSetVersion, val=43, success
// - Protocol registered, sqlSetVersion, val=43.0, success
// - Protocol registered, sqlSetVersion, val=43.0.1, success
// - Protocol registered, sqlSetVersion, val="", panic
// - Protocol registered, sqlSetVersion, val=0, ?
// - Protocol registered, sqlSetVersion, val=-18, panic
// - Protocol registered, sqlSetVersion, val=rat, panic
// - Protocol not registered, sqlSetVersion, val=43, panic
func TestSet(t *testing.T) {

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
