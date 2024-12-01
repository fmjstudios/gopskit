// package vault implements an HTTP API client for the secret-management solution Vault from
// HashCorp Inc.
package vault

import (
	"encoding/json"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/fmjstudios/gopskit/pkg/api"
	fs "github.com/fmjstudios/gopskit/pkg/fsi"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/vault-client-go"
)

type ServerConfig struct {
	DisableMLock bool        `hcl:"disable_mlock"`
	UI           bool        `hcl:"ui"`
	Seal         *sealConfig `hcl:"seal,block"`
	Remain       hcl.Body    `hcl:",remain"`
}

type sealConfig struct {
	Type   string   `hcl:"type,label"`
	Remain hcl.Body `hcl:",remain"`
}

// Auth is a custom type which is used to write and load Vault credentials to and from a file
type Auth struct {
	// Created is a timestamp to know when the credentials were created
	Created time.Time `json:"created,omitempty"`

	// Keys are the Vault unseal keys
	Keys []string `json:"keys"`

	// KeysB64 are the Vault unseal keys, base64-encoded
	KeysB64 []string `json:"keys_base64"`

	// Token is the Vault Auth token, used to authenticate to the API
	Token string `json:"token"`
}

// compile-time checks
// var _ api.Client = &Client{}
// var _ api.Credentials = &Auth{}

// Save implements the Credentials interface for Auth
func (a *Auth) Save(dst string) error {
	jsn, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}

	return fs.Write(dst, jsn)
}

// Load implements the Credentials interface for Auth
func (a *Auth) Load(src string) error {
	raw, err := fs.Read(src)
	if err != nil {
		return err
	}

	var creds Auth
	err = json.Unmarshal(raw, &creds)
	if err != nil {
		return err
	}

	a = &creds
	return nil
}

type Config struct {
	// Host defines the hostname for the new Keycloak API client
	Host string

	// Dir is the directory within we which we may or may not persist the API
	// credentials
	Dir string

	// TimeoutSeconds is the amount of time to elapse before prematurely cancelling
	// a HTTP request
	TimeoutSeconds int

	// InsecureTLS defined whether or not the client is set to verify TLS certificates
	// upon making a connection
	InsecureTLS bool
}

type Client struct {
	// api is the underlying gocloak Keycloak API client which we're wrapping
	api *vault.Client

	// auth are the OIDC credentials to use during API requests
	auth *Auth

	// cfg is the Config object the API client was initialized with
	cfg *Config

	// fs is the generic filesystem interface provided by spf13/afero
	// fs afero.Fs

	// path is the filesystem path at which we're persisting credentials, if
	// the user wants to do so
	path string
}

func New(conf *Config) *Client {
	var vaultOpts []vault.ClientOption

	switch {
	case conf.TimeoutSeconds != 0:
		vaultOpts = append(vaultOpts, vault.WithRequestTimeout(time.Duration(60)*time.Second))
	case conf.InsecureTLS:
		vaultOpts = append(vaultOpts, vault.WithTLS(vault.TLSConfiguration{
			InsecureSkipVerify: true,
		}))
	case conf.Host != "":
		vaultOpts = append(vaultOpts, vault.WithAddress(""))
	}

	client, err := vault.New(vaultOpts...)
	if err != nil {
		log.Fatalf("could not create Vault HTTP API client. Error: %v\n", err)
	}

	return &Client{
		api:  client,
		cfg:  conf,
		path: getAuthPath(conf.Dir),
	}
}

type LoginOptions struct {
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	ClientID     string `json:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
}

func (kc *Client) Login() error {
	// var err error
	// impl

	kc.auth.Save(kc.path)
	return nil
}

func (kc *Client) Refresh() error {
	// var err error
	// impl
	return nil
}

func (kc *Client) Auth() api.Credentials {
	return kc.auth
}

func (kc *Client) SetAuth(a api.Credentials) {
	ok := a.(*Auth)
	kc.auth = ok
}

// func (kc *Client) SetRealm(r string) {
// 	kc.cfg.Realm = r
// }

// func (kc *Client) Valid() bool {
// 	return !kc.hasTokenExpired()
// }

// func (kc *Client) AllowInsecureTLS() {
// 	client := kc.api.RestyClient()
// 	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
// }

func (kc *Client) handleAuthorizationError(err error) error {
	// impl

	return nil
}

func (kc *Client) isUnauthorizedErr(err error) bool {
	return strings.Contains(err.Error(), "401 Unauthorized")
}

// func (kc *Client) hasTokenExpired() bool {
// 	validUntil := kc.auth.Created.Add(time.Duration(kc.auth.JWT.ExpiresIn) * time.Second)
// 	return validUntil.Before(time.Now())
// }

func getAuthPath(baseDir string) string {
	return filepath.Join(baseDir, "creds", "vault-credentials.json")
}

func notEmptyStrings(s, t string) bool {
	return s != "" && t != ""
}

func emptyStrings(s, t string) bool {
	return s == "" && t == ""
}
