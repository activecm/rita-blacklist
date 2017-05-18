package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"reflect"

	"gopkg.in/yaml.v2"
)

var VERSION = "undefined"

type (
	SystemConfig struct {
		UseCustomBlacklist  bool              		`yaml:"UseCustomBlacklist"`
    CustomBlacklistCfg  []CustomBlacklistCfg	`yaml:"CustomBlacklist"`
		Version							string
	}

	CustomBlacklistCfg struct {
		Location  string `yaml:"Location"`
		ValTime 	string `yaml:"ValTime"`
		Name    	string `yaml:"Name"`
	}
)

// GetConfig retrieves a configuration in order of precedence
func GetConfig() (*SystemConfig, bool) {
	// Get the user's homedir
	user, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get user info: %s\n", err.Error())
	} else {

		conf, ok := loadSystemConfig(user.HomeDir + "/.rita/blacklist.yaml")
		if ok {
			return conf, ok
		}
	}

	// If none of the other configs have worked, go for the global config
	return loadSystemConfig("/etc/rita/blacklist.yaml")
}

// loadSystemConfig attempts to parse a config file
func loadSystemConfig(cfgPath string) (*SystemConfig, bool) {
	var config = new(SystemConfig)

	config.Version = VERSION

	if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
		cfgFile, err := ioutil.ReadFile(cfgPath)
		if err != nil {
			return config, false
		}
		err = yaml.Unmarshal(cfgFile, config)

		// expand env variables, config is a pointer
		// so we have to call elem on the reflect value
		expandConfig(reflect.ValueOf(config).Elem())

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read config: %s\n", err.Error())
			return config, false
		}
		return config, true
	}
	return config, false
}

// expandConfig expands environment variables in config strings
func expandConfig(reflected reflect.Value) {
	for i := 0; i < reflected.NumField(); i++ {
		f := reflected.Field(i)
		// process sub configs
		if f.Kind() == reflect.Struct {
			expandConfig(f)
		} else if f.Kind() == reflect.String {
			f.SetString(os.ExpandEnv(f.String()))
		}
	}
}
