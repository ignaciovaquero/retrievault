package retrievault

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/DatioBD/retrievault/utils/log"
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
	return new(Certs)
}

func (c *Certs) processSingleSecret(secret interface{}) ([]byte, error) {
	stringSecret, ok := secret.(string)
	if !ok {
		return []byte{}, fmt.Errorf("Error when getting secret as string")
	}
	return []byte(fmt.Sprintf("%s\n", stringSecret)), nil
}

func (c *Certs) processArraySecret(secret interface{}) ([]byte, error) {
	arraySecret, ok := secret.([]interface{})
	var CAs []byte
	if !ok {
		return []byte{}, fmt.Errorf("Error when getting secret as array")
	}
	for _, ca := range arraySecret {
		caString, found := ca.(string)
		if !found {
			return []byte{}, fmt.Errorf("Error when getting secret as array")
		}
		bCA := []byte(fmt.Sprintf("%s\n", caString))
		CAs = append(CAs, bCA...)
	}
	return CAs, nil
}

func (c *Certs) FetchSecret(ctx context.Context, vaultPath, dest string, client *api.Logical, e chan error) {
	log.Msg.WithField("vault_path", vaultPath).Debug("Fetching secret at path")
	secrets, err := client.Write(vaultPath, map[string]interface{}{
		"common_name": c.CommonName,
		"ttl":         c.TTL,
		"alt_names":   strings.Join(c.AltNames, ","),
		"ip_sans":     strings.Join(c.IPSans, ","),
	})
	if err != nil {
		e <- err
		return
	}

	er := make(chan error)
	var certificateData []byte
	var caChainData []byte
	wait := 0
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
			wait++

			var (
				file string
				data []byte
				perm os.FileMode
				err  error
			)

			file, perm, err = c.getDestAndPerms("ca.crt", c.CACert.fileParameters, dest)
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
			wait++

			var (
				file string
				data []byte
				perm os.FileMode
				err  error
			)

			file, perm, err = c.getDestAndPerms("cert.key", c.Key.fileParameters, dest)
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

			var (
				file string
				perm os.FileMode
				err  error
			)
			file, perm, err = c.getDestAndPerms("cert.crt", c.Cert.fileParameters, dest)
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

			if len(caChainData) > 0 {
				wait++
				certificateData = append(certificateData, caChainData...)
				go c.writeInFile(file, certificateData, perm, er)
			}

		case CAChain:

			var (
				file string
				perm os.FileMode
				err  error
			)
			file, perm, err = c.getDestAndPerms("cert.crt", c.Cert.fileParameters, dest)
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

			if len(certificateData) > 0 {
				wait++
				certificateData = append(certificateData, caChainData...)
				go c.writeInFile(file, certificateData, perm, er)
			}
		default:
			continue
		}
	}

	for i := 0; i < wait; i++ {
		select {
		case <-ctx.Done():
			log.Msg.Error("Parent context cancelled")
			e <- ctx.Err()
			return
		case err := <-er:
			if err != nil {
				log.Msg.Error("Error when writing secret to file")
				e <- err
				return
			}
		}
	}

	e <- nil
	return
}
