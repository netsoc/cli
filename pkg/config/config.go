package config

import "github.com/spf13/viper"

// SetDefaults sets config defaults
func SetDefaults() {
	viper.SetDefault("debug", false)
	viper.SetDefault("token", "")
	viper.SetDefault("allow_insecure", false)

	viper.SetDefault("urls.iam", "https://iam.netsoc.ie/v1")
}

// Config represents the Netsoc CLI config
type Config struct {
	Debug         bool
	Token         string
	AllowInsecure bool `mapstructure:"allow_insecure"`

	URLs struct {
		IAM string
	}
}
