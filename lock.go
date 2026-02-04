package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// FileLock represents a file-based lock
type FileLock struct {
	lockFile string
	timeout  time.Duration
}

// NewFileLock creates a new file lock
func NewFileLock(lockFile string, timeout time.Duration) *FileLock {
	return &FileLock{
		lockFile: lockFile,
		timeout:  timeout,
	}
}

// Acquire attempts to acquire the lock with timeout
func (fl *FileLock) Acquire() error {
	start := time.Now()
	for {
		// Ensure lock directory exists
		if err := os.MkdirAll(filepath.Dir(fl.lockFile), 0755); err != nil {
			return fmt.Errorf("failed to create lock directory: %w", err)
		}

		// Try to create the lock file exclusively
		file, err := os.OpenFile(fl.lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err == nil {
			// Successfully acquired lock, write our PID
			pid := os.Getpid()
			if _, writeErr := fmt.Fprintf(file, "%d\n", pid); writeErr != nil {
				file.Close()
				os.Remove(fl.lockFile)
				return fmt.Errorf("failed to write PID to lock file: %w", writeErr)
			}
			file.Close()
			return nil
		}

		// Check if we've exceeded timeout
		if time.Since(start) > fl.timeout {
			return fmt.Errorf("failed to acquire lock after %v", fl.timeout)
		}

		// Wait a bit before trying again
		time.Sleep(100 * time.Millisecond)
	}
}

// Release releases the lock
func (fl *FileLock) Release() error {
	// Check if we own the lock
	if file, err := os.Open(fl.lockFile); err == nil {
		defer file.Close()

		// Read the PID from the lock file
		var pid int
		if _, err := fmt.Fscanf(file, "%d", &pid); err == nil {
			// Check if we own this lock
			if pid == os.Getpid() {
				return os.Remove(fl.lockFile)
			}
		}
	}

	// Lock file doesn't exist or we don't own it
	return nil
}

// IsLocked checks if the lock file exists
func (fl *FileLock) IsLocked() bool {
	_, err := os.Stat(fl.lockFile)
	return err == nil
}

// ForceRelease forcibly removes the lock file (use with caution)
func (fl *FileLock) ForceRelease() error {
	return os.Remove(fl.lockFile)
}

// AcquireLock is a convenience function that acquires a lock with default timeout
func AcquireLock(lockFile string) (*FileLock, error) {
	lock := NewFileLock(lockFile, 10*time.Second)
	if err := lock.Acquire(); err != nil {
		return nil, err
	}
	return lock, nil
}

// ReleaseLock is a convenience function that releases a lock
func ReleaseLock(lock *FileLock) error {
	if lock != nil {
		return lock.Release()
	}
	return nil
}

// CleanupStaleLocks attempts to clean up stale lock files by checking if the PID is still running
func CleanupStaleLocks(lockFile string) error {
	file, err := os.Open(lockFile)
	if err != nil {
		// Lock file doesn't exist
		return nil
	}
	defer file.Close()

	var pid int
	if _, err := fmt.Fscanf(file, "%d", &pid); err != nil {
		// Can't read PID, remove the lock file
		return os.Remove(lockFile)
	}

	// Check if process is still running
	process, err := os.FindProcess(pid)
	if err != nil {
		// Process doesn't exist, remove lock
		return os.Remove(lockFile)
	}

	// On Unix systems, we can try to send signal 0 to check if process exists
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		// Process doesn't exist or we don't have permission, remove lock
		return os.Remove(lockFile)
	}

	// Process is still running, lock is valid
	return nil
}
