package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/echocrow/Mouser/pkg/bootstrap"
	"github.com/echocrow/Mouser/pkg/config"
	"github.com/echocrow/Mouser/pkg/log"
)

// version is the version of this app set at build-time.
var version = "0.0.0-dev"

func main() {
	var confPath string
	defConfPath, defConfPathErr := defaultConfigPath()
	flag.StringVar(&confPath, "config", "", fmt.Sprintf(
		"The path to the config file. Defaults to %s.",
		defConfPath,
	))

	var verbose bool
	flag.BoolVar(&verbose, "v", false, "Verbose")

	var getVersion bool
	flag.BoolVar(&getVersion, "version", false, "Print the app version & exit.")

	flag.Parse()

	if getVersion {
		exitMessage(0, fmt.Sprint("mouser ", version))
	}

	if confPath == "" {
		if defConfPathErr != nil {
			abort(2, defConfPathErr)
		}
		confPath = defConfPath
	}

	var logger log.Logger
	if verbose {
		logger = log.New("Mouser")
		logger.Printf("Version=%s", version)
		logger.Printf("ConfigPath=%s", confPath)
	}

	conf, err := parseConfig(confPath)
	if err != nil {
		abort(2, err)
	}

	if verbose {
		conf.Settings.Debug = true
	}

	run, stop, err := bootstrap.Bootstrap(conf)
	if err != nil {
		abort(1, err)
	}

	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		if err := stop(); err != nil {
			abort(1, err)
		}
	}()

	if err := run(); err != nil {
		abort(1, err)
	}
}

func defaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	confPath := filepath.Join(homeDir, ".config", "mouser", "config.yml")
	return confPath, nil
}

func abort(code int, err error) {
	exitMessage(code, fmt.Sprint("[Error] ", err))
}

func exitMessage(code int, msg string) {
	flag.CommandLine.SetOutput(nil)
	fmt.Fprintln(flag.CommandLine.Output(), msg)
	os.Exit(code)
}

func parseConfig(confPath string) (config.Config, error) {
	var err error

	if confPath == "" {
		return config.Config{}, fmt.Errorf("config path is required")
	}

	confPath, err = filepath.Abs(confPath)
	if err != nil {
		return config.Config{}, fmt.Errorf("error parsing config path: %s", err)
	}

	confFile, err := ioutil.ReadFile(confPath)
	if err != nil {
		return config.Config{}, fmt.Errorf("error reading config file: %s", err)
	}

	conf, err := config.ParseYAML(confFile)
	if err != nil {
		return config.Config{}, fmt.Errorf("error parseing config file: %s", err)
	}

	return conf, nil
}
