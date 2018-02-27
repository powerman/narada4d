package file

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sync"
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
	schemaver.RegisterProtocol(proto, initialize, new)
}

type lockType int

const (
	unlocked lockType = iota
	shared
	exclusive
)

type schemaVer struct {
	versionPath   string
	lockPath      string
	lockQueuePath string
	lockFile      *os.File
	lockQueueFile *os.File
	lockFD        int
	lockQueueFD   int
	mu            sync.Mutex
	lockType      lockType
	skipUnlock    int
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

func new(loc *url.URL) (schemaver.SchemaVer, error) {
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

	if os.Getenv(schemaver.EnvSkipLock) != "" {
		v.lockType = exclusive
	}
	return v, nil
}

func (v *schemaVer) initialized() bool {
	fi, err := os.Stat(v.versionPath)
	return err == nil && fi.Mode()&os.ModeSymlink != 0
}

// SharedLock acquire shared lock and return current version.
func (v *schemaVer) SharedLock() string {
	return v.lock(shared)
}

// ExclusiveLock acquire exclusive lock and return current version.
func (v *schemaVer) ExclusiveLock() string {
	return v.lock(exclusive)
}

func (v *schemaVer) lock(typ lockType) string {
	v.mu.Lock()
	defer v.mu.Unlock()

	switch v.lockType {
	case exclusive:
		v.skipUnlock++
	case shared:
		if typ == exclusive {
			panic("unable to acquire exclusive lock under shared lock")
		}
		v.skipUnlock++
	default:
		if err := syscall.Flock(v.lockQueueFD, syscall.LOCK_EX); err != nil {
			panic(err)
		}
		how := syscall.LOCK_EX
		if typ == shared {
			how = syscall.LOCK_SH
		}
		if err := syscall.Flock(v.lockFD, how); err != nil {
			panic(err)
		}
		if err := syscall.Flock(v.lockQueueFD, syscall.LOCK_UN); err != nil {
			panic(err)
		}
		if typ == exclusive {
			if err := os.Setenv(schemaver.EnvSkipLock, "1"); err != nil {
				panic(err)
			}
		}
		v.lockType = typ
	}

	return v.get()
}

// Unlock release lock acquired using SharedLock or ExclusiveLock.
func (v *schemaVer) Unlock() {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.skipUnlock > 0 {
		v.skipUnlock--
		return
	}

	if err := syscall.Flock(v.lockFD, syscall.LOCK_UN); err != nil {
		panic(err)
	}
	if err := os.Unsetenv(schemaver.EnvSkipLock); err != nil {
		panic(err)
	}
	v.lockType = unlocked
}

// Set change current version. It must be called under ExclusiveLock.
func (v *schemaVer) Set(ver string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.lockType != exclusive {
		panic("exclusive lock required")
	}

	if err := os.Symlink(ver, v.lockPath); err != nil {
		panic(err)
	}
}

func (v *schemaVer) get() string {
	ver, err := os.Readlink(v.versionPath)
	if err != nil {
		panic(err)
	}
	return ver
}
