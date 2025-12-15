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

type contextKey struct{}

var txContextKey = contextKey{}

// WithTx stores a transaction in the context.
func WithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txContextKey, tx)
}

// TxFromContext retrieves a transaction from the context.
// Returns nil if no transaction is present.
func TxFromContext(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(txContextKey).(*sql.Tx); ok {
		return tx
	}
	return nil
}

// HasTx checks if the context contains a transaction.
func HasTx(ctx context.Context) bool {
	return TxFromContext(ctx) != nil
}
