package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/frkr-io/frkr-common/migrate"
)

// TestDatabaseCompatibility verifies that migrations run successfully against
// both Postgres and CockroachDB when the correct URL scheme is used.
func TestDatabaseCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// 1. Create temporary migrations
	migrationsPath := createTempMigrations(t)
	defer os.RemoveAll(migrationsPath)

	// 2. Test Postgres
	t.Run("Postgres", func(t *testing.T) {
		// Start Postgres Container
		pgContainer, err := postgres.RunContainer(ctx,
			testcontainers.WithImage("postgres:15-alpine"),
			postgres.WithDatabase("frkr"),
			postgres.WithUsername("root"),
			postgres.WithPassword("password"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(30*time.Second)),
		)
		require.NoError(t, err)
		defer pgContainer.Terminate(ctx)

		connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
		require.NoError(t, err)

		// Postgres uses "postgres://" scheme by default from ConnectionString
		// Test migrations
		err = migrate.RunMigrations(connStr, migrationsPath)
		assert.NoError(t, err, "Migrations should succeed on Postgres")
		
		// Verify version
		version, dirty, err := migrate.GetVersion(connStr, migrationsPath)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), version)
		assert.False(t, dirty)
	})

	// 3. Test CockroachDB
	t.Run("CockroachDB", func(t *testing.T) {
		// Start CockroachDB Container
		// Note: CockroachDB module handles init
		crdbContainer, err := cockroachdb.RunContainer(ctx,
			testcontainers.WithImage("cockroachdb/cockroach:latest-v23.1"),
			// CRDB defaults: user=root, no pass, db=defaultdb. We need to create 'frkr'.
			// The module might not support creating arbitrary DBs easily via params, 
			// so we use the default and then exec sql, or just use 'defaultdb' for test.
			// Let's use 'defaultdb' as the target to simplify, migration doesn't care about DB name usually.
		)
		require.NoError(t, err)
		defer crdbContainer.Terminate(ctx)

		// Manually construct connection string to avoid module weirdness
		host, err := crdbContainer.Host(ctx)
		require.NoError(t, err)
		
		port, err := crdbContainer.MappedPort(ctx, "26257/tcp")
		require.NoError(t, err)

		// CRDB Default: user=root, db=defaultdb, sslmode=disable (insecure)
		// We use "cockroachdb://" scheme directly
		crdbConnStr := fmt.Sprintf("cockroachdb://root@%s:%s/defaultdb?sslmode=disable", host, port.Port())
		
		t.Logf("Testing with URL: %s", crdbConnStr)

		// Test migrations
		err = migrate.RunMigrations(crdbConnStr, migrationsPath)
		assert.NoError(t, err, "Migrations should succeed on CockroachDB with cockroachdb:// scheme")

		// Verify version
		version, dirty, err := migrate.GetVersion(crdbConnStr, migrationsPath)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), version)
		assert.False(t, dirty)
	})
}

func createTempMigrations(t *testing.T) string {
	dir, err := os.MkdirTemp("", "migrations")
	require.NoError(t, err)

	// Create a simple migration
	upSQL := `CREATE TABLE test_table (id SERIAL PRIMARY KEY, name TEXT);`
	err = os.WriteFile(filepath.Join(dir, "001_init.up.sql"), []byte(upSQL), 0644)
	require.NoError(t, err)
	
	downSQL := `DROP TABLE test_table;`
	err = os.WriteFile(filepath.Join(dir, "001_init.down.sql"), []byte(downSQL), 0644)
	require.NoError(t, err)

	return dir
}
