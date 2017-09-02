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
	LogToFile         string
	LogFile           *os.File
	wg                sync.WaitGroup
	closingChannel    = make(chan os.Signal, 1)
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
		cli.StringFlag{
			Name:        "logToFile",
			Usage:       "In the event you want to have this program aggreagte logs into a single file rather than stream to the terminal, use this option to pass a string to the location where you want to log, if a file is not present at that location, we will attempt to create one.",
			Destination: &LogToFile,
		},
	}
	app.Action = func(c *cli.Context) error {
		signal.Notify(closingChannel, os.Interrupt)

		useConfigFile := useConfigFile(Config)
		readConfigFile(useConfigFile, Config)

		if useConfigFile {
			ClearFilesOnClose = viper.GetBool("ClearLogsOnClose")
			LogToFile = viper.GetString("LogToFile")
		}
		if LogToFile != "" {
			LogFile, err := os.OpenFile(LogToFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

			if err != nil {
				log.Fatal(err)
			}
			log.SetOutput(LogFile)
		}

		logLocations, err := obtainLogLocations(locations, useConfigFile)
		if err != nil {
			return err
		}

		tails, err := ObtainTails(logLocations)
		if err != nil {
			log.Fatal(err)
		}

		for _, t := range tails {
			go tailFile(t)
		}

		<-closingChannel

		HandleClosing(tails)

		if ClearFilesOnClose {
			ClearFiles(logLocations)
		}

		if LogToFile != "" {
			LogFile.Close()
		}

		log.Println("Shutting down")
		return nil
	}
	app.Run(os.Args)
}

func tailFile(tail *tail.Tail) {
	for line := range tail.Lines {
		log.Println(line.Text)
	}
}

// HandleClosing should be called after the program is interrupted. Upon recieving this
// signal it will range over the tails argument. Tails represents the *tail.Tail
// value / pointer that is created for each and every file being tailed.
func HandleClosing(tails []*tail.Tail) {
	log.Println("Attempting to close aggregator")
	for _, t := range tails {
		t.Stop()
		t.Cleanup()
	}
}

// ObtainTails provided a slice of strings (that is used to represent the
// locations of log files you wish to aggregate). Will return a matching
// tail.Tail slice where each value in the slice is mapped from the argumented
// locations.
func ObtainTails(locations []string) ([]*tail.Tail, error) {
	var tails []*tail.Tail
	for _, l := range locations {
		t, err := tail.TailFile(l, tail.Config{Follow: true, MustExist: true})
		if err != nil {
			log.Println("There was an issue while attmpting to aggregate log file located at:", l)
			return tails, err
		}
		tails = append(tails, t)
	}
	return tails, nil
}

// ClearFiles as the name suggests will locate all of the log files within
// the locations argument - and write all of the bytes to be empty. This
// is intended to be used per the users settings; and can be handy in development
func ClearFiles(locations []string) {
	for _, l := range locations {
		log.Println("Clearning file at location:", l)

		err := ioutil.WriteFile(l, []byte{}, 0666)
		if err != nil {
			log.Println("Unable to clear file at location:", l)
			log.Println(err)
		}
	}
}
