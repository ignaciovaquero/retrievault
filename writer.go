package main

import (
	"io/ioutil"
	"os"
	"path"
)

type fileParameters struct {
	Name string `json:"name,omitempty"`
	Perm string `json:"perm,omitempty"`
}

type writer struct{}

func (w *writer) writeInFile(filePath string, secret []byte, perm os.FileMode, e chan error) {
	directory := path.Dir(filePath)
	if err := os.MkdirAll(directory, perm); err != nil {
		e <- err
	}
	if err := ioutil.WriteFile(filePath, secret, perm); err != nil {
		e <- err
	}
}
