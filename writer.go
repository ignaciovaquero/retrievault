package main

import (
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/DatioBD/retrievault/utils/log"
)

type fileParameters struct {
	Path string `json:"path,omitempty"`
	Perm string `json:"perm,omitempty"`
}

type writer struct {
	wg *sync.WaitGroup
}

func (w *writer) writeInFile(filePath string, secret []byte, perm os.FileMode, e chan error) {
	log.Msg.WithField("file", filePath).Debug("Writing secret in file")
	defer w.wg.Done()
	directory := path.Dir(filePath)
	if err := os.MkdirAll(directory, perm); err != nil {
		e <- err
	}
	if err := ioutil.WriteFile(filePath, secret, perm); err != nil {
		e <- err
	}
}
