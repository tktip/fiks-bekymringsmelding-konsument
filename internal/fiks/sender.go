package fiks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/tktip/fiks-bekymringsmelding-konsument/internal/log"
	"github.com/tktip/maskinporten/pkg/maskinporten"
)

// fiksMsgMetadata - message metadata expected by FIKS
type fiksMsgMetadata struct {
	MessageID          string `json:"meldingId,omitempty"`
	MessageType        string `json:"meldingType,omitempty"`
	SenderAccountID    string `json:"avsenderKontoId,omitempty"`
	RecipientAccountID string `json:"mottakerKontoId,omitempty"`

	TTL               int64             `json:"ttl,omitempty"`
	ResponseToMsgID   string            `json:"svarPaMelding,omitempty"`
	DocumentStorageID string            `json:"dokumentlagerId,omitempty"`
	Headers           map[string]string `json:"headere,omitempty"`
}

// fiksMsgResponse - response on FIKS send msg
type fiksMsgResponse struct {
	MessageType string `json:"meldingType,omitempty"`
	Message     string `json:"melding,omitempty"`
}

// Sender - helper struct to send messages to and from FIKS
type Sender struct {
	IntegrationID string `yaml:"integrationId"`
	IntegrationPW string `yaml:"integrationPassword"`
	URL           string `yaml:"url"`
	client        *http.Client
	Maskinporten  maskinporten.Handler
}

func (s *Sender) getClient() *http.Client {
	if s.client == nil {
		s.client = &http.Client{
			Timeout: time.Second * 30,
		}
	}
	return s.client
}

// createMultipartForm - creates multipart form data from metadata and body,
// returns a form data reader, content type (and error if failed).
func createMultipartForm(metaData fiksMsgMetadata, body io.Reader) (io.Reader, string, error) {
	requestBody, err := json.Marshal(metaData)
	if err != nil {
		return nil, "", err
	}

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	err = writer.WriteField("metadata", string(requestBody))
	if err != nil {
		return nil, "", err
	}

	// TODO: Verify that message body actually arrrives correctly @FIKS.
	if body != nil {
		part2, err := writer.CreateFormFile("data", uuid.New().String())
		if err != nil {
			return nil, "", err
		}

		_, err = io.Copy(part2, body)
		if err != nil {
			return nil, "", err
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, "", err
	}

	return payload, writer.FormDataContentType(), nil
}

// createFIKSHTTPRequest - helper method for creating http request sending multipart form data
func (s *Sender) createFIKSHTTPRequest(body io.Reader, contentType string) (*http.Request, error) {

	req, err := http.NewRequest(http.MethodPost, s.URL, body)
	accessToken, err := s.Maskinporten.CreateAccessToken()
	if err != nil {
		return nil, err
	}

	req.Header.Add("IntegrasjonId", s.IntegrationID)
	req.Header.Add("IntegrasjonPassord", s.IntegrationPW)
	req.Header.Add("Authorization", "Bearer "+accessToken.AccessToken)
	req.Header.Set("Content-Type", contentType)

	return req, nil
}

// Send - sends a message with given metadata and body as formdata content
func (s *Sender) Send(metaData fiksMsgMetadata, body io.Reader) error {

	formData, contentType, err := createMultipartForm(metaData, body)
	if err != nil {
		return err
	}

	req, err := s.createFIKSHTTPRequest(formData, contentType)
	if err != nil {
		return err
	}

	res, err := s.getClient().Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode > 299 ||
		res.StatusCode < 200 {
		responseBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Logger.Errorf("Failed to read fiksMsgResponse body: %s", err.Error())
		}
		return fmt.Errorf(
			"Invalid fiksMsgResponse code from KS send (%d): %s",
			res.StatusCode,
			responseBody,
		)
	}
	return nil
}
