package main

import (
	"log"
	"os"
	"time"

	"github.com/arcticfoxnv/oolong/state"
	"github.com/arcticfoxnv/oolong/wirelesstag"
	"github.com/urfave/cli"
)

const Version = "0.0.4"

func cmdHTTPServer(c *cli.Context) error {
	// Read config file
	config := ReadConfigFile(c.GlobalString("config"))

	// Channel to signal setup is complete.
	setupDone := make(chan int)
	startChan := make(chan string)

	// HTTP server is only needed to do OAuth fun.  The resulting token
	// is saved to the state file for reuse.
	log.Printf("Starting HTTP server...\n")
	go StartHTTPServer(config, startChan, setupDone)

	// Wait for the server to initialize
	startUrl := <-startChan

	// Print the URL for the user to visit to start the OAuth dance.
	log.Printf("Oolong ready!  Go to %s to begin.", startUrl)

	// Wait for setup to complete
	<-setupDone

	return nil
}

func cmdPoller(c *cli.Context) error {
	var st state.State
	var err error

	// Read config file
	config := ReadConfigFile(c.GlobalString("config"))

	// Initialize data storage client
	tsdbClient := NewOpenTSDBClient(config.OpenTSDB.Host, config.OpenTSDB.Port, config.OpenTSDB.MetricsPrefix)

	// Try to load state from backend
	switch config.Backend {
	case "file":
		st, err = state.NewStateFromFile(config.File.Filename)
	case "redis":
		st, err = state.NewStateFromRedis(config.Redis.Host, config.Redis.Port, config.Redis.Key)
	}
	if err != nil {
		log.Fatalf("Unable to restore state: %s\n", err.Error())
	}

	// Use token from state file to initialize the wireless tag client
	tagClient := wirelesstag.NewClient(st.GetAccessToken())

	// Retrieve stats from cloud and push to data storage
	StatsFetcher(config, st, tagClient, tsdbClient)

	return nil
}

func cmdBackfill(c *cli.Context) error {
	var st state.State
	var err error
	if c.String("date") == "" {
		log.Fatalln("--date is required")
	}
	// Read config file
	config := ReadConfigFile(c.GlobalString("config"))
	date, err := time.Parse("2006-01-02", c.String("date"))
	if err != nil {
		log.Fatalf("Failed to parse date: %s\n", err.Error())
	}

	// Initialize data storage client
	tsdbClient := NewOpenTSDBClient(config.OpenTSDB.Host, config.OpenTSDB.Port, config.OpenTSDB.MetricsPrefix)

	// Try to load state from backend
	switch config.Backend {
	case "file":
		st, err = state.NewStateFromFile(config.File.Filename)
	case "redis":
		st, err = state.NewStateFromRedis(config.Redis.Host, config.Redis.Port, config.Redis.Key)
	}
	if err != nil {
		log.Fatalf("Unable to restore state: %s\n", err.Error())
	}

	// Use token from state file to initialize the wireless tag client
	tagClient := wirelesstag.NewClient(st.GetAccessToken())

	// Retrieve stats from cloud and push to data storage
	Backfill(config, st, tagClient, tsdbClient, date)

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "oolong"
	app.Usage = "Polls wirelesstag.net cloud data"
	app.Version = Version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: "oolong.toml",
			Usage: "Specifies path to config file",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "init",
			Usage:  "Run the HTTP server to setup OAuth",
			Action: cmdHTTPServer,
		},
		{
			Name:   "run",
			Usage:  "Run the poller to periodically retrieve data from wirelesstag.net",
			Action: cmdPoller,
		},
		{
			Name:   "backfill",
			Usage:  "Retrieve a specific day and exit",
			Action: cmdBackfill,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "date",
					Usage: "Specify the date to retrieve (format: YYYY-MM-DD)",
				},
			},
		},
	}

	app.Run(os.Args)
}
