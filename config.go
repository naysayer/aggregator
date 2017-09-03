package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

var (
	errDualConfig        = errors.New("you can only choose to use the config command line flag, or the location + close flag, you cannot use both")
	errMustProvideConfig = errors.New("neither a command line flag nor a config file contained the locations of the files you wish to aggregate")
)

type appConfig struct {
	config            string
	locations         []string
	clearFilesOnClose bool
	logToFile         string
	logFile           *os.File
}

// newConfigFromFlags reads in the vars that are casted from command line flags
// and returns a new ponter to an appConfig.
func newConfigFromFlags() (*appConfig, error) {
	var ac = &appConfig{}
	ac.config = flagConfig
	if len(flagLocations) > 0 && ac.config != "" {
		return nil, errDualConfig
	}
	if len(flagLocations) < 1 && ac.config == "" {
		return nil, errMustProvideConfig
	}

	if ac.config != "" {
		err := readConfigFile(ac.config)
		if err != nil {
			return nil, err
		}
		ac.clearFilesOnClose = viper.GetBool("ClearLogsOnClose")
		ac.logToFile = viper.GetString("LogToFile")
		ac.locations = viper.GetStringSlice("Locations")
		return ac, nil
	}
	ac.clearFilesOnClose = flagClearFilesOnClose
	ac.logToFile = flagLogToFile
	ac.locations = strings.Split(flagLocations[0], ",")
	return ac, nil
}

// readConfigFile addersses leverage viper which is sort of a global var sort
// of thing which is why it is not an argument. If we choose to use the config
// file over command line flags, this function will then attempt to look for
// and read the config file.
func readConfigFile(configPath string) error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	return nil
}

// setLogOutput helps determine where we choose to log. Think 12 factor apps.
// as in development on may want to stream to the console. Where as a log file
// may be better elswhere. If a file is used for logging then said file
// is casted to the logFile attribute of the calling appconfig.
func (a *appConfig) setLogOutput() {
	if a.logToFile != "" {
		LogFile, err := os.OpenFile(a.logToFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

		if err != nil {
			log.Fatal(err)
		}

		log.SetOutput(LogFile)
		a.logFile = LogFile
	}
}

// clearFiles as the name suggests will locate all of the log files within
// the locations argument - and write all of the bytes to be empty. This is
// intended to be used per the users settings and can be handy in development.
// chose not to return the error here has it is only called on the control c /
// upon the program closing.
func (a *appConfig) clearFiles() {
	for _, l := range a.locations {
		err := ioutil.WriteFile(l, []byte{}, 0666)
		if err != nil {
			log.Println("Unable to clear file at location:", l)
			log.Println(err)
		}
	}
}

// shutdown per the values witin the calling appConfig will close the logging
// file or attempt to clear the files the app is polling from.
func (a *appConfig) shutdown() {
	if a.clearFilesOnClose {
		log.Println("Attempting to clear log files...")
		a.clearFiles()
	}

	if a.logToFile != "" {
		a.logFile.Close()
	}
}
