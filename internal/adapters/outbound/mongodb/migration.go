package mongodb

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mongodb" // register mongodb driver
	_ "github.com/golang-migrate/migrate/v4/source/file"      // register file source
	"go.uber.org/zap"

	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// MigrationRunner wraps golang-migrate for MongoDB.
type MigrationRunner struct {
	m   *migrate.Migrate
	log logger.Logger
}

// NewMigrationRunner creates a new MigrationRunner.
func NewMigrationRunner(uri, dbName, sourcePath string, log logger.Logger) (*MigrationRunner, error) {
	// Resolve to absolute path so file:// URL is always file:///abs/path,
	// not file://relative (which URL-parses "relative" as the hostname).
	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("mongodb.NewMigrationRunner: failed to resolve migrations path: %w", err)
	}

	dbURL := fmt.Sprintf("%s/%s", uri, dbName)
	sourceURL := fmt.Sprintf("file://%s", absPath)

	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		return nil, fmt.Errorf("mongodb.NewMigrationRunner: %w", err)
	}

	return &MigrationRunner{m: m, log: log}, nil
}

// Up runs all pending migrations.
func (r *MigrationRunner) Up() error {
	r.log.Info("running migrations up")
	if err := r.m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("mongodb.MigrationRunner.Up: %w", err)
	}
	r.log.Info("migrations up complete")
	return nil
}

// Down rolls back migrations by the given number of steps.
// If steps <= 0, all migrations are rolled back.
func (r *MigrationRunner) Down(steps int) error {
	r.log.Info("running migrations down", zap.Int("steps", steps))
	if steps <= 0 {
		if err := r.m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("mongodb.MigrationRunner.Down: %w", err)
		}
	} else {
		if err := r.m.Steps(-steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("mongodb.MigrationRunner.Down: %w", err)
		}
	}
	r.log.Info("migrations down complete")
	return nil
}

// Force sets the migration version without running the migration.
func (r *MigrationRunner) Force(version int) error {
	r.log.Info("forcing migration version", zap.Int("version", version))
	if err := r.m.Force(version); err != nil {
		return fmt.Errorf("mongodb.MigrationRunner.Force: %w", err)
	}
	return nil
}

// Version returns the current migration version and dirty state.
func (r *MigrationRunner) Version() (uint, bool, error) {
	version, dirty, err := r.m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return 0, false, fmt.Errorf("mongodb.MigrationRunner.Version: %w", err)
	}
	return version, dirty, nil
}

// Close closes the migration runner and releases resources.
func (r *MigrationRunner) Close() error {
	srcErr, dbErr := r.m.Close()
	if srcErr != nil {
		return fmt.Errorf("mongodb.MigrationRunner.Close: source: %w", srcErr)
	}
	if dbErr != nil {
		return fmt.Errorf("mongodb.MigrationRunner.Close: database: %w", dbErr)
	}
	return nil
}
