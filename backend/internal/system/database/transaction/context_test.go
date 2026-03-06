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
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
)

type ContextTestSuite struct {
	suite.Suite
}

func TestContextTestSuite(t *testing.T) {
	suite.Run(t, new(ContextTestSuite))
}

func (suite *ContextTestSuite) TestWithKeyedTx() {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	suite.Require().NoError(err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	tx, err := db.Begin()
	suite.Require().NoError(err)

	// Store transaction in context with key
	txCtx := WithKeyedTx(ctx, "test", tx)

	// Verify it's stored
	suite.NotNil(txCtx)
	suite.NotEqual(ctx, txCtx)
}

func (suite *ContextTestSuite) TestKeyedTxFromContext_WithTransaction() {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	suite.Require().NoError(err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	tx, err := db.Begin()
	suite.Require().NoError(err)

	// Store transaction in context with key
	txCtx := WithKeyedTx(ctx, "test", tx)

	// Retrieve it
	retrievedTx := KeyedTxFromContext(txCtx, "test")
	suite.NotNil(retrievedTx)
	suite.Equal(tx, retrievedTx)
}

func (suite *ContextTestSuite) TestKeyedTxFromContext_WithoutTransaction() {
	ctx := context.Background()

	// Try to retrieve transaction from empty context
	tx := KeyedTxFromContext(ctx, "test")
	suite.Nil(tx)
}

func (suite *ContextTestSuite) TestKeyedTxFromContext_WrongType() {
	ctx := context.Background()

	// Store something other than *sql.Tx in the context
	ctx = context.WithValue(ctx, getTxContextKey("test"), "not a transaction")

	// Should return nil
	tx := KeyedTxFromContext(ctx, "test")
	suite.Nil(tx)
}

func (suite *ContextTestSuite) TestHasKeyedTx_WithTransaction() {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	suite.Require().NoError(err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	tx, err := db.Begin()
	suite.Require().NoError(err)

	// Store transaction in context
	txCtx := WithKeyedTx(ctx, "test", tx)

	// Check if it has transaction
	suite.True(HasKeyedTx(txCtx, "test"))
}

func (suite *ContextTestSuite) TestHasKeyedTx_WithoutTransaction() {
	ctx := context.Background()

	// Check if empty context has transaction
	suite.False(HasKeyedTx(ctx, "test"))
}

func (suite *ContextTestSuite) TestHasKeyedTx_WrongType() {
	ctx := context.Background()

	// Store something other than *sql.Tx in the context
	ctx = context.WithValue(ctx, getTxContextKey("test"), "not a transaction")

	// Should return false
	suite.False(HasKeyedTx(ctx, "test"))
}
