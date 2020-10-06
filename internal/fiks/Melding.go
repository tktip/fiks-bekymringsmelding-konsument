package fiks

import (
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/tktip/fiks-bekymringsmelding-konsument/internal/log"
	"strings"
)

// fiksMessage - helper struct to give easy access to
// message data.
type fiksMessage struct {
	*amqp.Delivery
}

// GetMeldingID - returns the 'melding-id' header set by FIKS.
func (f fiksMessage) GetMeldingID() string {
	return f.Headers[headerMeldingID].(string)
}

// GetAvsenderID - returns the 'avsender-id' header set by FIKS.
func (f fiksMessage) GetAvsenderID() string {
	return f.Headers[headerAvsenderID].(string)
}

// GetSvarTil - returns the 'svar-til' header set by FIKS.
func (f fiksMessage) GetSvarTil() string {
	return f.Headers[headerSvarTil].(string)
}

// GetClientID - returns the TIP queue client ID.
func (f fiksMessage) GetClientID() string {
	return strings.Split(f.RoutingKey, ".")[2]
}

// LoggerWithFields - return a logrus entry with logging
// fields relevant to message.
func (f fiksMessage) LoggerWithFields() *logrus.Entry {
	return log.Logger.
		WithField("msg-id", f.GetMeldingID())
}
