package fiks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/tktip/fiks-bekymringsmelding-konsument/internal/log"
	"io"
	"net/url"
)

// RabbitMQConfig - struct for for rabbitmq
type RabbitMQConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Queue    string `yaml:"queue"`
	Consumer string `yaml:"consumer"`
}

// Connects to FIKSIO rabbitmq server
func (h *Handler) getAMQPClient() (*amqp.Connection, error) {

	//generate access token
	accesstoken, err := h.Maskinporten.CreateAccessToken()
	if err != nil {
		return nil, err
	}

	//create connection url
	combinedString := h.RabbitMQ.Password + " " + accesstoken.AccessToken
	t := &url.URL{Path: combinedString}
	encodedPW := t.String()

	url := fmt.Sprintf(
		"amqps://%s:%s@%s:%s/",
		h.RabbitMQ.Username,
		encodedPW,
		h.RabbitMQ.Host,
		h.RabbitMQ.Port,
	)

	log.Logger.Infof("Trying to connect to %s", url)

	//try to connect, 30s default timeout
	return amqp.Dial(url)
}

// Init - initialize the handler
func (h *Handler) Init() (err error) {

	err = h.Crypto.Init()
	if err != nil {
		return fmt.Errorf("Error during crypto init: %w", err)
	}

	err = h.FileHandler.Init()
	if err != nil {
		return fmt.Errorf("Error during file handler init: %w", err)
	}

	err = h.Maskinporten.Init()
	if err != nil {
		return fmt.Errorf("Error during maskinporten init: %w", err)
	}

	if h.LogLevel != "" {
		lvl, err := logrus.ParseLevel(h.LogLevel)
		if err != nil {
			return fmt.Errorf("Failed to parse log level: %w", err)
		}

		log.Logger.SetLevel(lvl)
	}

	h.Sender.Maskinporten = h.Maskinporten

	return nil
}

func (h *Handler) createEncryptedBody(status MessageType, body string) (io.Reader, error) {
	// create encrypted body
	f := fiksMsgResponse{
		MessageType: string(status),
		Message:     body,
	}

	data, _ := json.Marshal(f)
	return h.Crypto.EncryptASICE(bytes.NewReader(data))
}

// reportMsgResult - helper func that informs Fiks about result of processing message.
func (h *Handler) reportMsgResult(
	fiksMsg fiksMessage,
	status MessageType,
	body string,
) (err error) {
	log := fiksMsg.LoggerWithFields()

	// Inform FIKS about data processing status
	metaData := fiksMsgMetadata{
		ResponseToMsgID:    fiksMsg.GetMeldingID(),
		SenderAccountID:    fiksMsg.GetClientID(),
		RecipientAccountID: fiksMsg.GetAvsenderID(),
		MessageType:        string(status),
	}

	log.Debugf("Reporting status '%s' to FIKS.", status)

	var encryptedData io.Reader

	// If message type failed, error msg (body) should be included
	if status == MessageTypeFailed {
		encryptedData, err = h.createEncryptedBody(status, body)
		if err != nil {
			return err
		}
	}

	// send message to FIKS
	sendErr := h.Sender.Send(metaData, encryptedData)
	if sendErr != nil {
		return fmt.Errorf("could not send message to FIKS: %s", sendErr.Error())
	}

	log.Debugf("Successfully sent message to FIKS.")
	return nil
}

func (h *Handler) reportSuccess(delivery fiksMessage) error {
	return h.reportMsgResult(delivery, MessageTypeSucceeded, "")
}

func (h *Handler) reportFailedReturnOriginalErr(delivery fiksMessage, err error) error {
	sendErr := h.reportMsgResult(delivery, MessageTypeFailed, err.Error())

	//TODO: consider returning sendErr + err.
	if sendErr != nil {
		delivery.LoggerWithFields().Errorf("Unable to report error to FIKS: %s", sendErr.Error())
	}
	return err
}

// revive:disable-next-line:unused-receiver
// getMessageData - retrieves message data from message
// Currently only supports retrieving message from body,
// and does not respect references to Dokumentlager in headers.
func (h *Handler) getMessageData(delivery fiksMessage) (io.Reader, error) {
	//TODO: Support retrieving data from Dokumentlager.
	//		Currently only supports retrieving directly from msg body.
	//		check header 'dokumentlager-id' set, then download that ID from
	//		dokumentlager (todo: implement)
	return bytes.NewReader(delivery.Body), nil
}

// handleAMQMessage - handles amqp message. First decrypts contents, then
// writes these to a zip file. Sends a success/fail to FIKS.
func (h *Handler) handleAMQMessage(delivery fiksMessage) error {

	log := delivery.LoggerWithFields()

	log.Debugf("Handling AMQP message.")

	log.Debugf("Trying to decrypt file.")

	encrypted, err := h.getMessageData(delivery)
	if err != nil {
		return h.reportFailedReturnOriginalErr(delivery, err)
	}

	decrypted, err := h.Crypto.Decrypt(encrypted)
	if err != nil {
		return h.reportFailedReturnOriginalErr(delivery, err)
	}

	log.Debugf("Trying to write content to file.")
	err = h.FileHandler.writeContentToFile(
		decrypted.Bytes(),
		delivery.GetMeldingID(),
	)
	if err != nil {
		return h.reportFailedReturnOriginalErr(delivery, err)
	}

	log.Debugf("Successfully processed message")
	return h.reportSuccess(delivery)
}

// Run - starts fix handler, which reads new messages from activeMQ queue,
// decrypts them and writes unzipped content to specified path
func (h *Handler) Run() (err error) {

	// Connect to activemq
	client, err := h.getAMQPClient()
	if err != nil {
		log.Logger.Fatal("Dialing AMQP server:", err)
		return
	}
	log.Logger.Info("Connected")

	// Open queue channel for reading
	channel, err := client.Channel()
	if err != nil {
		return
	}

	queue, err := channel.Consume(
		h.RabbitMQ.Queue,
		h.RabbitMQ.Consumer,
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		return
	}

	// process messages
	for message := range queue {

		fiksMsg := fiksMessage{&message}

		log := fiksMsg.LoggerWithFields()
		err = h.handleAMQMessage(fiksMsg)
		if err != nil {
			log.Errorf("Failed to handle AMQP message: %s", fiksMsg.GetMeldingID(), err)
		} else {
			log.Infof("Successfully handled message")
		}

		// Always ack.
		//TODO: Consider alternate processing on errors
		//		I.e. if report failure fails.
		message.Ack(true)
	}

	// TODO: consider reconnecting rather than shutting down.
	// connection broken, return err.
	return
}
