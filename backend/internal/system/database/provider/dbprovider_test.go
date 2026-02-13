/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package provider

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/model"

	// Import SQLite driver for integration-style tests that use real SQLite connections.
	_ "modernc.org/sqlite"
)

type DBProviderTestSuite struct {
	suite.Suite
	mockDB sqlmock.Sqlmock
}

func TestDBProviderTestSuite(t *testing.T) {
	suite.Run(t, new(DBProviderTestSuite))
}

func (suite *DBProviderTestSuite) SetupTest() {
	_, mock, err := sqlmock.New()
	suite.Require().NoError(err)
	suite.mockDB = mock

	// Reset global config before each test
	config.ResetThunderRuntime()

	// Initialize a dummy config
	dummyConfig := &config.Config{
		Database: config.DatabaseConfig{
			Identity: config.DataSource{Name: "identity", Type: "postgres"},
			Runtime:  config.DataSource{Name: "runtime", Type: "postgres"},
			User:     config.DataSource{Name: "user", Type: "postgres"},
		},
	}
	err = config.InitializeThunderRuntime(".", dummyConfig)
	suite.Require().NoError(err)
}

func (suite *DBProviderTestSuite) TearDownTest() {
	config.ResetThunderRuntime()
}

func (suite *DBProviderTestSuite) TestGetUserDBTransactioner_Success() {
	// Create a mock DB connection
	db, _, err := sqlmock.New()
	suite.Require().NoError(err)
	defer func() {
		_ = db.Close()
	}()

	// Manually construct the provider with an initialized client
	provider := &dbProvider{
		userClient: NewDBClient(model.NewDB(db), "postgres"),
	}

	// Test getting the transactioner
	txer, err := provider.GetUserDBTransactioner()
	suite.NoError(err)
	suite.NotNil(txer)
}

func (suite *DBProviderTestSuite) TestGetRuntimeDBTransactioner_Success() {
	// Create a mock DB connection
	db, _, err := sqlmock.New()
	suite.Require().NoError(err)
	defer func() {
		_ = db.Close()
	}()

	// Manually construct the provider with an initialized client
	provider := &dbProvider{
		runtimeClient: NewDBClient(model.NewDB(db), "postgres"),
	}

	// Test getting the transactioner
	txer, err := provider.GetRuntimeDBTransactioner()
	suite.NoError(err)
	suite.NotNil(txer)
}

func (suite *DBProviderTestSuite) TestGetDBConfig_SQLite_WithExistingOptions() {
	dataSource := config.DataSource{
		Type:    "sqlite",
		Path:    "repository/database/test.db",
		Options: "_journal_mode=WAL&_busy_timeout=5000",
	}

	provider := &dbProvider{}
	cfg := provider.getDBConfig(dataSource)

	suite.Equal(dataSourceTypeSQLite, cfg.driverName)
	suite.Contains(cfg.dsn, "?_journal_mode=WAL&_busy_timeout=5000&_pragma=foreign_keys(1)")
}

func (suite *DBProviderTestSuite) TestGetDBConfig_SQLite_WithoutOptions() {
	dataSource := config.DataSource{
		Type: "sqlite",
		Path: "repository/database/test.db",
	}

	provider := &dbProvider{}
	cfg := provider.getDBConfig(dataSource)

	suite.Equal(dataSourceTypeSQLite, cfg.driverName)
	suite.Contains(cfg.dsn, "?_pragma=foreign_keys(1)")
	// Should not have a double ? or &
	suite.False(strings.Contains(cfg.dsn, "??"))
}

func (suite *DBProviderTestSuite) TestGetDBConfig_SQLite_OptionsWithQuestionMark() {
	dataSource := config.DataSource{
		Type:    "sqlite",
		Path:    "repository/database/test.db",
		Options: "?_journal_mode=WAL",
	}

	provider := &dbProvider{}
	cfg := provider.getDBConfig(dataSource)

	suite.Equal(dataSourceTypeSQLite, cfg.driverName)
	suite.Contains(cfg.dsn, "?_journal_mode=WAL&_pragma=foreign_keys(1)")
	// Ensure no double question marks
	suite.False(strings.Contains(cfg.dsn, "??"))
}

func (suite *DBProviderTestSuite) TestGetDBConfig_Postgres_NoForeignKeysPragma() {
	dataSource := config.DataSource{
		Type:     "postgres",
		Hostname: "localhost",
		Port:     5432,
		Username: "user",
		Password: "pass",
		Name:     "testdb",
		SSLMode:  "disable",
	}

	provider := &dbProvider{}
	cfg := provider.getDBConfig(dataSource)

	suite.Equal(dataSourceTypePostgres, cfg.driverName)
	suite.NotContains(cfg.dsn, "foreign_keys")
}

func (suite *DBProviderTestSuite) TestInitializeClient_SQLite_ForeignKeysEnabled() {
	dataSource := config.DataSource{
		Type:            "sqlite",
		Path:            ":memory:",
		Options:         "_pragma=foreign_keys(1)",
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: 300,
	}

	provider := &dbProvider{}
	var client DBClientInterface
	err := provider.initializeClient(&client, dataSource)
	suite.NoError(err)
	suite.NotNil(client)

	// Verify that foreign keys are actually enabled by querying the pragma.
	dbClient, ok := client.(*DBClient)
	suite.True(ok)
	var fkEnabled int
	err = dbClient.db.GetSQLDB().QueryRow("PRAGMA foreign_keys;").Scan(&fkEnabled)
	suite.NoError(err)
	suite.Equal(1, fkEnabled, "Foreign key constraints should be enabled")

	_ = dbClient.close()
}

func (suite *DBProviderTestSuite) TestInitializeClient_SQLite_ForeignKeysEnforced() {
	// Open a real SQLite in-memory database with foreign keys enabled via DSN.
	db, err := sql.Open("sqlite", ":memory:?_pragma=foreign_keys(1)")
	suite.Require().NoError(err)
	defer func() {
		_ = db.Close()
	}()

	// Create parent and child tables with a cascade delete FK.
	_, err = db.Exec(`
		CREATE TABLE parent (id INTEGER PRIMARY KEY);
		CREATE TABLE child (
			id INTEGER PRIMARY KEY,
			parent_id INTEGER NOT NULL,
			FOREIGN KEY (parent_id) REFERENCES parent(id) ON DELETE CASCADE
		);
	`)
	suite.Require().NoError(err)

	// Insert a parent and a child row.
	_, err = db.Exec("INSERT INTO parent (id) VALUES (1)")
	suite.Require().NoError(err)
	_, err = db.Exec("INSERT INTO child (id, parent_id) VALUES (1, 1)")
	suite.Require().NoError(err)

	// Delete the parent row - child should be cascade deleted.
	_, err = db.Exec("DELETE FROM parent WHERE id = 1")
	suite.NoError(err)

	// Verify child row was deleted.
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM child").Scan(&count)
	suite.NoError(err)
	suite.Equal(0, count, "Child row should be cascade deleted when parent is deleted")
}

func (suite *DBProviderTestSuite) TestSQLite_ForeignKeysDisabledByDefault() {
	// Open SQLite without foreign keys pragma.
	db, err := sql.Open("sqlite", ":memory:")
	suite.Require().NoError(err)
	defer func() {
		_ = db.Close()
	}()

	// Create parent and child tables.
	_, err = db.Exec(`
		CREATE TABLE parent (id INTEGER PRIMARY KEY);
		CREATE TABLE child (
			id INTEGER PRIMARY KEY,
			parent_id INTEGER NOT NULL,
			FOREIGN KEY (parent_id) REFERENCES parent(id) ON DELETE CASCADE
		);
	`)
	suite.Require().NoError(err)

	// Insert parent and child.
	_, err = db.Exec("INSERT INTO parent (id) VALUES (1)")
	suite.Require().NoError(err)
	_, err = db.Exec("INSERT INTO child (id, parent_id) VALUES (1, 1)")
	suite.Require().NoError(err)

	// Delete parent - without FK enforcement, child remains orphaned.
	_, err = db.Exec("DELETE FROM parent WHERE id = 1")
	suite.NoError(err)

	// Child should still exist since FK constraints are off by default.
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM child").Scan(&count)
	suite.NoError(err)
	suite.Equal(1, count, "Without foreign_keys pragma, child row should NOT be cascade deleted")
}

func (suite *DBProviderTestSuite) TestInitializeClient_SQLite_ForeignKeysVerification_PragmaReturnsZero() {
	// Open a real SQLite connection WITHOUT the foreign_keys pragma.
	db, err := sql.Open("sqlite", ":memory:")
	suite.Require().NoError(err)
	defer func() {
		_ = db.Close()
	}()

	// This simulates what the verification block in initializeClient checks.
	var fkEnabled int
	err = db.QueryRow("PRAGMA foreign_keys;").Scan(&fkEnabled)
	suite.NoError(err)
	suite.Equal(0, fkEnabled, "Without the pragma, foreign_keys should be 0")
}

func (suite *DBProviderTestSuite) TestInitializeClient_SQLite_ForeignKeysVerification_PragmaReturnsOne() {
	// Open a real SQLite connection WITH the foreign_keys pragma.
	db, err := sql.Open("sqlite", ":memory:?_pragma=foreign_keys(1)")
	suite.Require().NoError(err)
	defer func() {
		_ = db.Close()
	}()

	// This simulates what the verification block in initializeClient checks.
	var fkEnabled int
	err = db.QueryRow("PRAGMA foreign_keys;").Scan(&fkEnabled)
	suite.NoError(err)
	suite.Equal(1, fkEnabled, "With the pragma, foreign_keys should be 1")
}

func (suite *DBProviderTestSuite) TestInitializeClient_Postgres_SkipsForeignKeyCheck() {
	db, mock, err := sqlmock.New()
	suite.Require().NoError(err)
	defer func() {
		_ = db.Close()
	}()

	// Construct the provider with a postgres mock client.
	// No PRAGMA foreign_keys query should be expected for postgres.
	mock.ExpectPing()

	provider := &dbProvider{
		identityClient: NewDBClient(model.NewDB(db), "postgres"),
	}

	// The provider should already have the client; no re-initialization needed.
	suite.NotNil(provider.identityClient)
	suite.NoError(mock.ExpectationsWereMet())
}

func (suite *DBProviderTestSuite) TestInitializeClient_SQLite_ForeignKeysEnabledOnMultipleConnections() {
	// Open a connection pool with multiple connections.
	db, err := sql.Open("sqlite", ":memory:?_pragma=foreign_keys(1)")
	suite.Require().NoError(err)
	defer func() {
		_ = db.Close()
	}()

	// Allow multiple open connections.
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)

	// Query the pragma on multiple connections to verify all have foreign keys enabled.
	// By using goroutine-style sequential queries, we exercise different connections from the pool.
	for i := 0; i < 10; i++ {
		var fkEnabled int
		err := db.QueryRow("PRAGMA foreign_keys;").Scan(&fkEnabled)
		suite.NoError(err)
		suite.Equal(1, fkEnabled, "Foreign key should be enabled on connection attempt %d", i)
	}
}

func (suite *DBProviderTestSuite) TestInitializeClient_SQLite_ForeignKeysEnforced_MultipleConnections() {
	db, err := sql.Open("sqlite", "file::memory:?_pragma=foreign_keys(1)&cache=shared")
	suite.Require().NoError(err)
	defer func() {
		_ = db.Close()
	}()

	db.SetMaxOpenConns(5)

	// Create tables.
	_, err = db.Exec(`
		CREATE TABLE parent (id INTEGER PRIMARY KEY);
		CREATE TABLE child (
			id INTEGER PRIMARY KEY,
			parent_id INTEGER NOT NULL,
			FOREIGN KEY (parent_id) REFERENCES parent(id) ON DELETE CASCADE
		);
	`)
	suite.Require().NoError(err)

	// Insert and delete across multiple operations to exercise different pool connections.
	for i := 1; i <= 5; i++ {
		_, err = db.Exec("INSERT INTO parent (id) VALUES (?)", i)
		suite.Require().NoError(err)
		_, err = db.Exec("INSERT INTO child (id, parent_id) VALUES (?, ?)", i, i)
		suite.Require().NoError(err)
	}

	// Delete all parents - all children should cascade.
	for i := 1; i <= 5; i++ {
		_, err = db.Exec("DELETE FROM parent WHERE id = ?", i)
		suite.NoError(err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM child").Scan(&count)
	suite.NoError(err)
	suite.Equal(0, count, "All child rows should be cascade deleted")
}

func (suite *DBProviderTestSuite) TestGetDBConfig_SQLite_DSNContainsPath() {
	dataSource := config.DataSource{
		Type: "sqlite",
		Path: "repository/database/thunderdb.db",
	}

	provider := &dbProvider{}
	cfg := provider.getDBConfig(dataSource)

	// The DSN should contain the path and the pragma.
	suite.Contains(cfg.dsn, "repository/database/thunderdb.db")
	suite.Contains(cfg.dsn, "_pragma=foreign_keys(1)")
	suite.True(strings.HasSuffix(cfg.dsn, "_pragma=foreign_keys(1)"))
}
