package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type customLogWriter struct {
	file        *os.File
	logDir      string
	logFile     string
	maxSize     int64
	currentSize int64
}

func newCustomLogWriter(logFilePath string, maxSizeMB int) (*customLogWriter, error) {
	logDir := filepath.Dir(logFilePath)
	logFile := filepath.Base(logFilePath)
	maxSize := int64(maxSizeMB * 1024 * 1024)

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return &customLogWriter{
		file:        file,
		logDir:      logDir,
		logFile:     logFile,
		maxSize:     maxSize,
		currentSize: stat.Size(),
	}, nil
}

func (w *customLogWriter) Write(p []byte) (n int, err error) {
	if w.currentSize+int64(len(p)) > w.maxSize {
		if err := w.rotateLogFile(); err != nil {
			return 0, err
		}
	}
	n, err = w.file.Write(p)
	w.currentSize += int64(n)
	return n, err
}

func (w *customLogWriter) rotateLogFile() error {
	if err := w.file.Close(); err != nil {
		return err
	}

	timestamp := time.Now().Format("20060102_150405")
	backupLogFile := filepath.Join(w.logDir, w.logFile+"."+timestamp)
	if err := os.Rename(filepath.Join(w.logDir, w.logFile), backupLogFile); err != nil {
		return err
	}

	newFile, err := os.OpenFile(filepath.Join(w.logDir, w.logFile), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	w.file = newFile
	w.currentSize = 0
	return nil
}

func initLogging(config *Config) error {
	if err := os.MkdirAll(filepath.Dir(config.LogFilePath), 0777); err != nil {
		return err
	}
	writer, err := newCustomLogWriter(config.LogFilePath, config.LogMaxSizeMB)
	if err != nil {
		return err
	}

	log.SetOutput(io.MultiWriter(os.Stdout, writer))
	return nil
}
