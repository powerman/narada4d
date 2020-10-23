// Package schemaver provides a way to manage your data schema version.
package schemaver

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	// NoVersion will be set as version value while initialization.
	NoVersion = "none"
	// BadVersion must be set as version value in case data schema may
	// be incorrect (as result of interrupted migration or restoring
	// backup).
	BadVersion = "dirty"
	// EnvLocation must contain url pointing to location of data
	// schema version. For example: file:///some/path/.
	EnvLocation = "NARADA4D"
	// EnvSkipLock must be set to non-empty value in case
	// ExclusiveLock was already acquired (by parent process).
	EnvSkipLock = "NARADA4D_SKIP_LOCK"
)

var errUnknownProtocol = errors.New("narada4d: unknown protocol")

func parseLocation(location string) (*url.URL, error) {
	loc, err := url.Parse(location)
	if err != nil {
		return nil, fmt.Errorf("narada4d: %w", err)
	}
	if registered[loc.Scheme] == nil {
		return nil, fmt.Errorf("%w: %q", errUnknownProtocol, loc.Scheme)
	}
	return loc, nil
}

// Initialize initialize version at location provided in $NARADA4D.
//
// Version must not be already initialized.
func Initialize() error {
	loc, err := parseLocation(os.Getenv(EnvLocation))
	if err != nil {
		return err
	}
	return registered[loc.Scheme].Initialize(loc)
}

type lockType int

const (
	unlocked lockType = iota
	shared
	exclusive
)

// SchemaVer manage data schema versions.
type SchemaVer struct {
	loc        *url.URL
	backend    Manage
	mu         sync.Mutex
	lockType   lockType
	skipUnlock int
	sharedVer  string
	callbacks  []func(string)
	holdWG     sync.WaitGroup
	holdQuit   chan struct{}
}

// New creates object for managing data schema version at location
// provided in $NARADA4D.
//
// Will initialize version if it's not initialized yet.
func New() (*SchemaVer, error) {
	return NewAt(os.Getenv(EnvLocation))
}

// NewAt creates object for managing data schema version at location.
//
// Will initialize version if it's not initialized yet.
func NewAt(location string) (*SchemaVer, error) {
	loc, err := parseLocation(location)
	if err != nil {
		return nil, err
	}
	backend, err := registered[loc.Scheme].New(loc)
	if err != nil {
		return nil, err
	}

	v := &SchemaVer{
		loc:      loc,
		backend:  backend,
		holdQuit: make(chan struct{}),
	}
	if v.isSkipLock() {
		v.lockType = exclusive
	}

	return v, nil
}

//nolint:gochecknoglobals // By design.
var muEnv sync.Mutex

func (v *SchemaVer) isSkipLock() bool {
	muEnv.Lock()
	defer muEnv.Unlock()

	envs := strings.Fields(os.Getenv(EnvSkipLock))
	if len(envs) == 0 {
		return false
	}
	if _, err := parseLocation(envs[0]); err != nil {
		return true // For compatibility with NARADA4D_SKIP_LOCK=1 used before.
	}
	for i := range envs {
		if loc, err := parseLocation(envs[i]); err == nil && loc.String() == v.loc.String() {
			return true
		}
	}
	return false
}

func (v *SchemaVer) setSkipLock() {
	muEnv.Lock()
	defer muEnv.Unlock()

	envs := strings.Fields(os.Getenv(EnvSkipLock))
	for i := range envs {
		if loc, err := parseLocation(envs[i]); err == nil && loc.String() == v.loc.String() {
			return
		}
	}

	envs = append(envs, v.loc.String())
	err := os.Setenv(EnvSkipLock, strings.Join(envs, " "))
	if err != nil {
		panic(err)
	}
}

func (v *SchemaVer) unsetSkipLock() {
	muEnv.Lock()
	defer muEnv.Unlock()

	envs := strings.Fields(os.Getenv(EnvSkipLock))
	for i := len(envs) - 1; i >= 0; i-- {
		if loc, err := parseLocation(envs[i]); err == nil && loc.String() == v.loc.String() {
			copy(envs[i:], envs[i+1:])
			envs = envs[:len(envs)-1]
		}
	}

	var err error
	if len(envs) > 0 {
		err = os.Setenv(EnvSkipLock, strings.Join(envs, " "))
	} else {
		err = os.Unsetenv(EnvSkipLock)
	}
	if err != nil {
		panic(err)
	}
}

// HoldSharedLock will start goroutine which will acquire SharedLock and
// keep it until Close or ctx.Done. It'll release and immediately
// re-acquire SharedLock every relockEvery to give someone else a chance
// to get ExclusiveLock.
//
// This is recommended optimization in case you've to do a lot of
// short-living SharedLock every second.
func (v *SchemaVer) HoldSharedLock(ctx context.Context, relockEvery time.Duration) {
	v.holdWG.Add(1)
	go func() {
		hold := true
		select {
		case <-v.holdQuit:
			hold = false
		default:
		}
		for hold {
			v.SharedLock()
			select {
			case <-time.After(relockEvery):
			case <-ctx.Done():
				hold = false
			case <-v.holdQuit:
				hold = false
			}
			v.Unlock()
		}
		v.holdWG.Done()
	}()
}

// SharedLock acquire shared lock and return current version.
//
// It may be called recursively, under already acquired SharedLock
// or ExclusiveLock (in this case it'll do nothing).
func (v *SchemaVer) SharedLock() string {
	v.mu.Lock()
	defer v.mu.Unlock()

	var ver string
	switch v.lockType {
	case exclusive:
		v.skipUnlock++
		ver = v.backend.Get()
	case shared:
		v.skipUnlock++
		ver = v.sharedVer
	case unlocked:
		v.backend.SharedLock()
		v.lockType = shared
		v.sharedVer = v.backend.Get()
		ver = v.sharedVer
	default:
		panic("never here")
	}

	for _, callback := range v.callbacks {
		callback(ver)
	}
	return ver
}

// ExclusiveLock acquire exclusive lock and return current version.
//
// It may be called recursively, under already acquired ExclusiveLock
// (in this case it'll do nothing).
func (v *SchemaVer) ExclusiveLock() string {
	v.mu.Lock()
	defer v.mu.Unlock()

	switch v.lockType {
	case exclusive:
		v.skipUnlock++
	case shared:
		panic("unable to acquire exclusive lock under shared lock")
	case unlocked:
		v.backend.ExclusiveLock()
		v.lockType = exclusive
		v.setSkipLock()
	default:
		panic("never here")
	}

	ver := v.backend.Get()
	for _, callback := range v.callbacks {
		callback(ver)
	}
	return ver
}

// Unlock release lock acquired using SharedLock or ExclusiveLock.
//
// When called to unlock previous SharedLock or ExclusiveLock
// which did nothing (because lock was already acquired) then it
// will do nothing too.
func (v *SchemaVer) Unlock() {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.lockType == unlocked {
		panic("can't unlock, no lock acquired")
	}
	if v.skipUnlock > 0 {
		v.skipUnlock--
	} else {
		v.backend.Unlock()
		v.lockType = unlocked
		v.unsetSkipLock()
	}
}

// Get returns current version.
//
// It must be called under SharedLock or ExclusiveLock.
func (v *SchemaVer) Get() string {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.lockType == unlocked {
		panic("require SharedLock or ExclusiveLock")
	}
	return v.backend.Get()
}

// Set change current version.
//
// It must be called under ExclusiveLock.
func (v *SchemaVer) Set(ver string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.lockType != exclusive {
		panic("require ExclusiveLock")
	}
	v.backend.Set(ver)
}

// AddCallback registers user-provided function which will be
// called with current version in parameter by each SharedLock or
// ExclusiveLock before they returns.
//
// Usually this function should check is current version is
// supported by this application and call log.Fatal if not.
func (v *SchemaVer) AddCallback(callback func(string)) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if callback == nil {
		panic("require callback")
	}

	v.callbacks = append(v.callbacks, callback)
}

// Close release any resources used to manage schema version.
//
// No other methods should be called after Close.
func (v *SchemaVer) Close() error {
	close(v.holdQuit)
	v.holdWG.Wait()

	v.mu.Lock()
	defer v.mu.Unlock()
	return v.backend.Close()
}
