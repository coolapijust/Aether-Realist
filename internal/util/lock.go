package util

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileLock represents a lock on a file.
type FileLock struct {
	path string
	file *os.File
}

// AcquireLock tries to acquire a lock on the specified path.
// If the lock cannot be acquired (e.g., another instance is running), it returns an error.
func AcquireLock(name string) (*FileLock, error) {
	// Use temp directory for the lock file
	path := filepath.Join(os.TempDir(), fmt.Sprintf("%s.lock", name))
	
	// Open with O_CREATE and O_EXCL to ensure only one instance creates/owns it
	// On Windows, O_EXCL ensures that if the file exists, OpenFile fails.
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf("another instance of %s is already running (lock file: %s)", name, path)
		}
		return nil, fmt.Errorf("failed to create lock file: %w", err)
	}

	return &FileLock{
		path: path,
		file: file,
	}, nil
}

// Release removes the lock file and closes the file handle.
func (l *FileLock) Release() {
	if l.file != nil {
		l.file.Close()
		os.Remove(l.path)
		l.file = nil
	}
}
