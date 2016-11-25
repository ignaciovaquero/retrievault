package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/DatioBD/retrievault/utils/log"
	"github.com/DatioBD/retrievault/utils/os/permissions"
	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
)

const (
	issuingCA   = "issuing_ca"
	CAChain     = "ca_chain"
	privateKey  = "private_key"
	certificate = "certificate"
)

type Certs struct {
	CommonName string     `json:"common_name"`
	TTL        string     `json:"ttl,omitempty"`
	AltNames   []string   `json:"alt_names,omitempty"`
	IPSans     []string   `json:"ip_sans,omitempty"`
	Key        certParams `json:"key,omitempty"`
	Cert       certParams `json:"cert,omitempty"`
	CACert     certParams `json:"ca_cert,omitempty"`
	secret     *api.Secret
	writer
}

type certParams struct {
	fileParameters
}

func NewCerts() *Certs {
	return &Certs{writer: writer{wg: new(sync.WaitGroup)}}
}

func (c *Certs) getDestAndPerms(defaultFile string, params certParams, dest string) (string, os.FileMode, error) {
	perm := os.FileMode(0644)
	file := fmt.Sprintf("%s/%s", dest, defaultFile)
	if params.Path != "" {
		if path.IsAbs(path.Clean(params.Path)) {
			file = params.Path
		} else {
			file = fmt.Sprintf("%s/%s", dest, params.Path)
		}
	}
	if params.Perm != "" {
		var err error
		perm, err = permissions.StringToFileMode(params.Perm)
		if err != nil {
			return file, 0, fmt.Errorf("Wrong permission format. Must be something like \"0644\" or \"0600\"")
		}
	}
	return path.Clean(file), perm, nil
}

func (c *Certs) processSingleSecret(secret interface{}) ([]byte, error) {
	stringSecret, ok := secret.(string)
	if !ok {
		return []byte{}, fmt.Errorf("Error when getting secret as string")
	}
	return []byte(stringSecret), nil
}

func (c *Certs) processArraySecret(secret interface{}) ([]byte, error) {
	arraySecret, ok := secret.([]string)
	var CAs []byte
	if !ok {
		return []byte{}, fmt.Errorf("Error when getting secret as array")
	}
	for _, ca := range arraySecret {
		bCA := []byte(ca)
		CAs = append(CAs, bCA...)
	}
	return CAs, nil
}

func (c *Certs) FetchSecret(ctx context.Context, vaultPath, dest string, client *api.Logical, wg *sync.WaitGroup, e chan error) {
	log.Msg.WithField("vault_path", vaultPath).Debug("Fetching secret at path")
	defer wg.Done()
	secrets, err := client.Write(vaultPath, map[string]interface{}{
		"common_name": c.CommonName,
		"ttl":         c.TTL,
		"alt_names":   c.AltNames,
		"ip_sans":     c.IPSans,
	})
	if err != nil {
		e <- err
		return
	}

	er := make(chan error)
	var certificateData []byte
	var caChainData []byte
	written := false

	for key, secret := range secrets.Data {
		select {
		case <-ctx.Done():
			log.Msg.Error("Parent context cancelled")
			e <- ctx.Err()
			return
		default:
		}

		switch key {
		case issuingCA:
			c.wg.Add(1)

			var (
				file string
				data []byte
				perm os.FileMode
				err  error
			)

			file, perm, err = c.getDestAndPerms("ca.crt", c.CACert, dest)
			if err != nil {
				log.Msg.WithFields(logrus.Fields{
					"secret":      key,
					"permissions": perm,
				}).Error(err.Error())
				e <- err
				return
			}
			data, err = c.processSingleSecret(secret)

			if err != nil {
				log.Msg.WithFields(logrus.Fields{
					"secret":      key,
					"permissions": perm,
				}).Error(err.Error())
				e <- err
				return
			}

			go c.writeInFile(file, data, perm, er)

		case privateKey:
			c.wg.Add(1)

			var (
				file string
				data []byte
				perm os.FileMode
				err  error
			)

			file, perm, err = c.getDestAndPerms("cert.key", c.Key, dest)
			if err != nil {
				log.Msg.WithFields(logrus.Fields{
					"secret": key,
				}).Error(err.Error())
				e <- err
				return
			}
			data, err = c.processSingleSecret(secret)

			if err != nil {
				log.Msg.WithFields(logrus.Fields{
					"secret": key,
				}).Error(err.Error())
				e <- err
				return
			}

			go c.writeInFile(file, data, perm, er)

		case certificate:
			c.wg.Add(1)

			var (
				file string
				perm os.FileMode
				err  error
			)
			file, perm, err = c.getDestAndPerms("cert.crt", c.Cert, dest)
			if err != nil {
				log.Msg.WithFields(logrus.Fields{
					"secret":      key,
					"permissions": perm,
				}).Error(err.Error())
				e <- err
				return
			}

			certificateData, err = c.processSingleSecret(secret)

			if err != nil {
				log.Msg.WithFields(logrus.Fields{
					"secret": key,
				}).Error(err.Error())
				e <- err
				return
			}

			if len(caChainData) > 0 && !written {
				certificateData = append(certificateData, caChainData...)
				go c.writeInFile(file, certificateData, perm, er)
				written = true
			}

		case CAChain:
			c.wg.Add(1)

			var (
				file string
				perm os.FileMode
				err  error
			)
			file, perm, err = c.getDestAndPerms("cert.crt", c.Cert, dest)
			if err != nil {
				log.Msg.WithFields(logrus.Fields{
					"secret":      key,
					"permissions": perm,
				}).Error(err.Error())
				e <- err
				return
			}

			caChainData, err = c.processArraySecret(secret)

			if err != nil {
				log.Msg.WithFields(logrus.Fields{
					"secret": key,
				}).Error(err.Error())
				e <- err
				return
			}

			if len(certificateData) > 0 && !written {
				certificateData = append(certificateData, caChainData...)
				go c.writeInFile(file, certificateData, perm, er)
				written = true
			}

		default:
			continue
		}

		select {
		case err := <-er:
			log.Msg.WithField("secret", key).Error("Error when writing secret to file")
			e <- err
			return
		default:
		}
	}
	c.wg.Wait()
}
