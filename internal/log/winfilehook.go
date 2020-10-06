// +build windows

package log

import (
	"github.com/sirupsen/logrus"
	"os"
)

// FileHook to send logs via windows log.
type FileHook struct {
	path string
	file *os.File
}

// AddFileHook creates and returns a new FileHook wrapped around anything
// that implements the debug.Log interface
func AddFileHook(path string) error {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	Logger.AddHook(&FileHook{path: path, file: file})
	return nil
}

//Fire - trigger new entry event
func (hook *FileHook) Fire(entry *logrus.Entry) error {
	msg, err := entry.String()
	if err != nil {
		return err
	}

	_, err = hook.file.WriteString(msg)
	return err
}

// revive:disable:unused-receiver

// Levels - log to all levels
func (hook *FileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
