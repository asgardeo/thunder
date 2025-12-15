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

package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/asgardeo/thunder/internal/system/log"
)

// Transactioner provides transaction management with automatic nesting detection.
type Transactioner interface {
	// Transact executes the given function within a transaction.
	// If a transaction already exists in the context, it reuses it.
	// Otherwise, it creates a new transaction and commits/rolls back automatically.
	Transact(ctx context.Context, txFunc func(context.Context) error) error

	// TransactWithOptions executes the given function within a transaction with specific options.
	// Options include isolation level, read-only mode, etc.
	// If a transaction already exists in the context, options are ignored and the existing transaction is reused.
	TransactWithOptions(ctx context.Context, opts *sql.TxOptions, txFunc func(context.Context) error) error
}

// transactioner is the default implementation of Transactioner.
type transactioner struct {
	db *sql.DB
}

// NewTransactioner creates a new Transactioner instance.
func NewTransactioner(db *sql.DB) Transactioner {
	return &transactioner{db: db}
}

// Transact executes the given function within a transaction.
// It automatically detects if a transaction already exists in the context and reuses it.
func (t *transactioner) Transact(ctx context.Context, txFunc func(context.Context) error) error {
	// Check if we're already in a transaction
	if HasTx(ctx) {
		// Already in a transaction - just execute the function without creating a new one
		return txFunc(ctx)
	}

	// Not in a transaction - start a new one with default options
	return t.TransactWithOptions(ctx, nil, txFunc)
}

// TransactWithOptions executes the given function within a transaction with specific options.
// If a transaction already exists in the context, the options are ignored and the existing transaction is reused.
func (t *transactioner) TransactWithOptions(
	ctx context.Context,
	opts *sql.TxOptions,
	txFunc func(context.Context) error,
) (err error) {
	// Check if we're already in a transaction
	if HasTx(ctx) {
		// Already in a transaction - reuse it
		// Note: opts are ignored since we can't change transaction properties mid-transaction
		return txFunc(ctx)
	}

	// Start a new transaction
	tx, txErr := t.db.BeginTx(ctx, opts)
	if txErr != nil {
		return txErr
	}

	// Use defer to handle commit/rollback and panic recovery
	defer func() {
		if p := recover(); p != nil {
			// Panic occurred - rollback and convert panic to error
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.GetLogger().Error("failed to rollback transaction after unexpected error", log.Error(rollbackErr))
				// Convert panic to error and join with rollback error
				switch v := p.(type) {
				case error:
					err = errors.Join(fmt.Errorf("transaction aborted unexpectedly: %w", v), rollbackErr)
				default:
					err = errors.Join(fmt.Errorf("transaction aborted unexpectedly: %v", v), rollbackErr)
				}
			} else {
				// Convert panic to error
				switch v := p.(type) {
				case error:
					err = fmt.Errorf("transaction aborted unexpectedly: %w", v)
				default:
					err = fmt.Errorf("transaction aborted unexpectedly: %v", v)
				}
			}
		} else if err != nil {
			// Error occurred - rollback
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.GetLogger().Error("failed to rollback transaction", log.Error(rollbackErr))
				err = errors.Join(err, rollbackErr)
			}
		} else {
			// Success - commit
			err = tx.Commit()
		}
	}()

	// Execute the function with the transaction stored in the context
	err = txFunc(WithTx(ctx, tx))
	return err
}
