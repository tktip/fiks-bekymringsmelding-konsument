package log

import (
	"github.com/sirupsen/logrus"
	nested "github.com/antonfisher/nested-logrus-formatter"
)

// Logger singleton logger for entire project. Present to simplify
// different logging on windows and linux
var Logger *logrus.Logger

func init() {
	Logger = logrus.New()

	//Formatter to get fields first for easier readability.
	Logger.SetFormatter(&nested.Formatter{FieldsOrder: []string{"msg-id"}})
}
