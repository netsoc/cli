package util

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/netsoc/cli/pkg/config"
	iam "github.com/netsoc/iam/client"
)

// CmdFactory provides methods to obtain commonly used structures
type CmdFactory struct {
	Config    func() (*config.Config, error)
	Claims    func() (*UserClaims, error)
	IAMClient func() (*iam.APIClient, error)
}

// NewDefaultCmdFactory creates a new command factory
func NewDefaultCmdFactory(configFlag, debugFlag *pflag.Flag) *CmdFactory {
	var cachedConfig *config.Config
	configFunc := func() (*config.Config, error) {
		if cachedConfig != nil {
			return cachedConfig, nil
		}

		config.SetDefaults()

		configFile := os.Getenv("NETSOC_CONFIG")
		if configFile == "" || configFlag.Changed {
			configFile = configFlag.Value.String()
		}
		viper.SetConfigFile(configFile)

		// Config from environment
		viper.SetEnvPrefix("netsoc")
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.AutomaticEnv()

		// Config from flags
		viper.BindPFlag("debug", debugFlag)

		err := viper.ReadInConfig()

		IsDebug = viper.GetBool("debug")
		if IsDebug {
			log.SetFlags(log.Ltime | log.Lmicroseconds | log.Llongfile)
		} else {
			log.SetFlags(0)
		}

		if err == nil {
			Debugf("Loaded config file: %v", viper.ConfigFileUsed())
		}

		if err := viper.Unmarshal(&cachedConfig); err != nil {
			return nil, fmt.Errorf("failed to parse configuration: %w", err)
		}

		return cachedConfig, nil
	}

	return &CmdFactory{
		Config: configFunc,
		Claims: func() (*UserClaims, error) {
			c, err := configFunc()
			if err != nil {
				return nil, fmt.Errorf("failed to load config: %w", err)
			}

			t, _, err := jwt.NewParser().ParseUnverified(c.Token, &UserClaims{})
			if err != nil {
				return nil, err
			}

			return t.Claims.(*UserClaims), nil
		},
		IAMClient: func() (*iam.APIClient, error) {
			c, err := configFunc()
			if err != nil {
				return nil, fmt.Errorf("failed to load config: %w", err)
			}

			cfg := iam.NewConfiguration()
			cfg.BasePath = c.URLs.IAM
			if c.AllowInsecure {
				cfg.HTTPClient = &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: true,
						},
					},
				}
			}

			return iam.NewAPIClient(cfg), nil
		},
	}
}
