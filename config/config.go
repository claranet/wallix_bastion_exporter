package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultTimeout = 10
)

// All configuration available for the user.
type Config struct {
	ListenAddress  string `mapstructure:"listen-address"`
	TelemetryPath  string `mapstructure:"telemetry-path"`
	ScrapeURI      string `mapstructure:"scrape-uri"`
	SkipVerify     bool   `mapstructure:"skip-verify"`
	Timeout        int    `mapstructure:"timeout"`
	WallixUsername string `mapstructure:"wallix-username"`
	WallixPassword string `mapstructure:"wallix-password"`
}

// Entry point function to load the configuration with the following precedence order:
// - flag
// - env var
// - config file
// The config file is optional.
func LoadConfig(path string) (config Config, err error) {
	if err := SetFlags(); err != nil {
		return config, err
	}

	// Set variables from config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	if err := viper.ReadInConfig(); err != nil {
		// Config file is optional, so ignore error if it does not exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok { //nolint:errorlint
			return config, err
		}
	}

	// Set variables from environment
	viper.AutomaticEnv()

	// Required to bind flag with env var
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)

	// Check mandatory parameters
	if !viper.IsSet("wallix-username") {
		return config, fmt.Errorf("wallix-username is a mandatory input")
	}
	if !viper.IsSet("wallix-password") {
		return config, fmt.Errorf("wallix-password is a mandatory input")
	}

	if err := viper.Unmarshal(&config); err != nil {
		return config, err
	}

	return config, nil
}

// Set flags and default variables.
func SetFlags() (err error) {
	helpFlag := pflag.BoolP("help", "h", false, "help message")
	pflag.String("listen-address", ":9191", "Address to listen on for web interface and telemetry")
	pflag.String("telemetry-path", "/metrics", "Path under which to expose metrics")
	pflag.StringP("scrape-uri", "w", "https://127.0.0.1/api", "URI on which to scrape Wallix Bastion API")
	pflag.StringP("wallix-username", "u", "", "The username used for authentication to request Wallix Bastion API")
	pflag.StringP("wallix-password", "p", "", "The password used for authentication to request Wallix Bastion API")

	pflag.BoolP("skip-verify", "s", false, "Flag that disables TLS certificate verification for the scrape URI")
	pflag.IntP("timeout", "t", defaultTimeout, "Timeout in seconds for requests to Wallix Bastion API")
	pflag.Parse()

	// Bind to viper all other flags
	if err := viper.BindPFlag("listen-address", pflag.Lookup("listen-address")); err != nil {
		return err
	}
	if err := viper.BindPFlag("telemetry-path", pflag.Lookup("telemetry-path")); err != nil {
		return err
	}
	if err := viper.BindPFlag("scrape-uri", pflag.Lookup("scrape-uri")); err != nil {
		return err
	}
	if err := viper.BindPFlag("wallix-username", pflag.Lookup("wallix-username")); err != nil {
		return err
	}
	if err := viper.BindPFlag("wallix-password", pflag.Lookup("wallix-password")); err != nil {
		return err
	}
	if err := viper.BindPFlag("skip-verify", pflag.Lookup("skip-verify")); err != nil {
		return err
	}
	if err := viper.BindPFlag("timeout", pflag.Lookup("timeout")); err != nil {
		return err
	}

	// Handle special help flag not binded to viper
	if *helpFlag {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		pflag.PrintDefaults()
		os.Exit(0)
	}

	return nil
}
