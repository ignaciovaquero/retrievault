package retrievault

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/DatioBD/retrievault/utils/log"
	"github.com/DatioBD/retrievault/utils/os/permissions"
)

type writer struct{}

type fileParameters struct {
	Path string `json:"path,omitempty"`
	Perm string `json:"perm,omitempty"`
}

func (w *writer) getDestAndPerms(defaultFile string, params fileParameters, dest string) (string, os.FileMode, error) {
	perm := os.FileMode(0644)
	if dest == "" {
		dest = "." // relative to current directory
	}
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

func (w *writer) writeInFile(filePath string, secret []byte, perm os.FileMode, e chan error) {
	log.Msg.WithField("file", filePath).Debug("Writing secret in file")
	directory := path.Dir(filePath)
	fi, err := os.Stat(directory)
	if err != nil {
		log.Msg.WithField("directory", directory).Debug("Directory doesn't exist. Creating...")
		if err = os.MkdirAll(directory, os.FileMode(0700)); err != nil {
			e <- err
			return
		}
	} else if !fi.IsDir() {
		e <- fmt.Errorf("Not a directory: %s", directory)
		return
	}
	if err := ioutil.WriteFile(filePath, secret, perm); err != nil {
		e <- err
		return
	}
	e <- nil
	return
}
