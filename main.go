package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudflare/cloudflare-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Before will setup the logger.
func Before(c *cli.Context) error {
	if c.Bool("json") {
		// Configure the JSON logger if enabled.
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	if c.Bool("debug") {
		// Set the debug log level if enabled.
		logrus.SetLevel(logrus.DebugLevel)
	}

	return nil
}

// Action will perform the update operation.
func Action(c *cli.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var api *cloudflare.API

	var err error

	if c.String("token") != "" {
		api, err = cloudflare.NewWithAPIToken(c.String("token"))
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}
	} else if c.String("key") != "" && c.String("email") != "" {
		api, err = cloudflare.New(c.String("key"), c.String("email"))
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}
	} else {
		return cli.Exit("either --key and --email or --token must be defined", 1)
	}

	if err := UpdateDomain(ctx, api, c.String("domain"), c.String("ipendpoint")); err != nil {
		return cli.Exit(err.Error(), 1)
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "cloudflare-ddns"
	app.Version = fmt.Sprintf("%v, commit %v, built at %v", version, commit, date)
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "token",
			EnvVars: []string{"CF_API_TOKEN"},
			Usage:   "The API Token that has the Zone.DNS permission for the specific zone.",
		},
		&cli.StringFlag{
			Name:    "key",
			EnvVars: []string{"CF_API_KEY"},
			Usage:   "The Global (not CA) Cloudflare API Key generated on the \"My Account\" page.",
		},
		&cli.StringFlag{
			Name:    "email",
			EnvVars: []string{"CF_API_EMAIL"},
			Usage:   "Email address associated with your Cloudflare account.",
		},
		&cli.StringFlag{
			Name:     "domain",
			Required: true,
			EnvVars:  []string{"CF_DOMAIN"},
			Usage:    "Comma separated domain names that should be updated. (i.e. mypage.example.com OR example.com)",
		},
		&cli.StringFlag{
			Name:    "ipendpoint",
			Value:   "https://api.ipify.org/",
			EnvVars: []string{"CF_IP_ENDPOINT"},
			Usage:   "Alternative ip address service endpoint.",
		},
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "Enables debug logging.",
		},
		&cli.BoolFlag{
			Name:  "json",
			Usage: "Enables JSON output for the logging.",
		},
	}
	app.Before = Before
	app.Action = Action

	if err := app.Run(os.Args); err != nil {
		logrus.WithError(err).Fatal()
	}
}
