package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/DatioBD/retrievault/utils/log"
	"github.com/DatioBD/retrievault/utils/os/permissions"
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
	return &Generic{writer: writer{wg: new(sync.WaitGroup)}}
}

func (g *Generic) FetchSecret(ctx context.Context, vaultPath, dest string, client *api.Logical, wg *sync.WaitGroup, e chan error) {
	log.Msg.WithField("vault_path", vaultPath).Debug("Fetching secret at path")
	defer wg.Done()
	secrets, err := client.Read(vaultPath)
	if err != nil {
		e <- err
		return
	}
	er := make(chan error)
	g.wg.Add(len(secrets.Data))
	for key, secret := range secrets.Data {
		perm := os.FileMode(0644)
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
		file := fmt.Sprintf("%s/%s", dest, key)
		if fileName, ok := g.Keys[key]; ok {
			if fileName.Path != "" {
				if path.IsAbs(path.Clean(fileName.Path)) {
					file = fileName.Path
				} else {
					file = fmt.Sprintf("%s/%s", dest, fileName.Path)
				}
			}
			if fileName.Perm != "" {
				var err error
				perm, err = permissions.StringToFileMode(fileName.Perm)
				if err != nil {
					log.Msg.WithField("permissions", fileName.Perm).Error("Wrong permission format. Must be something like \"0644\" or \"0600\"")
					e <- err
					return
				}
			}
		}
		log.Msg.WithField("perm", perm).Debug("Permissions applied")
		go g.writeInFile(path.Clean(file), []byte(stringSecret), perm, er)
		select {
		case err := <-er:
			log.Msg.WithField("secret", key).Error("Error when writing secret to file")
			e <- err
			return
		default:
		}
	}
	g.wg.Wait()
}
