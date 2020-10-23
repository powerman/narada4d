// Package file registers schemaver.Backend implemented using lock-files.
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
	versionFileName   = ".version"
	lockFileName      = ".lock"
	lockQueueFileName = ".lock.queue"
)

type storage struct {
	versionPath   string
	lockPath      string
	lockQueuePath string
	lockFile      *os.File
	lockQueueFile *os.File
	lockFD        int
	lockQueueFD   int
}

func init() {
	schemaver.RegisterProtocol("file", schemaver.Backend{
		Initialize: initialize,
		New:        newInitializedStorage,
	})
}

func initialize(loc *url.URL) error {
	s, err := newStorage(loc)
	if err != nil {
		return err
	}

	if s.initialized() {
		return fmt.Errorf("version already initialized at %q", s.versionPath)
	}
	return s.init()
}

func newInitializedStorage(loc *url.URL) (schemaver.Manage, error) {
	s, err := newStorage(loc)
	if err != nil {
		return nil, err
	}
	if !s.initialized() {
		if err := s.init(); err != nil {
			return nil, err
		}
	}
	err = s.open()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func newStorage(loc *url.URL) (*storage, error) {
	if loc.User != nil || loc.Host != "" || loc.RawQuery != "" || loc.Fragment != "" {
		return nil, errors.New("location must contain only path")
	}

	dir := filepath.Clean(loc.Path)
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		return nil, errors.New("location path must be existing directory")
	}

	s := &storage{
		versionPath:   filepath.Join(loc.Path, versionFileName),
		lockPath:      filepath.Join(loc.Path, lockFileName),
		lockQueuePath: filepath.Join(loc.Path, lockQueueFileName),
	}
	return s, nil
}

func (s *storage) initialized() bool {
	fi, err := os.Lstat(s.versionPath)
	return err == nil && fi.Mode()&os.ModeSymlink != 0
}

func (s *storage) init() error {
	err := ioutil.WriteFile(s.lockPath, nil, 0444)
	if err == nil {
		err = ioutil.WriteFile(s.lockQueuePath, nil, 0444)
	}
	if err == nil {
		err = os.Symlink(schemaver.NoVersion, s.versionPath)
	}
	return err
}

func (s *storage) open() (err error) {
	s.lockFile, err = os.Open(s.lockPath)
	if err != nil {
		return err
	}
	s.lockQueueFile, err = os.Open(s.lockQueuePath)
	if err != nil {
		return err
	}
	s.lockFD = int(s.lockFile.Fd())
	s.lockQueueFD = int(s.lockQueueFile.Fd())
	return nil
}

func (s *storage) SharedLock() {
	s.lock(syscall.LOCK_SH)
}

func (s *storage) ExclusiveLock() {
	s.lock(syscall.LOCK_EX)
}

func (s *storage) lock(how int) {
	if err := syscall.Flock(s.lockQueueFD, syscall.LOCK_EX); err != nil {
		panic(err)
	}
	if err := syscall.Flock(s.lockFD, how); err != nil {
		panic(err)
	}
	if err := syscall.Flock(s.lockQueueFD, syscall.LOCK_UN); err != nil {
		panic(err)
	}
}

func (s *storage) Unlock() {
	if err := syscall.Flock(s.lockFD, syscall.LOCK_UN); err != nil {
		panic(err)
	}
}

func (s *storage) Get() string {
	ver, err := os.Readlink(s.versionPath)
	if err != nil {
		panic(err)
	}
	return ver
}

func (s *storage) Set(ver string) {
	tmpPath := s.versionPath + ".tmp"
	_ = os.Remove(tmpPath)
	if err := os.Symlink(ver, tmpPath); err != nil {
		panic(err)
	}
	if err := os.Rename(tmpPath, s.versionPath); err != nil {
		panic(err)
	}
}

func (s *storage) Close() error {
	err := s.lockFile.Close()
	if err == nil {
		err = s.lockQueueFile.Close()
	}
	return err
}
