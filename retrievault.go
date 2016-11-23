package main

import (
	"context"
	"encoding/json"
	"io/ioutil"

	env "github.com/DatioBD/retrievault/utils/environment"
	"github.com/DatioBD/retrievault/utils/log"
	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
)

// Retriever is an interface that wraps the basic FetchSecret method.
type Retriever interface {

	// FetchSecret fetches a secret from Vault and returns any error encountered.
	FetchSecret(ctx context.Context, s Secret) (api.Secret, error)
}

type destPath string

// RetrieveVault is a struct which holds the configuration for the application
type RetrieVault struct {

	// LogLevel is the level for logging
	LogLevel string `json:"log_level,omitempty"`

	// LogFile is the output file for logging
	LogFile string `json:"log_file,omitempty"`

	// Secrets is a map that has path values as keys, and Secret structures as
	// values. The secrets, once fetched, will be stored at the path indicated
	// in the keys
	Secrets map[destPath]Secret `json:"secrets"`

	// CACertPath is the path to a PEM-encoded CA cert file to use to verify the
	// Vault server SSL certificate.
	CACertPath string `json:"ca_cert_path,omitempty"`

	// Insecure enables or disables SSL verification
	Insecure bool `json:"insecure,omitempty"`

	// VaultAddr is the address of the Vault server. This should be a complete
	// URL such as "http://vault.example.com". If you need a custom SSL
	// cert or want to enable insecure mode, you need to specify a custom
	// HttpClient. If not set, "https://127.0.0.1:8200" will be taken as default.
	VaultAddr string `json:"vault_addr,omitempty"`

	// VaultToken is the Vault token used to retrieve all secrets
	VaultToken string `json:"vault_token,omitempty"`
	client     *api.Logical
}

// Secret is a struct that contains information about how to retrieve
// a secret from Vault. Type can only be one of: certs, generic.
type Secret struct {
	Type       string          `json:"type"`
	VaultPath  string          `json:"vault_path"`
	Parameters json.RawMessage `json:"parameters,omitempty"`
}

func (retrievault *RetrieVault) readConfiguration(path string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, retrievault)
	if err != nil {
		return err
	}
	return nil
}

func setupApp(configPath string) (*RetrieVault, error) {
	retrievault := new(RetrieVault)
	if err := retrievault.readConfiguration(configPath); err != nil {
		log.Msg.WithField("msg", err.Error()).Error("Error when reading configuration")
		return nil, err
	}

	// Setting log configuration
	if err := log.SetLogLevel(retrievault.LogLevel); err != nil {
		log.Msg.WithFields(logrus.Fields{
			"log_level": retrievault.LogLevel,
			"msg":       err.Error(),
		}).Warn("Error when setting log level.")
	}
	if err := log.SetOutput(retrievault.LogFile); err != nil {
		log.Msg.WithFields(logrus.Fields{
			"log_file": retrievault.LogFile,
			"msg":      err.Error(),
		}).Warn("Error when setting log output file.")
	}

	// Setting Vault client configuration
	config := api.DefaultConfig()
	if retrievault.CACertPath != "" {
		tlsconfig := &api.TLSConfig{
			CACert:   retrievault.CACertPath,
			Insecure: retrievault.Insecure,
		}
		if err := config.ConfigureTLS(tlsconfig); err != nil {
			log.Msg.WithFields(logrus.Fields{
				"msg":       err.Error(),
				"tlsconfig": tlsconfig,
			}).Error("Error when applying TLS configuration")
			return nil, err
		}
	}
	if retrievault.VaultAddr != "" {
		config.Address = retrievault.VaultAddr
	}
	if err := config.ReadEnvironment(); err != nil {
		log.Msg.WithField("msg", err.Error()).Warn("Error when loading configuration from environment")
	}
	client, err := api.NewClient(config)
	if err != nil {
		log.Msg.WithFields(logrus.Fields{
			"msg":    err.Error(),
			"config": config,
		}).Error("Error when creating Vault client from configuration")
		return nil, err
	}
	if env.GetOrElse("VAULT_TOKEN", "") == "" && retrievault.VaultToken != "" {
		client.SetToken(retrievault.VaultToken)
	}
	retrievault.client = client.Logical()
	return retrievault, nil
}

func (r *RetrieVault) FetchSecrets(ctx context.Context) error {
	for path, secret := range r.Secrets {

	}
}
