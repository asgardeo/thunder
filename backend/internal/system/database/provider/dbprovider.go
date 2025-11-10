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
	"github.com/asgardeo/thunder/internal/system/database"
	"github.com/asgardeo/thunder/internal/system/log"
)

const (
	dataSourceTypePostgres = "postgres"
	dataSourceTypeSQLite   = "sqlite"
)

// dbConfig represents the local database configuration.
type dbConfig struct {
	dsn        string
	driverName string
}

// DBProviderInterface defines the interface for getting database clients.
type DBProviderInterface interface {
	GetUserDBClient() (DBClientInterface, error)
	GetRuntimeDBClient() (DBClientInterface, error)
	GetConfigDBClient() (DBClientInterface, error)
}

// DBProviderCloser is a separate interface for closing the provider.
// Only the lifecycle manager should use this interface.
type DBProviderCloser interface {
	Close() error
}

// dbProvider is the implementation of DBProviderInterface.
type dbProvider struct {
	configClient  DBClientInterface
	configMutex   sync.RWMutex
	runtimeClient DBClientInterface
	runtimeMutex  sync.RWMutex
	userClient    DBClientInterface
	userMutex     sync.RWMutex
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

// GetUserDBClient returns a database client for user datasource.
// Not required to close the returned client manually since it manages its own connection pool.
func (d *dbProvider) GetUserDBClient() (DBClientInterface, error) {
	userDBConfig := config.GetThunderRuntime().Config.Database.User
	return d.getOrInitClient(&d.userClient, &d.userMutex, userDBConfig)
}

// GetRuntimeDBClient returns a database client for runtime datasource.
// Not required to close the returned client manually since it manages its own connection pool.
func (d *dbProvider) GetRuntimeDBClient() (DBClientInterface, error) {
	runtimeDBConfig := config.GetThunderRuntime().Config.Database.Runtime
	return d.getOrInitClient(&d.runtimeClient, &d.runtimeMutex, runtimeDBConfig)
}

// GetConfigDBClient returns a database client for config datasource.
// Not required to close the returned client manually since it manages its own connection pool.
func (d *dbProvider) GetConfigDBClient() (DBClientInterface, error) {
	configDBConfig := config.GetThunderRuntime().Config.Database.Config
	return d.getOrInitClient(&d.configClient, &d.configMutex, configDBConfig)
}

// initializeAllClients initializes both identity and runtime clients at startup.
func (d *dbProvider) initializeAllClients() {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "DBProvider"))

	configDBConfig := config.GetThunderRuntime().Config.Database.Config
	err := d.initializeClient(&d.configClient, configDBConfig)
	if err != nil {
		logger.Error("Failed to initialize identity database client", log.Error(err))
	}

	runtimeDBConfig := config.GetThunderRuntime().Config.Database.Runtime
	err = d.initializeClient(&d.runtimeClient, runtimeDBConfig)
	if err != nil {
		logger.Error("Failed to initialize runtime database client", log.Error(err))
	}

	userDBConfig := config.GetThunderRuntime().Config.Database.User
	err = d.initializeClient(&d.userClient, userDBConfig)
	if err != nil {
		logger.Error("Failed to initialize user database client", log.Error(err))
	}
}

// getOrInitClient gets or initializes a DB client with locking.
func (d *dbProvider) getOrInitClient(
	clientPtr *DBClientInterface,
	mutex *sync.RWMutex,
	dataSource config.DataSource,
) (DBClientInterface, error) {
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

	if err := d.initializeClient(clientPtr, dataSource); err != nil {
		return nil, err
	}

	return *clientPtr, nil
}

// initializeClient initializes a database client and assigns it to the provided pointer.
func (d *dbProvider) initializeClient(clientPtr *DBClientInterface, dataSource config.DataSource) error {
	dbConfig := d.getDBConfig(dataSource)
	dbName := dataSource.Name

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

	*clientPtr = NewDBClient(db, dbConfig.driverName)
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

	configErr := d.closeClient(&d.configClient, &d.configMutex, database.ConfigDBClientName)
	runtimeErr := d.closeClient(&d.runtimeClient, &d.runtimeMutex, database.RuntimeDBClientName)
	userErr := d.closeClient(&d.userClient, &d.userMutex, database.UserDBClientName)
	return errors.Join(configErr, runtimeErr, userErr)
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
