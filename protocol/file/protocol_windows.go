//go:build windows

package file

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/powerman/narada4d/schemaver"
)

var errWindowsNotSupported = errors.New(
	`protocol "file" is not supported on Windows`,
)

func init() {
	schemaver.RegisterProtocol("file", schemaver.Backend{
		Initialize: initialize,
		New:        newInitializedStorage,
	})
}

type storage struct{}

func initialize(loc *url.URL) error {
	return fmt.Errorf("%w", errWindowsNotSupported)
}

func newInitializedStorage(loc *url.URL) (schemaver.Manage, error) {
	return nil, fmt.Errorf("%w", errWindowsNotSupported)
}
