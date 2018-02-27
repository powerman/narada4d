package schemaver

import (
	"fmt"
	"net/url"
	"os"
)

const (
	NoVersion   = "none"
	BadVersion  = "dirty"
	EnvLocation = "NARADA4D"
	EnvSkipLock = "NARADA4D_SKIP_LOCK"
)

// SchemaVer provides an interface for managing data schema versions.
type SchemaVer interface {
	// SharedLock acquire shared lock and return current version.
	SharedLock() string
	// ExclusiveLock acquire exclusive lock and return current version.
	ExclusiveLock() string
	// Unlock release lock acquired using SharedLock or ExclusiveLock.
	Unlock()
	// Set change current version. It must be called under
	// ExclusiveLock.
	Set(string)
}

type InitSchemaVer func(*url.URL) error

// NewSchemaVer return implementation of SchemaVer working with version at
// given location or error if location is incorrect.
type NewSchemaVer func(*url.URL) (SchemaVer, error)

var (
	initSchemaVer = make(map[string]InitSchemaVer)
	newSchemaVer  = make(map[string]NewSchemaVer)
)

// RegisterProtocol must be called by packages which implement some
// protocol before first call to InitSchemaVer or NewSchemaVer.
func RegisterProtocol(proto string, init InitSchemaVer, new NewSchemaVer) {
	if initSchemaVer[proto] != nil {
		panic(fmt.Sprintf("protocol %q already registered", proto))
	} else if init == nil || new == nil {
		panic(fmt.Sprintf("can't register protocol %q with nil implementation", proto))
	}

	initSchemaVer[proto] = init
	newSchemaVer[proto] = new
}

// Initialize initialize version at location provided in $NARADA4D.
//
// Version must not be already initialized.
func Initialize() error {
	loc, err := location()
	if err != nil {
		return err
	}
	return initSchemaVer[loc.Scheme](loc)
}

// New creates object for managing data schema version at location
// provided in $NARADA4D.
//
// Version must be already initialized.
func New() (SchemaVer, error) {
	loc, err := location()
	if err != nil {
		return nil, err
	}
	return newSchemaVer[loc.Scheme](loc)
}

func location() (*url.URL, error) {
	loc, err := url.Parse(os.Getenv(EnvLocation))
	if err != nil {
		return nil, fmt.Errorf("failed to parse $%s: %v", EnvLocation, err)
	}
	if initSchemaVer[loc.Scheme] == nil {
		return nil, fmt.Errorf("unknown protocol in $%s: %v", EnvLocation, loc.Scheme)
	}
	return loc, nil
}
