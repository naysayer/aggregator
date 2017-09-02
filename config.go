package main

import (
	"errors"
	"strings"

	"github.com/spf13/viper"
)

var (
	errDualConfig        = errors.New("you can only choose to use the config command line flag, or the location + close flag, you cannot use both")
	errMustProvideConfig = errors.New("neither a command line flag nor a config file contained the locations of the files you wish to aggregate")
)

// useConfigFile should be passed the command like flag that has the path to
// the configiration file. If that string is present the program will know that
// it should eventuall be looking for a config.yaml file.
func useConfigFile(config string) bool {
	if config == "" {
		return false
	}
	return true
}

// readConfigFile addersses leverage viper which is sort of a global var sort
// of thing which is why it is not an argument. If we choose to use the config
// file over command line flags, this function will then attempt to look for
// and read the config file.
func readConfigFile(useConfigFile bool, configPath string) error {
	if useConfigFile {
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
		viper.AddConfigPath(configPath)
		err := viper.ReadInConfig() // Find and read the config file
		if err != nil {             // Handle errors reading the config file
			return err
		}
	}
	return nil
}

// Please note this depends heavily on global vars related to configuration.
// obtainLogLocations returns a slice of strings and error. This program
// supports 2 forms of config, command line and a config.yml file. This attempts
// to read in the list of desired log file locations from either source. In the
// event that both a present, the command line flag takes presidences over
// the config file.

// obtainLogLocations When starting the app they a user can do 1 of 2 things,
// then can do config, entirely from the command line, or by only using the
// config option.) Using the locations flag requires additional formating,
// so this function, returns a slice of strings with the correctly formatted
// locations of log files the user wishes to aggregate.
// The locs var that is passed in should be the raw version of the command
// line flag that supplies log file locations.
func obtainLogLocations(locs []string, useConfig bool) ([]string, error) {
	if len(locs) > 0 && Config != "" {
		return nil, errDualConfig
	}
	if len(locs) < 1 && Config == "" {
		return nil, errMustProvideConfig
	}

	if useConfig {
		locs := viper.GetStringSlice("Locations")
		return locs, nil
	}

	return strings.Split(locs[0], ","), nil
}
