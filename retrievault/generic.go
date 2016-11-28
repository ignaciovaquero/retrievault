package retrievault

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/DatioBD/retrievault/utils/log"
	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
)

type Generic struct {
	Keys   map[string]genericParams `json:"keys,omitempty"`
	secret *api.Secret
	writer
}

type genericParams struct {
	fileParameters
}

func NewGeneric() *Generic {
	return new(Generic)
}

func (g *Generic) FetchSecret(ctx context.Context, vaultPath, dest string, client *api.Logical, e chan error) {
	log.Msg.WithField("vault_path", vaultPath).Debug("Fetching secret at path")
	secrets, err := client.Read(vaultPath)
	if err != nil {
		e <- err
		return
	}
	er := make(chan error)
	for key, secret := range secrets.Data {
		select {
		case <-ctx.Done():
			log.Msg.Error("Parent context cancelled")
			e <- ctx.Err()
			return
		default:
		}
		stringSecret, ok := secret.(string)
		if !ok {
			errMsg := "Error when getting secret as string"
			log.Msg.WithField("secret", key).Error(errMsg)
			e <- fmt.Errorf(errMsg)
			return
		}
		var (
			perm os.FileMode
			file string
			err  error
		)
		if fparams, ok := g.Keys[key]; ok {
			file, perm, err = g.getDestAndPerms(key, fparams.fileParameters, dest)
		} else {
			file, perm, err = g.getDestAndPerms(key, fileParameters{}, dest)
		}
		if err != nil {
			log.Msg.WithFields(logrus.Fields{
				"secret":      key,
				"permissions": perm,
			}).Error(err.Error())
			e <- err
			return
		}
		go g.writeInFile(path.Clean(file), []byte(stringSecret), perm, er)
	}

	for i := 0; i < len(secrets.Data); i++ {
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
