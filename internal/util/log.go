package util

import (
	"log"
	"os"
	"path/filepath"
)

var debugLogger *log.Logger

func InitLogger() (*os.File, error) {
	dir := os.TempDir()
	path := filepath.Join(dir, "lazycode-debug.log")

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	debugLogger = log.New(f, "", log.LstdFlags|log.Lshortfile)
	return f, nil
}

func Debug(format string, args ...any) {
	if debugLogger != nil {
		debugLogger.Printf(format, args...)
	}
}
