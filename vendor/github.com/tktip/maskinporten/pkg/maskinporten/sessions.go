package maskinporten

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/dgrijalva/jwt-go"
)

//Handler - provides access to maskinporten functionality
type Handler struct {
	signKey   *rsa.PrivateKey
	x5CHeader []string

	PrivateKey    string `yaml:"privateKey"`
	PublicKey     string `yaml:"publicKey"`
	TokenEndpoint string `yaml:"tokenEndpoint"`
	Scope         string `yaml:"scope"`
	Audience      string `yaml:"aud"`
	Issuer        string `yaml:"iss"`

	client *http.Client
}

// TokenResponse is the form of the response when fetching an access token from ID-porten
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

//Init - initializes the handler. Reads and parses priv and pub key
func (h *Handler) Init() (err error) {
	priv, x5c, err := readCertFiles(h.PrivateKey, h.PublicKey)
	if err != nil {
		return
	}
	h.signKey = priv
	h.x5CHeader = x5c

	return
}

// readCertFiles read
func readCertFiles(privFile string, pubFile string) (
	priv *rsa.PrivateKey, x5c []string, err error,
) {
	signingBytes, err := ioutil.ReadFile(privFile)
	if err != nil {
		err = fmt.Errorf("error reading key file: %v", err)
		return
	}
	priv, err = jwt.ParseRSAPrivateKeyFromPEM(signingBytes)
	if err != nil {
		err = fmt.Errorf("error parsing key bytes: %v", err)
		return
	}

	pubBytes, err := ioutil.ReadFile(pubFile)
	if err != nil {
		err = fmt.Errorf("error reading pub file: %v", err)
		return
	}

	x5c, err = convertPublicKeyToX5CHeader([]string{}, pubBytes)
	if err != nil {
		err = fmt.Errorf("error parsing pubfile to x5c-header: %v", err)
		return
	}
	return
}

func convertPublicKeyToX5CHeader(soFar []string, bs []byte) ([]string, error) {
	block, rest := pem.Decode(bs)
	if block == nil {
		return nil, errors.New("invalid key: failed to parse header")
	}

	soFar = append(soFar, base64.StdEncoding.EncodeToString(block.Bytes))

	if len(rest) == 0 {
		return soFar, nil
	}

	return convertPublicKeyToX5CHeader(soFar, rest)
}

func (h *Handler) createToken() (string, error) {
	claims := jwt.MapClaims{
		"scope": h.Scope,
		"aud":   h.Audience,
		"iss":   h.Issuer,
		"jti":   uuid.New().String(),
		"exp":   time.Now().Add(time.Minute * 2).Unix(),
		"iat":   time.Now().Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	t.Header["x5c"] = h.x5CHeader
	tokenString, err := t.SignedString(h.signKey)
	if err != nil {
		return "", fmt.Errorf("error creating signed token: %v", err)
	}

	return tokenString, nil
}

func (h *Handler) getClient() *http.Client {
	if h.client == nil {
		h.client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	return h.client
}

//CreateAccessToken - creates an Maskinporten access token
func (h *Handler) CreateAccessToken() (tokenRes TokenResponse, err error) {
	tokenContent, err := h.createToken()
	if err != nil {
		return
	}

	body := "grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Ajwt-bearer&assertion=" + tokenContent

	req, err := http.NewRequest(
		http.MethodPost,
		h.TokenEndpoint,
		strings.NewReader(body),
	)
	if err != nil {
		err = fmt.Errorf("error creating request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := h.getClient().Do(req)
	if err != nil {
		err = fmt.Errorf("error doing request: %v", err)
		return
	}
	defer res.Body.Close()

	resBod, err := ioutil.ReadAll(res.Body)
	if err != nil {
		err = fmt.Errorf("error reading response: %v", err)
		return
	}

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("expected response code 200, got %d. Response body was: %s",
			res.StatusCode, resBod)
		return
	}

	err = json.Unmarshal(resBod, &tokenRes)
	if err != nil {
		err = fmt.Errorf("error unmarshalling response-body: %v", err)
		return
	}
	return
}
