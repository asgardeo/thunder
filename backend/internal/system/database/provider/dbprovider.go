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

// Package provider provides functionality for managing database connections and clients.
package provider

import (
	"database/sql"
	"errors"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/transaction"
	"github.com/asgardeo/thunder/internal/system/log"
)

const (
	dataSourceTypePostgres = "postgres"
	dataSourceTypeSQLite   = "sqlite"

	dbNameConfig  = "config"
	dbNameRuntime = "runtime"
	dbNameUser    = "user"
)

// dbConfig represents the local database configuration.
type dbConfig struct {
	dsn        string
	driverName string
}

// DBProviderInterface defines the interface for getting database clients and transactioners.
type DBProviderInterface interface {
	GetConfigDBClient() (DBClientInterface, error)
	GetRuntimeDBClient() (DBClientInterface, error)
	GetUserDBClient() (DBClientInterface, error)
	GetConfigDBTransactioner() (transaction.Transactioner, error)
	GetUserDBTransactioner() (transaction.Transactioner, error)
	GetRuntimeDBTransactioner() (transaction.Transactioner, error)
}

// DBProviderCloser is a separate interface for closing the provider.
// Only the lifecycle manager should use this interface.
type DBProviderCloser interface {
	Close() error
}

// dbProvider is the implementation of DBProviderInterface.
type dbProvider struct {
	identityClient DBClientInterface
	identityMutex  sync.RWMutex
	runtimeClient  DBClientInterface
	runtimeMutex   sync.RWMutex
	userClient     DBClientInterface
	userMutex      sync.RWMutex
}

var (
	instance *dbProvider
	once     sync.Once
)

// initDBProvider initializes the singleton instance of DBProvider.
func initDBProvider() {
	once.Do(func() {
		instance = &dbProvider{}
		instance.initializeAllClients()
	})
}

// GetDBProvider returns the instance of DBProvider.
func GetDBProvider() DBProviderInterface {
	initDBProvider()
	return instance
}

// GetDBProviderCloser returns the DBProvider with closing capability.
// This should only be called from the main lifecycle manager.
func GetDBProviderCloser() DBProviderCloser {
	initDBProvider()
	return instance
}

// GetConfigDBClient returns a database client for config datasource.
// Not required to close the returned client manually since it manages its own connection pool.
func (d *dbProvider) GetConfigDBClient() (DBClientInterface, error) {
	identityDBConfig := config.GetThunderRuntime().Config.Database.Identity
	return d.getOrInitClient(&d.identityClient, &d.identityMutex, identityDBConfig, dbNameConfig)
}

// GetRuntimeDBClient returns a database client for runtime datasource.
// Not required to close the returned client manually since it manages its own connection pool.
func (d *dbProvider) GetRuntimeDBClient() (DBClientInterface, error) {
	runtimeDBConfig := config.GetThunderRuntime().Config.Database.Runtime
	return d.getOrInitClient(&d.runtimeClient, &d.runtimeMutex, runtimeDBConfig, dbNameRuntime)
}

// GetUserDBClient returns a database client for runtime datasource.
// Not required to close the returned client manually since it manages its own connection pool.
func (d *dbProvider) GetUserDBClient() (DBClientInterface, error) {
	userDBConfig := config.GetThunderRuntime().Config.Database.User
	return d.getOrInitClient(&d.userClient, &d.userMutex, userDBConfig, dbNameUser)
}

// GetConfigDBTransactioner returns a transactioner for the config database.
// The transactioner manages database transactions with automatic nesting detection.
func (d *dbProvider) GetConfigDBTransactioner() (transaction.Transactioner, error) {
	return d.getTransactioner(d.GetConfigDBClient, dbNameConfig)
}

// GetUserDBTransactioner returns a transactioner for the user database.
// The transactioner manages database transactions with automatic nesting detection.
func (d *dbProvider) GetUserDBTransactioner() (transaction.Transactioner, error) {
	return d.getTransactioner(d.GetUserDBClient, dbNameUser)
}

// GetRuntimeDBTransactioner returns a transactioner for the runtime database.
// The transactioner manages database transactions with automatic nesting detection.
func (d *dbProvider) GetRuntimeDBTransactioner() (transaction.Transactioner, error) {
	return d.getTransactioner(d.GetRuntimeDBClient, dbNameRuntime)
}

// getTransactioner is a helper method that creates a transactioner for a given database client.
func (d *dbProvider) getTransactioner(
	clientGetter func() (DBClientInterface, error),
	dbName string,
) (transaction.Transactioner, error) {
	client, err := clientGetter()
	if err != nil {
		return nil, fmt.Errorf("failed to get %s database client: %w", dbName, err)
	}

	return client.GetTransactioner()
}

// initializeAllClients initializes both identity and runtime clients at startup.
func (d *dbProvider) initializeAllClients() {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "DBProvider"))

	identityDBConfig := config.GetThunderRuntime().Config.Database.Identity
	err := d.initializeClient(&d.identityClient, identityDBConfig, dbNameConfig)
	if err != nil {
		logger.Error("Failed to initialize identity database client", log.Error(err))
	}

	runtimeDBConfig := config.GetThunderRuntime().Config.Database.Runtime
	err = d.initializeClient(&d.runtimeClient, runtimeDBConfig, dbNameRuntime)
	if err != nil {
		logger.Error("Failed to initialize runtime database client", log.Error(err))
	}

	userDBConfig := config.GetThunderRuntime().Config.Database.User
	err = d.initializeClient(&d.userClient, userDBConfig, dbNameUser)
	if err != nil {
		logger.Error("Failed to initialize user database client", log.Error(err))
	}
}

// getOrInitClient gets or initializes a DB client with locking.
func (d *dbProvider) getOrInitClient(
	clientPtr *DBClientInterface,
	mutex *sync.RWMutex,
	dataSource config.DataSource,
	dbName string,
) (DBClientInterface, error) {
	// Return error if database type is not configured
	if dataSource.Type == "" {
		return nil, fmt.Errorf("database type is not configured")
	}

	mutex.RLock()
	if *clientPtr != nil {
		client := *clientPtr
		mutex.RUnlock()
		return client, nil
	}
	mutex.RUnlock()

	mutex.Lock()
	defer mutex.Unlock()

	if *clientPtr != nil {
		return *clientPtr, nil
	}

	if err := d.initializeClient(clientPtr, dataSource, dbName); err != nil {
		return nil, err
	}

	return *clientPtr, nil
}

// initializeClient initializes a database client and assigns it to the provided pointer.
func (d *dbProvider) initializeClient(clientPtr *DBClientInterface, dataSource config.DataSource, dbName string) error {
	dbConfig := d.getDBConfig(dataSource)

	db, err := sql.Open(dbConfig.driverName, dbConfig.dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database %s: %w", dbName, err)
	}

	// Configure connection pool using values from configuration
	db.SetMaxOpenConns(dataSource.MaxOpenConns)
	db.SetMaxIdleConns(dataSource.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(dataSource.ConnMaxLifetime) * time.Second)

	// Test the database connection.
	if err := db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return fmt.Errorf("failed to ping database %s: %w (close error: %w)", dbName, err, closeErr)
		}
		return fmt.Errorf("failed to ping database %s: %w", dbName, err)
	}

	// Enable foreign key constraints for SQLite databases
	if dbConfig.driverName == dataSourceTypeSQLite {
		_, err := db.Exec("PRAGMA foreign_keys = ON;")
		if err != nil {
			if closeErr := db.Close(); closeErr != nil {
				return fmt.Errorf("failed to enable foreign key constraints for %s: %w (close error: %w)",
					dbName, err, closeErr)
			}
			return fmt.Errorf("failed to enable foreign key constraints for %s: %w", dbName, err)
		}
	}

	*clientPtr = NewDBClient(model.NewDB(db), dbConfig.driverName, dbName)
	return nil
}

// getDBConfig returns the database configuration based on the provided data source.
func (d *dbProvider) getDBConfig(dataSource config.DataSource) dbConfig {
	var dbConfig dbConfig

	switch dataSource.Type {
	case dataSourceTypePostgres:
		dbConfig.driverName = dataSourceTypePostgres
		dbConfig.dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			dataSource.Hostname, dataSource.Port, dataSource.Username, dataSource.Password,
			dataSource.Name, dataSource.SSLMode)
	case dataSourceTypeSQLite:
		dbConfig.driverName = dataSourceTypeSQLite
		options := dataSource.Options
		if options != "" && options[0] != '?' {
			options = "?" + options
		}
		dbConfig.dsn = fmt.Sprintf("%s%s", path.Join(config.GetThunderRuntime().ThunderHome, dataSource.Path), options)
	}

	return dbConfig
}

// Close closes the database connections. This should only be called by the lifecycle manager during shutdown.
func (d *dbProvider) Close() error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "DBProvider"))
	logger.Debug("Closing database connections")

	identityErr := d.closeClient(&d.identityClient, &d.identityMutex, "identity")
	runtimeErr := d.closeClient(&d.runtimeClient, &d.runtimeMutex, "runtime")
	userErr := d.closeClient(&d.userClient, &d.userMutex, "user")
	return errors.Join(identityErr, runtimeErr, userErr)
}

// closeClient is a helper to close a DB client with locking.
func (d *dbProvider) closeClient(clientPtr *DBClientInterface, mutex *sync.RWMutex, clientName string) error {
	mutex.Lock()
	defer mutex.Unlock()
	if *clientPtr != nil {
		if client, ok := (*clientPtr).(*DBClient); ok {
			if err := client.close(); err != nil {
				return fmt.Errorf("failed to close %s client: %w", clientName, err)
			}
		}
		*clientPtr = nil
	}
	return nil
}
