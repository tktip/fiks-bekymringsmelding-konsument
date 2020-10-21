// +build windows

package hooks

import (
	"github.com/kardianos/service"
	"github.com/sirupsen/logrus"
)

// eventHook implements the logrus.Hook interface
type eventHook struct {
	logger service.Logger
}

//Fire - trigger new entry event
func (e *eventHook) Fire(entry *logrus.Entry) error {
	msg, err := entry.String()
	if err != nil {
		return err
	}

	switch entry.Level {
	case logrus.WarnLevel:
		e.logger.Warning(msg)
	case logrus.FatalLevel:
		e.logger.Error(msg)
	default:
		e.logger.Info(msg)
	}
	return nil
}

// revive:disable:unused-receiver

// Levels - log to all levels
func (e *eventHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// AddEventLogHook - adds an event hook to the logger for given service.
// Creates log entries for service in windows event log.
func AddEventLogHook(log *logrus.Logger, service service.Service) error {
	logger, err := service.SystemLogger(nil)
	if err != nil {
		return err
	}
	log.AddHook(&eventHook{logger})
	return nil
}
