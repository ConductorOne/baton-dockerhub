package main

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/spf13/cobra"
)

// config defines the external configuration required for the connector to run.
type config struct {
	cli.BaseConfig `mapstructure:",squash"` // Puts the base config options in the same place as the connector options

	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
	Orgs     []string `mapstructure:"orgs"`
}

// validateConfig is run after the configuration is loaded, and should return an error if it isn't valid.
func validateConfig(ctx context.Context, cfg *config) error {
	if cfg.Username == "" || cfg.Password == "" {
		return fmt.Errorf("username and password must be provided")
	}

	if len(cfg.Orgs) == 0 {
		return fmt.Errorf("at least one organization must be provided")
	}

	return nil
}

// cmdFlags sets the cmdFlags required for the connector.
func cmdFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("username", "", "The DockerHub username used to connect to the DockerHub API. ($BATON_USERNAME)")
	cmd.PersistentFlags().String("password", "", "The DockerHub password used to connect to the DockerHub API. ($BATON_PASSWORD)")
	cmd.PersistentFlags().StringSlice("orgs", nil, "Limit syncing to specific organizations by providing organization slugs. ($BATON_ORGS)")
}
