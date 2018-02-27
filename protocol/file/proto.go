package file

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"syscall"

	"github.com/powerman/narada4d/schemaver"
)

const (
	proto             = "file"
	versionFileName   = ".version"
	lockFileName      = ".lock"
	lockQueueFileName = ".lock.queue"
)

func init() {
	schemaver.RegisterProtocol(proto, schemaver.Backend{
		Initialize: initialize,
		New:        new,
	})
}

type schemaVer struct {
	versionPath   string
	lockPath      string
	lockQueuePath string
	lockFile      *os.File
	lockQueueFile *os.File
	lockFD        int
	lockQueueFD   int
}

func parse(loc *url.URL) (*schemaVer, error) {
	if loc.User != nil || loc.Host != "" || loc.RawQuery != "" || loc.Fragment != "" {
		return nil, errors.New("location must contain only path")
	}

	dir := filepath.Clean(loc.Path)
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		return nil, errors.New("location path must be existing directory")
	}

	return &schemaVer{
		versionPath:   filepath.Join(loc.Path, versionFileName),
		lockPath:      filepath.Join(loc.Path, lockFileName),
		lockQueuePath: filepath.Join(loc.Path, lockQueueFileName),
	}, nil
}

func initialize(loc *url.URL) error {
	v, err := parse(loc)
	if err == nil && v.initialized() {
		err = fmt.Errorf("version already initialized at %v", loc)
	}
	if err == nil {
		err = ioutil.WriteFile(v.lockPath, nil, 0444)
	}
	if err == nil {
		err = ioutil.WriteFile(v.lockQueuePath, nil, 0444)
	}
	if err == nil {
		err = os.Symlink(schemaver.NoVersion, v.versionPath)
	}
	return err
}

func new(loc *url.URL) (schemaver.Manage, error) {
	v, err := parse(loc)
	if err != nil {
		return nil, err
	}
	if !v.initialized() {
		return nil, fmt.Errorf("version is not initialized at %v", loc)
	}

	v.lockFile, err = os.Open(v.lockPath)
	if err != nil {
		return nil, err
	}
	v.lockQueueFile, err = os.Open(v.lockQueuePath)
	if err != nil {
		return nil, err
	}
	v.lockFD = int(v.lockFile.Fd())
	v.lockQueueFD = int(v.lockQueueFile.Fd())

	return v, nil
}

func (v *schemaVer) initialized() bool {
	fi, err := os.Lstat(v.versionPath)
	return err == nil && fi.Mode()&os.ModeSymlink != 0
}

// SharedLock implements schemaver.Backend interface.
func (v *schemaVer) SharedLock() {
	v.lock(syscall.LOCK_SH)
}

// ExclusiveLock implements schemaver.Backend interface.
func (v *schemaVer) ExclusiveLock() {
	v.lock(syscall.LOCK_EX)
}

func (v *schemaVer) lock(how int) {
	if err := syscall.Flock(v.lockQueueFD, syscall.LOCK_EX); err != nil {
		panic(err)
	}
	if err := syscall.Flock(v.lockFD, how); err != nil {
		panic(err)
	}
	if err := syscall.Flock(v.lockQueueFD, syscall.LOCK_UN); err != nil {
		panic(err)
	}
}

// Unlock implements schemaver.Backend interface.
func (v *schemaVer) Unlock() {
	if err := syscall.Flock(v.lockFD, syscall.LOCK_UN); err != nil {
		panic(err)
	}
}

// Get implements schemaver.Backend interface.
func (v *schemaVer) Get() string {
	ver, err := os.Readlink(v.versionPath)
	if err != nil {
		panic(err)
	}
	return ver
}

// Set implements schemaver.Backend interface.
func (v *schemaVer) Set(ver string) {
	if err := os.Symlink(ver, v.lockPath); err != nil {
		panic(err)
	}
}
