package retrievault

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	env "github.com/DatioBD/retrievault/utils/environment"
	"github.com/DatioBD/retrievault/utils/log"
	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
)

const (
	certs   = "certs"
	generic = "generic"
)

// Retriever is an interface that wraps the basic FetchSecret method.
type Retriever interface {

	// FetchSecret fetches a secret from Vault and returns any error encountered.
	FetchSecret(ctx context.Context, vaultPath, dest string, client *api.Logical, e chan error)
}

// RetrieveVault is a struct which holds the configuration for the application
type RetrieVault struct {

	// LogLevel is the level for logging
	LogLevel string `json:"log_level,omitempty"`

	// LogFile is the output file for logging
	LogFile string `json:"log_file,omitempty"`

	// Secrets is a map that has path values as keys, and Secret structures as
	// values. The secrets, once fetched, will be stored at the path indicated
	// in the keys
	Secrets []*Secret `json:"secrets"`

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

	client *api.Logical
}

// Secret is a struct that contains information about how to retrieve
// a secret from Vault. Type can only be one of: certs, generic.
type Secret struct {
	Type       string          `json:"type"`
	Path       string          `json:"path"`
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

func SetupApp(configPath, logPath, loglevel string) (*RetrieVault, error) {
	retrievault := new(RetrieVault)
	if err := retrievault.readConfiguration(configPath); err != nil {
		log.Msg.WithField("msg", err.Error()).Error("Error when reading configuration")
		return nil, err
	}

	if retrievault.LogFile == "" {
		retrievault.LogFile = logPath
	}

	if retrievault.LogLevel == "" {
		retrievault.LogLevel = loglevel
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
	if retrievault.CACertPath != "" || retrievault.Insecure {
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
	if retrievault.VaultAddr != "" { // this allows to use localhost if not set
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
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	e := make(chan error)
	wait := 0
	for _, secret := range r.Secrets {
		if secret.Path == "" {
			log.Msg.WithField("secret", secret.VaultPath).Warn("No destination path specified for secret")
			continue
		}
		select {
		// If we cancel the parent context, we must return inmediately
		case <-ctx.Done():
			log.Msg.WithField("msg", ctx.Err().Error()).Error("Context cancelled")
			return ctx.Err()
		default:
		}
		var err error
		var retr Retriever
		switch secret.Type {
		case certs:
			retr = NewCerts()
			err = json.Unmarshal(secret.Parameters, retr)
		case generic:
			retr = NewGeneric()
			err = json.Unmarshal(secret.Parameters, retr)
		default:
			log.Msg.WithField("secret_type", secret.Type).Error("Invalid type.")
			return fmt.Errorf("Invalid secret type %s", secret.Type)
		}

		if err != nil {
			log.Msg.WithFields(logrus.Fields{
				"msg":         err.Error(),
				"secret_type": secret.Type,
			}).Error("Unable to unmarshall parameters for this secret Type")
			return err
		}
		go retr.FetchSecret(cancelCtx, secret.VaultPath, secret.Path, r.client, e)
		wait++
	}

	for i := 0; i < wait; i++ {
		select {
		case <-ctx.Done():
			log.Msg.WithField("msg", ctx.Err().Error()).Error("Context cancelled")
			return ctx.Err()
		case err := <-e:
			if err != nil {
				log.Msg.WithFields(logrus.Fields{
					"msg": err.Error(),
				}).Error("Error when fetching secret")
				return err
			}
		}
	}
	return nil
}
