//go:build !windows

// Package file registers schemaver.Backend implemented using lock-files.
//
// It is not available on Windows.
package file

import (
	"errors"
	"fmt"
	urlpkg "net/url"
	"os"
	"path/filepath"
	"syscall"

	"github.com/powerman/narada4d/schemaver"
)

const (
	versionFileName   = ".version"
	lockFileName      = ".lock"
	lockQueueFileName = ".lock.queue"
	lockFilePerm      = 0o400
)

var (
	errVersionAlreadyInitialized = errors.New("version is already initialized")
	errLocationInvalid           = errors.New("location must contain only path")
	errLocationWrongPath         = errors.New("location path must be existing directory")
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

func init() { //nolint:gochecknoinits // Registration pattern.
	schemaver.RegisterProtocol("file", schemaver.Backend{
		Initialize: initialize,
		New:        newInitializedStorage,
	})
}

func initialize(loc *urlpkg.URL) error {
	s, err := newStorage(loc)
	if err != nil {
		return err
	}

	if s.initialized() {
		return fmt.Errorf("%w at %q", errVersionAlreadyInitialized, s.versionPath)
	}
	return s.init()
}

func newInitializedStorage(loc *urlpkg.URL) (schemaver.Manage, error) {
	s, err := newStorage(loc)
	if err != nil {
		return nil, err
	}
	if !s.initialized() {
		err := s.init()
		if err != nil {
			return nil, err
		}
	}
	err = s.open()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func newStorage(loc *urlpkg.URL) (*storage, error) {
	if loc.User != nil || loc.Host != "" || loc.RawQuery != "" || loc.Fragment != "" {
		return nil, errLocationInvalid
	}

	dir := filepath.Clean(loc.Path)
	fi, err := os.Stat(dir)
	if err != nil || !fi.IsDir() {
		return nil, errLocationWrongPath
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
	err := os.WriteFile(s.lockPath, nil, lockFilePerm)
	if err == nil {
		err = os.WriteFile(s.lockQueuePath, nil, lockFilePerm)
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
	err := syscall.Flock(s.lockQueueFD, syscall.LOCK_EX)
	if err != nil {
		panic(err)
	}
	err = syscall.Flock(s.lockFD, how)
	if err != nil {
		panic(err)
	}
	err = syscall.Flock(s.lockQueueFD, syscall.LOCK_UN)
	if err != nil {
		panic(err)
	}
}

func (s *storage) Unlock() {
	err := syscall.Flock(s.lockFD, syscall.LOCK_UN)
	if err != nil {
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
	err := os.Symlink(ver, tmpPath)
	if err != nil {
		panic(err)
	}
	err = os.Rename(tmpPath, s.versionPath)
	if err != nil {
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
