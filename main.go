package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/DatioBD/retrievault/retrievault"
	"github.com/DatioBD/retrievault/utils/log"
	"gopkg.in/urfave/cli.v2"
)

var app = cli.NewApp()
var appName = "retrievault"

func init() {
	app.Name = appName
	app.Usage = "Retrieve Vault secrets and expose them into files"
	app.Author = "Devops Datio Big Data"
	app.Email = "devops@datiobd.com"
	app.Version = "0.1.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config",
			Value:  "/etc/retrievault/config/config.json",
			Usage:  "Path to the configuration file",
			EnvVar: "RETRIEVAULT_CONFIG_FILE",
		},
		cli.StringFlag{
			Name:   "log-file",
			Value:  "/var/log/retrievault.log",
			Usage:  "Path to the log file. Can be set to \"stderr\" or \"stdout\"",
			EnvVar: "RETRIEVAULT_LOG_FILE",
		},
	}
	app.Action = run
}

func run(c *cli.Context) error {
	rvault, err := retrievault.SetupApp(c.String("config"), c.String("log-file"))
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Error setting up %s", appName), 1)
	}
	timeout, _ := time.ParseDuration("30s")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	log.Msg.Info("Fetching secrets...")
	err = rvault.FetchSecrets(ctx)
	if err != nil {
		return cli.NewExitError("Error retrieving secrets", 1)
	}
	log.Msg.Info("All secrets fetched successfully!")
	return nil
}

func main() {
	app.Run(os.Args)
}
