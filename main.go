package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/hpcloud/tail"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

var (
	locations         cli.StringSlice
	Config            string
	ClearFilesOnClose bool
	wg                sync.WaitGroup
	Tails             []*tail.Tail
)

func main() {
	app := cli.NewApp()
	app.Name = "aggregator"
	app.Usage = "Aggregates logs into a single file. Dont't hate, aggregate."
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config",
			Usage:       "In the event you do not want to use the Command line flags, you can use a config file to list where your log files to aggregate are located. This flag tells aggregator where to find the config.yaml file. The string you provide is the path to where your config.yaml file is located.",
			Destination: &Config,
		},
		cli.StringSliceFlag{
			Name:  "logFiles",
			Usage: "Provide a comma seperated list of strings, this tells aggregator the locations of the log files you want to aggregate.",
			Value: &locations,
		},
		cli.BoolFlag{
			Name:        "clear",
			Usage:       "This will clear the log files you are aggregating upon termination of this program. This is good for development, use with caution.",
			Destination: &ClearFilesOnClose,
		},
	}
	app.Action = func(c *cli.Context) error {
		useConfigFile := useConfigFile(Config)
		readConfigFile(useConfigFile, Config)

		logLocations, err := obtainLogLocations(useConfigFile)
		if err != nil {
			return err
		}

		if useConfigFile {
			ClearFilesOnClose = viper.GetBool("ClearLogsOnClose")
		}

		go handleClosing()
		aggregateLogs(logLocations) // Spawns 1 go routine per call

		log.Println("should clear log file", ClearFilesOnClose)
		if ClearFilesOnClose {
			clearFiles(logLocations)
		}
		log.Println("Shutting down")
		return nil
	}

	app.Run(os.Args)
}

func handleClosing() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			log.Println("Attempting cleanup")
			for _, t := range Tails {
				t.Stop()
				t.Cleanup()
			}
			wg.Done()
		}
	}()
}

func aggregateLogs(locations []string) {
	wg.Add(1)
	for _, l := range locations {
		log.Println("Attempting to integrate log file located at", l)

		t, err := tail.TailFile(l, tail.Config{Follow: true, MustExist: true})
		if err != nil {
			log.Println("There was an issue while attmpting to aggregate log file located at", l)
			log.Fatal(err)
		}

		go tailFile(t)
		Tails = append(Tails, t)
	}
	wg.Wait()
}

func tailFile(tail *tail.Tail) {
	for line := range tail.Lines {
		log.Println(line.Text)
	}
}

func clearFiles(locations []string) {
	for _, l := range locations {
		log.Println("Clearning file at location:", l)

		err := ioutil.WriteFile(l, []byte{}, 0666)
		if err != nil {
			log.Println("Unable to clear file at location:", l)
			log.Println(err)
		}
	}
}
