package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/chatchomphu1000/go-starter/internal/adapters/outbound/mongodb"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration commands",
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Run all pending migrations",
	RunE: func(_ *cobra.Command, _ []string) error {
		runner, err := newMigrationRunner()
		if err != nil {
			return err
		}
		defer runner.Close()
		return runner.Up()
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback migrations",
	RunE: func(cmd *cobra.Command, _ []string) error {
		runner, err := newMigrationRunner()
		if err != nil {
			return err
		}
		defer runner.Close()

		steps, _ := cmd.Flags().GetInt("steps")
		all, _ := cmd.Flags().GetBool("all")

		if all {
			steps = 0
		}
		if steps == 0 && !all {
			steps = 1
		}

		return runner.Down(steps)
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current migration version",
	RunE: func(_ *cobra.Command, _ []string) error {
		runner, err := newMigrationRunner()
		if err != nil {
			return err
		}
		defer runner.Close()

		version, dirty, err := runner.Version()
		if err != nil {
			return err
		}

		log.Info("migration status",
			zap.Uint("version", version),
			zap.Bool("dirty", dirty),
		)
		fmt.Printf("Version: %d, Dirty: %v\n", version, dirty)
		return nil
	},
}

var migrateVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print current migration version",
	RunE: func(_ *cobra.Command, _ []string) error {
		runner, err := newMigrationRunner()
		if err != nil {
			return err
		}
		defer runner.Close()

		version, dirty, err := runner.Version()
		if err != nil {
			return err
		}

		fmt.Printf("Version: %d, Dirty: %v\n", version, dirty)
		return nil
	},
}

var migrateForceCmd = &cobra.Command{
	Use:   "force",
	Short: "Force set migration version",
	RunE: func(cmd *cobra.Command, _ []string) error {
		runner, err := newMigrationRunner()
		if err != nil {
			return err
		}
		defer runner.Close()

		version, _ := cmd.Flags().GetInt("version")
		return runner.Force(version)
	},
}

func newMigrationRunner() (*mongodb.MigrationRunner, error) {
	return mongodb.NewMigrationRunner(cfg.Mongo.URI, cfg.Mongo.Database, "migrations", log)
}

func init() {
	migrateDownCmd.Flags().IntP("steps", "s", 1, "Number of migrations to rollback")
	migrateDownCmd.Flags().Bool("all", false, "Rollback all migrations")
	migrateForceCmd.Flags().IntP("version", "v", 0, "Version to force")
	_ = migrateForceCmd.MarkFlagRequired("version")

	migrateCmd.AddCommand(migrateUpCmd, migrateDownCmd, migrateStatusCmd, migrateVersionCmd, migrateForceCmd)
	rootCmd.AddCommand(migrateCmd)
}
