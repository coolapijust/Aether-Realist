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
func AcquireLock(name string) (*FileLock, error) {
	path := filepath.Join(os.TempDir(), fmt.Sprintf("%s.lock", name))
	
	// Check if lock file exists
	if _, err := os.Stat(path); err == nil {
		// File exists, check if process is alive
		data, err := os.ReadFile(path)
		if err == nil {
			var pid int
			if _, err := fmt.Sscanf(string(data), "%d", &pid); err == nil {
				// Check if process is alive
				proc, err := os.FindProcess(pid)
				if err == nil {
					// On Windows, FindProcess always succeeds, we must call Signal(0) to check
					if err := proc.Signal(syscall.Signal(0)); err == nil {
						return nil, fmt.Errorf("another instance of %s (PID %d) is already running", name, pid)
					}
				}
			}
		}
		// Stale lock file, remove it
		_ = os.Remove(path)
	}

	// Create new lock file
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to create lock file: %w", err)
	}

	// Write PID to lock file
	fmt.Fprintf(file, "%d", os.Getpid())

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
