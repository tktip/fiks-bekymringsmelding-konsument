package crypt

import (
	"archive/zip"
	"github.com/tktip/fiks-bekymringsmelding-konsument/internal/log"
	"github.com/tktip/pkcs7"
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"io/ioutil"
)

// Handler - Provides methods for decryption and encryption
type Handler struct {
	JarLocation string `yaml:"jarLocation"`
	PrivKeyPath string `yaml:"privKeyPath"`
	PubKeyPath  string `yaml:"pubKeyPath"`
	cert        *x509.Certificate
	privateKey  crypto.PrivateKey
}

// Init - initialize decryptor, reading and parsing keys.
func (d *Handler) Init() error {
	log.Logger.Debug("Initializing decryptor")
	pubKey, err := ioutil.ReadFile(d.PubKeyPath)
	if err != nil {
		return err
	}

	pubBlock, rest := pem.Decode(pubKey)
	if len(rest) > 0 {
		return errors.New("public key decode remainder not empty")
	}

	cert, err := x509.ParseCertificate(pubBlock.Bytes)
	if err != nil {
		return err
	}

	// Read the private pubKey
	privKey, err := ioutil.ReadFile(d.PrivKeyPath)
	if err != nil {
		return err
	}

	// Extract the PEM-encoded data pubBlock
	privBlock, rest := pem.Decode(privKey)
	if err != nil {
		return errors.New("private key decode remainder not empty")
	}

	// Decode the RSA private pubKey
	priv, err := x509.ParsePKCS8PrivateKey(privBlock.Bytes)
	if err != nil {
		return err
	}

	d.cert = cert
	d.privateKey = priv.(crypto.PrivateKey)
	log.Logger.Debug("Handler initialized")

	//Set default encryption algorithm to id-RSAES-OAEP
	pkcs7.KeyEncryptionAlgorithm = pkcs7.OIDEncryptionAlgorithmidRSAESOAEP
	return nil
}

// Decrypt - decrypts encrypted input based on metadata in file
// see pkcs7 package for more.
func (d *Handler) Decrypt(
	encrypted io.Reader,
) (
	*bytes.Buffer,
	error,
) {

	data, err := ioutil.ReadAll(encrypted)
	pkcs7Obj, err := pkcs7.Parse(data)
	if err != nil {
		return nil, err
	}

	data, err = pkcs7Obj.Decrypt(d.cert, d.privateKey)
	return bytes.NewBuffer(data), err
}

// EncryptASICE - encrypts data in Difi's ASCI-E format.
// A zipped file encrypted with id-RSA-OAEP (1, 2, 840, 113549, 1, 1, 7)
func (d *Handler) EncryptASICE(
	data io.Reader,
) (
	buf *bytes.Buffer,
	err error,
) {

	var tmp bytes.Buffer

	//Zip file
	writer := zip.NewWriter(&tmp)
	var file io.Writer
	file, err = writer.Create("data")
	if err != nil {
		return
	}

	_, err = io.Copy(file, data)
	if err != nil {
		return
	}

	err = writer.Close()
	if err != nil {
		return
	}

	var encrypted []byte
	//encrypt zipped file
	encrypted, err = pkcs7.Encrypt(tmp.Bytes(), []*x509.Certificate{d.cert})
	if err != nil {
		return
	}

	//enclose in buffer and return
	return bytes.NewBuffer(encrypted), nil
}
