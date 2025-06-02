package updater

import (
	"errors"
	"os"
)

type autoDeleteReadCloser struct {
	*os.File
	path string
}

func (rc *autoDeleteReadCloser) Close() error {
	errClose := rc.File.Close()
	errRemove := os.Remove(rc.path)
	return errors.Join(errClose, errRemove)
}
