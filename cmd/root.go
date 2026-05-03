// Package cmd implements the Cobra CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chatchomphu1000/go-starter/internal/config"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

var (
	cfgFile string
	cfg     *config.Config
	log     logger.Logger
)

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use:   "go-starter",
	Short: "Go Starter API server",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load .env file in development (optional, ignore if not present).
		_ = godotenv.Load()

		// Override APP_ENV if --env flag is provided.
		if env := cmd.Flag("env"); env != nil && env.Changed {
			viper.Set("app.env", env.Value.String())
		}

		var err error
		if cfgFile != "" {
			cfg, err = config.LoadFromFile(cfgFile)
		} else {
			cfg, err = config.Load()
		}
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		log, err = logger.NewLogger(logger.LoggerConfig{
			Level:       cfg.Logger.Level,
			Format:      cfg.Logger.Format,
			Development: cfg.Logger.Development,
		})
		if err != nil {
			return fmt.Errorf("failed to create logger: %w", err)
		}

		logger.SetGlobal(log)

		return nil
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.PersistentFlags().String("env", "", "override APP_ENV (development|staging|production)")
}
