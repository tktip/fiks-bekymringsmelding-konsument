package fiks

import (
	"github.com/tktip/fiks-bekymringsmelding-konsument/internal/crypt"
	"github.com/tktip/maskinporten/pkg/maskinporten"
)

const (
	headerAvsenderID = "avsender-id"
	headerMeldingID  = "melding-id"
	headerSvarTil    = "svar-til"

	headerDokumentlagerID = "dokumentlager-id"
	headerAvsenderNavn    = "avsender-navn"
	headerMessageType     = "type"
)

// MessageType - message type specification is expected by FIKS IO. Helper struct
// to promote use of correct types.
type MessageType string

const (
	//MessageTypeSucceeded - processing of message succeeded
	MessageTypeSucceeded = MessageType("no.ks.fiks.bekymringsmelding.mottatt.v1")

	//MessageTypeFailed - processing of message failed (any reason)
	MessageTypeFailed = MessageType("no.ks.fiks.bekymringsmelding.feilet.v1")
)

// Handler - provides functionality relevant to document retrieval and storage.
// Use the Run method to start automatic retrieval.
type Handler struct {
	LogLevel     string         `yaml:"logLevel"`
	Crypto       crypt.Handler  `yaml:"decryptor"`
	RabbitMQ     RabbitMQConfig `yaml:"rabbitmq"`
	Sender       Sender         `yaml:"sender"`
	FileHandler  FileHandler    `yaml:"fileHandler"`
	Maskinporten maskinporten.Handler
}
