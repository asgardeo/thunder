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

// Package transaction provides database transaction management capabilities.
package transaction

import (
	"context"
	"database/sql"
)

type contextKey string

// There is no default context key to enforce explicit database naming in transactions.

func getTxContextKey(dbName string) contextKey {
	return contextKey("tx_" + dbName)
}

// WithKeyedTx stores a transaction in the context with a database name.
func WithKeyedTx(ctx context.Context, dbName string, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, getTxContextKey(dbName), tx)
}

// KeyedTxFromContext retrieves a transaction from the context with a database name.
func KeyedTxFromContext(ctx context.Context, dbName string) *sql.Tx {
	if tx, ok := ctx.Value(getTxContextKey(dbName)).(*sql.Tx); ok {
		return tx
	}
	return nil
}

// HasKeyedTx checks if the context contains a transaction for a database name.
func HasKeyedTx(ctx context.Context, dbName string) bool {
	return KeyedTxFromContext(ctx, dbName) != nil
}
