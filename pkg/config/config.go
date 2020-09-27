package config

import (
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// DecoderOptions enables necessary mapstructure decode hook functions
func DecoderOptions(config *mapstructure.DecoderConfig) {
	config.ErrorUnused = true
	config.DecodeHook = mapstructure.ComposeDecodeHookFunc(
		config.DecodeHook,
		mapstructure.StringToTimeHookFunc(time.RFC3339Nano),
	)
}

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

	LastUpdateCheck time.Time `mapstructure:"last_update_check"`
}
