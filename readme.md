# Aggregator

This is a log aggregator, there are many like it, but this one is mine. Written in Go (golang) this simple program aggregates text files - while keeping up whith them as they are being updated (very much like an aggregator of multiple files, where all of them are bening monitored by tail -f). It is useful while developing, espically if you work on a lot of different servies (at once) that communicate with each other. 

## Command line flags
### Please note: this program supports 2 forms of configuration:
You can either use the -config flag, which points to a config.yaml with your configuration outlined at that location, or you can use same configuration by using the verity of other command line flags. Please note you should not mix the config flag with other command line flags. This multi config setup was implemented to prevent issues when trying to aggrage a large list of log files - as managing a config file is much easier that a long list of flags. 
| Command line flag | Value Type | Description | Example |
| ------ | ------ | ------ | ----- |
| -config | String | In the event you do not want to use the Command line flags, you can use a config file to list where your log files to aggregate are located. This flag tells aggregator where to find the config.yaml file. The string you provide is the path to where your config.yaml file is located| ./aggregator -config=Path/to/config
| -logFiles | Array of strings - seperated by comma | Provide a comma seperated list of strings, this tells aggregator the locations of the log files you want to aggregate. | ./aggregator -logFiles=/path/one.txt,/path/two.txt
| -clear | boolean | This will clear the log files you are aggregating upon termination of this program. This is good for development, use with caution.| ./aggregator -clear=true (default is false)


## Built with the following open source golang packages

* [Viper](https://github.com/spf13/viper) 
* [Cli](https://github.com/urfave/cli) 
* [Tail](https://github.com/hpcloud/tail) 

## Getting Started
If you are using the config file option, see the section titled "Config file setup" 

This project ships with the binary, so after cloning or downloading the codebase, you can cd into the directory where the binary is located and run the following:

If you are using a configuration file (and the config is located in the same directory as the binary):

```sh
$ ./aggregator -config=./
```

Otherwise:
```sh
$ ./aggregator -logFiles=/path/one.txt,/path/two.txt
```

## Config File Setup
Within this project is a file called config.yaml.sample, simply copy file - editing it to your liking to a file named config.yaml

Again, in order for the project to know you wish to this config.yaml file, you must use the -config flag - which is the path to the DIRECTORY that contains the config.yaml file. 

## License:
MIT