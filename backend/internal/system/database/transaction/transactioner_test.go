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
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
)

type TransactionerTestSuite struct {
	suite.Suite
	db            *sql.DB
	mock          sqlmock.Sqlmock
	transactioner Transactioner
}

func TestTransactionerTestSuite(t *testing.T) {
	suite.Run(t, new(TransactionerTestSuite))
}

func (suite *TransactionerTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	suite.Require().NoError(err)

	suite.db = db
	suite.mock = mock
	suite.transactioner = NewTransactioner(db, "test")
}

func (suite *TransactionerTestSuite) TearDownTest() {
	_ = suite.db.Close()
}

func (suite *TransactionerTestSuite) TestTransact_Success() {
	ctx := context.Background()

	// Expect transaction to be started and committed
	suite.mock.ExpectBegin()
	suite.mock.ExpectCommit()

	executed := false
	err := suite.transactioner.Transact(ctx, func(txCtx context.Context) error {
		executed = true
		// Verify transaction is in context
		suite.True(HasKeyedTx(txCtx, "test"))
		suite.NotNil(KeyedTxFromContext(txCtx, "test"))
		return nil
	})

	suite.NoError(err)
	suite.True(executed)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TransactionerTestSuite) TestTransact_Error() {
	ctx := context.Background()
	expectedErr := errors.New("business logic error")

	// Expect transaction to be started and rolled back
	suite.mock.ExpectBegin()
	suite.mock.ExpectRollback()

	executed := false
	err := suite.transactioner.Transact(ctx, func(txCtx context.Context) error {
		executed = true
		return expectedErr
	})

	suite.Error(err)
	suite.Equal(expectedErr, err)
	suite.True(executed)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TransactionerTestSuite) TestTransact_Panic() {
	ctx := context.Background()

	// Expect transaction to be started and rolled back
	suite.mock.ExpectBegin()
	suite.mock.ExpectRollback()

	executed := false
	err := suite.transactioner.Transact(ctx, func(txCtx context.Context) error {
		executed = true
		panic("something went wrong")
	})

	suite.Error(err)
	suite.Contains(err.Error(), "transaction aborted unexpectedly")
	suite.Contains(err.Error(), "something went wrong")
	suite.True(executed)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TransactionerTestSuite) TestTransact_NestedTransaction() {
	ctx := context.Background()

	// Expect only ONE transaction to be started and committed
	suite.mock.ExpectBegin()
	suite.mock.ExpectCommit()

	outerExecuted := false
	innerExecuted := false

	err := suite.transactioner.Transact(ctx, func(txCtx1 context.Context) error {
		outerExecuted = true
		suite.True(HasKeyedTx(txCtx1, "test"))
		tx1 := KeyedTxFromContext(txCtx1, "test")
		suite.NotNil(tx1)

		// Nested call - should reuse the same transaction
		err := suite.transactioner.Transact(txCtx1, func(txCtx2 context.Context) error {
			innerExecuted = true
			suite.True(HasKeyedTx(txCtx2, "test"))
			tx2 := KeyedTxFromContext(txCtx2, "test")
			suite.NotNil(tx2)

			// Should be the same transaction
			suite.Equal(tx1, tx2)
			return nil
		})

		return err
	})

	suite.NoError(err)
	suite.True(outerExecuted)
	suite.True(innerExecuted)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TransactionerTestSuite) TestTransact_NestedTransactionWithError() {
	ctx := context.Background()
	expectedErr := errors.New("inner error")

	// Expect only ONE transaction to be started and rolled back
	suite.mock.ExpectBegin()
	suite.mock.ExpectRollback()

	outerExecuted := false
	innerExecuted := false

	err := suite.transactioner.Transact(ctx, func(txCtx1 context.Context) error {
		outerExecuted = true

		// Nested call that returns an error
		err := suite.transactioner.Transact(txCtx1, func(txCtx2 context.Context) error {
			innerExecuted = true
			return expectedErr
		})

		return err
	})

	suite.Error(err)
	suite.Equal(expectedErr, err)
	suite.True(outerExecuted)
	suite.True(innerExecuted)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TransactionerTestSuite) TestTransact_BeginError() {
	ctx := context.Background()
	expectedErr := errors.New("begin transaction failed")

	// Expect BeginTx to fail
	suite.mock.ExpectBegin().WillReturnError(expectedErr)

	executed := false
	err := suite.transactioner.Transact(ctx, func(txCtx context.Context) error {
		executed = true
		return nil
	})

	suite.Error(err)
	suite.ErrorIs(err, expectedErr)
	suite.Contains(err.Error(), "failed to begin transaction")
	suite.False(executed) // Function should not be executed if BeginTx fails
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TransactionerTestSuite) TestTransact_CommitError() {
	ctx := context.Background()
	expectedErr := errors.New("commit failed")

	// Expect transaction to be started, but commit should fail
	suite.mock.ExpectBegin()
	suite.mock.ExpectCommit().WillReturnError(expectedErr)

	executed := false
	err := suite.transactioner.Transact(ctx, func(txCtx context.Context) error {
		executed = true
		suite.True(HasKeyedTx(txCtx, "test"))
		return nil
	})

	suite.Error(err)
	suite.Equal(expectedErr, err)
	suite.True(executed) // Function should be executed even if commit fails
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TransactionerTestSuite) TestTransact_RollbackErrorOnPanic() {
	ctx := context.Background()
	rollbackErr := errors.New("rollback failed")

	// Expect transaction to be started and rolled back with error
	suite.mock.ExpectBegin()
	suite.mock.ExpectRollback().WillReturnError(rollbackErr)

	executed := false
	err := suite.transactioner.Transact(ctx, func(txCtx context.Context) error {
		executed = true
		panic("something went wrong")
	})

	// Should still get the panic error, not the rollback error
	suite.Error(err)
	suite.Contains(err.Error(), "transaction aborted unexpectedly")
	suite.Contains(err.Error(), "something went wrong")
	suite.True(executed)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TransactionerTestSuite) TestTransact_RollbackErrorOnError() {
	ctx := context.Background()
	businessErr := errors.New("business logic error")
	rollbackErr := errors.New("rollback failed")

	// Expect transaction to be started and rolled back with error
	suite.mock.ExpectBegin()
	suite.mock.ExpectRollback().WillReturnError(rollbackErr)

	executed := false
	err := suite.transactioner.Transact(ctx, func(txCtx context.Context) error {
		executed = true
		return businessErr
	})

	// Should get both the original business error and the rollback error
	suite.Error(err)
	suite.ErrorIs(err, businessErr)
	suite.ErrorIs(err, rollbackErr)
	suite.Contains(err.Error(), "business logic error")
	suite.Contains(err.Error(), "rollback failed")
	suite.True(executed)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TransactionerTestSuite) TestTransact_PanicWithNonErrorValue() {
	ctx := context.Background()

	// Expect transaction to be started and rolled back
	suite.mock.ExpectBegin()
	suite.mock.ExpectRollback()

	executed := false
	err := suite.transactioner.Transact(ctx, func(txCtx context.Context) error {
		executed = true
		panic("string panic value") // Panic with non-error type
	})

	// Should convert string panic to error
	suite.Error(err)
	suite.Contains(err.Error(), "transaction aborted unexpectedly")
	suite.Contains(err.Error(), "string panic value")
	suite.True(executed)
	suite.NoError(suite.mock.ExpectationsWereMet())
}
