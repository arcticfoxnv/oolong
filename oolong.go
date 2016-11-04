package main

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/arcticfoxnv/oolong/state"
	"github.com/arcticfoxnv/oolong/wirelesstag"
	"github.com/urfave/cli"
)

const Version = "0.0.3"
const oolongStateFile = "state.json"

type Config struct {
	OAuth        OAuthConfig
	HTTP         HTTPConfig
	PollInterval int      `toml:"poll_interval"`
	QueryStats   []string `toml:"query_stats"`
	ConvertToF   bool     `toml:"convert_to_f"`
	OpenTSDB     OpenTSDBConfig
}

type HTTPConfig struct {
	Port int
}

// These are found at https://mytaglist.com/eth/oauth2_apps.html
type OAuthConfig struct {
	ID     string
	Secret string
}

type OpenTSDBConfig struct {
	Host          string
	Port          int
	MetricsPrefix string `toml:"metrics_prefix"`
}

func ReadConfigFile(filename string) *Config {
	config := new(Config)
	tomlData, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read config file: %s\n", err.Error())
	}
	_, err = toml.Decode(string(tomlData), config)
	if err != nil {
		log.Fatalf("Failed to parse config file: %s\n", err.Error())
	}
	return config
}

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
	// Read config file
	config := ReadConfigFile(c.GlobalString("config"))

	// Initialize data storage client
	tsdbClient := NewOpenTSDBClient(config.OpenTSDB.Host, config.OpenTSDB.Port, config.OpenTSDB.MetricsPrefix)

	// Try to load state from file
	state, err := state.NewStateFromFile(oolongStateFile)
	if err != nil {
		log.Fatalf("Unable to restore state: %s\n", err.Error())
	}

	// Use token from state file to initialize the wireless tag client
	tagClient := wirelesstag.NewClient(state.GetAccessToken())

	// Retrieve stats from cloud and push to data storage
	StatsFetcher(config, state, tagClient, tsdbClient)

	return nil
}

func cmdBackfill(c *cli.Context) error {
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

	// Try to load state from file
	state, err := state.NewStateFromFile(oolongStateFile)
	if err != nil {
		log.Fatalf("Unable to restore state: %s\n", err.Error())
	}

	// Use token from state file to initialize the wireless tag client
	tagClient := wirelesstag.NewClient(state.GetAccessToken())

	// Retrieve stats from cloud and push to data storage
	Backfill(config, state, tagClient, tsdbClient, date)

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
