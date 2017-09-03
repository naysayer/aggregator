package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/hpcloud/tail"
	"github.com/urfave/cli"
)

var (
	flagLocations         cli.StringSlice
	flagConfig            string
	flagClearFilesOnClose bool
	flagLogToFile         string

	closingChannel = make(chan os.Signal, 1)
	tailConfig     = tail.Config{Follow: true, MustExist: true}
)

func main() {
	app := cli.NewApp()
	app.Name = "aggregator"
	app.Usage = "Aggregates logs into a single file. Dont't hate, aggregate."
	app.Version = "0.0.2"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config",
			Usage:       "In the event you do not want to use the Command line flags, you can use a config file to list where your log files to aggregate are located. This flag tells aggregator where to find the config.yaml file. The string you provide is the path to where your config.yaml file is located.",
			Destination: &flagConfig,
		},
		cli.StringSliceFlag{
			Name:  "logFiles",
			Usage: "Provide a comma seperated list of strings, this tells aggregator the locations of the log files you want to aggregate.",
			Value: &flagLocations,
		},
		cli.BoolFlag{
			Name:        "clear",
			Usage:       "This will clear the log files you are aggregating upon termination of this program. This is good for development, use with caution.",
			Destination: &flagClearFilesOnClose,
		},
		cli.StringFlag{
			Name:        "logToFile",
			Usage:       "In the event you want to have this program aggreagte logs into a single file rather than stream to the terminal, use this option to pass a string to the location where you want to log, if a file is not present at that location, we will attempt to create one.",
			Destination: &flagLogToFile,
		},
	}
	app.Action = func(c *cli.Context) error {
		appConfig, err := newConfigFromFlags()
		if err != nil {
			return err
		}

		signal.Notify(closingChannel, os.Interrupt)
		appConfig.setLogOutput()
		log.Println("Aggregator started...")

		tails, err := obtainTails(appConfig.locations)
		if err != nil {
			return err
		}

		for _, t := range tails {
			go tailFile(t)
		}

		<-closingChannel

		log.Println("Shutting down...")
		closeTails(tails)
		appConfig.shutdown()

		return nil
	}
	app.Run(os.Args)
}

// obtainTails provided a slice of strings (that is used to represent the
// locations of log files you wish to aggregate). Will return a matching
// tail.Tail slice where each value in the slice is mapped from the argumented
// locations.
func obtainTails(locations []string) ([]*tail.Tail, error) {
	var tails []*tail.Tail
	for _, l := range locations {
		t, err := tail.TailFile(l, tailConfig)
		if err != nil {
			log.Println("There was an issue while attmpting to aggregate log file located at:", l)
			return nil, err
		}
		tails = append(tails, t)
	}
	return tails, nil
}

// tailFile depends on the tail package, it essentially polls the file it
// is supposed to log, writing the data to the corresponding log location.
func tailFile(tail *tail.Tail) {
	for line := range tail.Lines {
		log.Println(line.Text)
	}
}

// closeTails should be called after the program is interrupted. Upon recieving
// this signal it will range over the tails argument. Tails represents the
// *tail.Tail pointer that is created for each and every file being tailed.
func closeTails(tails []*tail.Tail) {
	log.Println("Attempting to close aggregator")
	for _, t := range tails {
		t.Stop()
		t.Cleanup()
	}
}
