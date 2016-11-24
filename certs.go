package main

import (
	"context"
	"sync"

	"github.com/hashicorp/vault/api"
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
	wg         sync.WaitGroup
	writer
}

type certParams struct {
	fileParameters
}

func (c *Certs) FetchSecret(ctx context.Context, vaultPath, dest string, client *api.Logical, e chan error) {
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
	c.wg.Add(3)
	for key, secret := range secrets.Data {
		if key ==
	}
}
