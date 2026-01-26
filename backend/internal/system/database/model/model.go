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

// Package model defines the data structures and interfaces for database operations.
package model

import (
	"context"
	"database/sql"
)

// DBInterface defines the wrapper interface for database operations.
type DBInterface interface {
	Query(query string, args ...any) (*sql.Rows, error)
	Exec(query string, args ...any) (sql.Result, error)
	Begin() (*sql.Tx, error)
	Close() error
	// GetSQLDB returns the underlying *sql.DB for advanced operations like transaction management.
	GetSQLDB() *sql.DB
}

// DB is the implementation of DBInterface for managing database connections.
type DB struct {
	internal *sql.DB
}

// NewDB creates a new instance of DB with the provided sql.DB.
func NewDB(db *sql.DB) DBInterface {
	return &DB{
		internal: db,
	}
}

// Query executes a query that returns rows, typically a SELECT, and returns the result as *sql.Rows.
func (d *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return d.internal.Query(query, args...)
}

// Exec executes a query without returning data in any rows, and returns sql.Result.
func (d *DB) Exec(query string, args ...any) (sql.Result, error) {
	return d.internal.Exec(query, args...)
}

// Begin starts a new database transaction and returns *sql.Tx.
func (d *DB) Begin() (*sql.Tx, error) {
	return d.internal.Begin()
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.internal.Close()
}

// GetSQLDB returns the underlying *sql.DB instance.
// This is used for creating transactioners and other advanced database operations.
func (d *DB) GetSQLDB() *sql.DB {
	return d.internal
}

// TxInterface defines the wrapper interface for transaction management.
type TxInterface interface {
	Commit() error
	Rollback() error
	Exec(query DBQuery, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query DBQuery, args ...any) (sql.Result, error)
	Query(query DBQuery, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query DBQuery, args ...any) (*sql.Rows, error)
}

// Tx is the implementation of TxInterface for managing database transactions.
type Tx struct {
	internal *sql.Tx
	dbType   string
}

// NewTx creates a new instance of Tx with the provided sql.Tx.
func NewTx(tx *sql.Tx, dbType string) TxInterface {
	return &Tx{
		internal: tx,
		dbType:   dbType,
	}
}

// Commit commits the transaction.
func (t *Tx) Commit() error {
	return t.internal.Commit()
}

// Rollback rolls back the transaction.
func (t *Tx) Rollback() error {
	return t.internal.Rollback()
}

// Exec executes a query with the given arguments.
func (t *Tx) Exec(query DBQuery, args ...any) (sql.Result, error) {
	return t.ExecContext(context.Background(), query, args...)
}

// ExecContext executes a query with the given arguments and context.
func (t *Tx) ExecContext(ctx context.Context, query DBQuery, args ...any) (sql.Result, error) {
	sqlQuery := query.GetQuery(t.dbType)
	return t.internal.ExecContext(ctx, sqlQuery, args...)
}

// Query executes a query that returns rows, typically a SELECT, and returns the result as *sql.Rows.
func (t *Tx) Query(query DBQuery, args ...any) (*sql.Rows, error) {
	return t.QueryContext(context.Background(), query, args...)
}

// QueryContext executes a query that returns rows with the given context.
func (t *Tx) QueryContext(ctx context.Context, query DBQuery, args ...any) (*sql.Rows, error) {
	sqlQuery := query.GetQuery(t.dbType)
	return t.internal.QueryContext(ctx, sqlQuery, args...)
}
