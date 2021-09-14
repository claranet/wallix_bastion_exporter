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

type Config struct {
	ListenAddress  string `mapstructure:"listen-address"`
	TelemetryPath  string `mapstructure:"telemetry-path"`
	ScrapeURI      string `mapstructure:"scrape-uri"`
	SkipVerify     bool   `mapstructure:"skip-verify"`
	Timeout        int    `mapstructure:"timeout"`
	WallixUsername string `mapstructure:"wallix-username"`
	WallixPassword string `mapstructure:"wallix-password"`
}

func LoadConfig(path string) (config Config, err error) {
	if err := SetFlags(); err != nil {
		return config, err
	}

	// Set variables from config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok { //nolint:errorlint
			return config, err
		}
	}

	// Set variables from environment
	viper.AutomaticEnv()
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)

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

func SetFlags() (err error) {
	// Set flags and default variables
	helpFlag := pflag.BoolP("help", "h", false, "help message")
	pflag.String("listen-address", ":9191", "Address to listen on for web interface and telemetry")
	pflag.String("telemetry-path", "/metrics", "Path under which to expose metrics")
	pflag.StringP("scrape-uri", "w", "https://127.0.0.1/api", "Path under which to expose metrics")
	pflag.StringP("wallix-username", "u", "", "The username used for authentication to request Wallix Bastion API")
	pflag.StringP("wallix-password", "p", "", "The password used for authentication to request Wallix Bastion API")

	pflag.BoolP("skip-verify", "s", false, "Flag that disables TLS certificate verification for the scrape URI")
	pflag.IntP("timeout", "t", defaultTimeout, "Timeout for trying to get stats from Wallix Bastion")
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
