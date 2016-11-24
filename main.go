package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/urfave/cli.v2"
)

var _ = logrus.DebugLevel
var app = cli.NewApp()
var appName = "retrievault"

func init() {
	app.Name = appName
	app.Usage = "Retrieve Vault secrets and expose them into files or environment variables"
	app.Version = "0.1.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: "/etc/retrievault/config.json",
			Usage: "Path to the configuration file",
		},
	}
	app.Action = run
}

func run(c *cli.Context) error {
	retrievault, err := setupApp(c.String("config"))
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Error setting up %s", appName), 1)
	}
	timeout, _ := time.ParseDuration("30s")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err = retrievault.FetchSecrets(ctx)
	if err != nil {
		return cli.NewExitError("Error retrieving secrets", 1)
	}
	return nil
}

func main() {
	app.Run(os.Args)
}
