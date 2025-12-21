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

package model

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
)

type DBTestSuite struct {
	suite.Suite
	db     *sql.DB
	mock   sqlmock.Sqlmock
	dbImpl DBInterface
}

func TestDBTestSuite(t *testing.T) {
	suite.Run(t, new(DBTestSuite))
}

func (suite *DBTestSuite) SetupTest() {
	var err error
	suite.db, suite.mock, err = sqlmock.New()
	suite.Require().NoError(err)
	suite.dbImpl = NewDB(suite.db)
}

func (suite *DBTestSuite) TearDownTest() {
	_ = suite.db.Close()
}

func (suite *DBTestSuite) TestNewDB() {
	db := NewDB(suite.db)
	suite.NotNil(db)
	suite.IsType(&DB{}, db)
}

func (suite *DBTestSuite) TestDB_Query_Success() {
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "test")
	suite.mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(rows)

	result, err := suite.dbImpl.Query("SELECT * FROM users WHERE id = $1", 1)
	suite.NoError(err)
	suite.NotNil(result)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *DBTestSuite) TestDB_Query_Error() {
	suite.mock.ExpectQuery("SELECT \\* FROM users").
		WillReturnError(errors.New("query error"))

	result, err := suite.dbImpl.Query("SELECT * FROM users")
	suite.Error(err)
	suite.Nil(result)
	suite.Equal("query error", err.Error())
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *DBTestSuite) TestDB_Exec_Success() {
	suite.mock.ExpectExec("INSERT INTO users").
		WithArgs("john", "john@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := suite.dbImpl.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", "john", "john@example.com")
	suite.NoError(err)
	suite.NotNil(result)

	rowsAffected, err := result.RowsAffected()
	suite.NoError(err)
	suite.Equal(int64(1), rowsAffected)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *DBTestSuite) TestDB_Exec_Error() {
	suite.mock.ExpectExec("INSERT INTO users").
		WillReturnError(errors.New("exec error"))

	result, err := suite.dbImpl.Exec("INSERT INTO users (name) VALUES ($1)", "john")
	suite.Error(err)
	suite.Nil(result)
	suite.Equal("exec error", err.Error())
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *DBTestSuite) TestDB_Begin_Success() {
	suite.mock.ExpectBegin()

	tx, err := suite.dbImpl.Begin()
	suite.NoError(err)
	suite.NotNil(tx)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *DBTestSuite) TestDB_Begin_Error() {
	suite.mock.ExpectBegin().WillReturnError(errors.New("begin error"))

	tx, err := suite.dbImpl.Begin()
	suite.Error(err)
	suite.Nil(tx)
	suite.Equal("begin error", err.Error())
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *DBTestSuite) TestDB_Close_Success() {
	suite.mock.ExpectClose()

	err := suite.dbImpl.Close()
	suite.NoError(err)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *DBTestSuite) TestDB_Close_Error() {
	suite.mock.ExpectClose().WillReturnError(errors.New("close error"))

	err := suite.dbImpl.Close()
	suite.Error(err)
	suite.Equal("close error", err.Error())
	suite.NoError(suite.mock.ExpectationsWereMet())
}

type TxTestSuite struct {
	suite.Suite
	db     *sql.DB
	mock   sqlmock.Sqlmock
	tx     *sql.Tx
	txImpl TxInterface
}

func TestTxTestSuite(t *testing.T) {
	suite.Run(t, new(TxTestSuite))
}

func (suite *TxTestSuite) SetupTest() {
	var err error
	suite.db, suite.mock, err = sqlmock.New()
	suite.Require().NoError(err)

	suite.mock.ExpectBegin()
	suite.tx, err = suite.db.Begin()
	suite.Require().NoError(err)

	suite.txImpl = NewTx(suite.tx, "postgres")
}

func (suite *TxTestSuite) TearDownTest() {
	_ = suite.db.Close()
}

func (suite *TxTestSuite) TestNewTx() {
	tx := NewTx(suite.tx, "postgres")
	suite.NotNil(tx)
	suite.IsType(&Tx{}, tx)
}

func (suite *TxTestSuite) TestTx_Commit_Success() {
	suite.mock.ExpectCommit()

	err := suite.txImpl.Commit()
	suite.NoError(err)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TxTestSuite) TestTx_Commit_Error() {
	suite.mock.ExpectCommit().WillReturnError(errors.New("commit error"))

	err := suite.txImpl.Commit()
	suite.Error(err)
	suite.Equal("commit error", err.Error())
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TxTestSuite) TestTx_Rollback_Success() {
	suite.mock.ExpectRollback()

	err := suite.txImpl.Rollback()
	suite.NoError(err)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TxTestSuite) TestTx_Rollback_Error() {
	suite.mock.ExpectRollback().WillReturnError(errors.New("rollback error"))

	err := suite.txImpl.Rollback()
	suite.Error(err)
	suite.Equal("rollback error", err.Error())
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TxTestSuite) TestTx_Exec_Success_DefaultQuery() {
	query := DBQuery{
		ID:    "TEST-001",
		Query: "INSERT INTO users (name) VALUES ($1)",
	}

	suite.mock.ExpectExec("INSERT INTO users").
		WithArgs("john").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := suite.txImpl.Exec(query, "john")
	suite.NoError(err)
	suite.NotNil(result)

	rowsAffected, err := result.RowsAffected()
	suite.NoError(err)
	suite.Equal(int64(1), rowsAffected)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TxTestSuite) TestTx_Exec_Success_PostgresQuery() {
	suite.txImpl = NewTx(suite.tx, "postgres")

	query := DBQuery{
		ID:            "TEST-002",
		Query:         "INSERT INTO users (name) VALUES (?)",
		PostgresQuery: "INSERT INTO users (name) VALUES ($1)",
	}

	suite.mock.ExpectExec("INSERT INTO users").
		WithArgs("john").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := suite.txImpl.Exec(query, "john")
	suite.NoError(err)
	suite.NotNil(result)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TxTestSuite) TestTx_Exec_Success_SQLiteQuery() {
	suite.mock.ExpectBegin()
	tx, err := suite.db.Begin()
	suite.Require().NoError(err)
	suite.txImpl = NewTx(tx, "sqlite")

	query := DBQuery{
		ID:          "TEST-003",
		Query:       "INSERT INTO users (name) VALUES ($1)",
		SQLiteQuery: "INSERT INTO users (name) VALUES (?)",
	}

	suite.mock.ExpectExec("INSERT INTO users").
		WithArgs("john").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := suite.txImpl.Exec(query, "john")
	suite.NoError(err)
	suite.NotNil(result)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TxTestSuite) TestTx_Exec_Error() {
	query := DBQuery{
		ID:    "TEST-004",
		Query: "INSERT INTO users (name) VALUES ($1)",
	}

	suite.mock.ExpectExec("INSERT INTO users").
		WillReturnError(errors.New("exec error"))

	result, err := suite.txImpl.Exec(query, "john")
	suite.Error(err)
	suite.Nil(result)
	suite.Equal("exec error", err.Error())
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TxTestSuite) TestTx_Query_Success_DefaultQuery() {
	query := DBQuery{
		ID:    "TEST-005",
		Query: "SELECT * FROM users WHERE id = $1",
	}

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "john")
	suite.mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(rows)

	result, err := suite.txImpl.Query(query, 1)
	suite.NoError(err)
	suite.NotNil(result)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TxTestSuite) TestTx_Query_Success_PostgresQuery() {
	suite.txImpl = NewTx(suite.tx, "postgres")

	query := DBQuery{
		ID:            "TEST-006",
		Query:         "SELECT * FROM users WHERE id = ?",
		PostgresQuery: "SELECT * FROM users WHERE id = $1",
	}

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "john")
	suite.mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(rows)

	result, err := suite.txImpl.Query(query, 1)
	suite.NoError(err)
	suite.NotNil(result)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TxTestSuite) TestTx_Query_Success_SQLiteQuery() {
	suite.mock.ExpectBegin()
	tx, err := suite.db.Begin()
	suite.Require().NoError(err)
	suite.txImpl = NewTx(tx, "sqlite")

	query := DBQuery{
		ID:          "TEST-007",
		Query:       "SELECT * FROM users WHERE id = $1",
		SQLiteQuery: "SELECT * FROM users WHERE id = ?",
	}

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "john")
	suite.mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\?").
		WithArgs(1).
		WillReturnRows(rows)

	result, err := suite.txImpl.Query(query, 1)
	suite.NoError(err)
	suite.NotNil(result)
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *TxTestSuite) TestTx_Query_Error() {
	query := DBQuery{
		ID:    "TEST-008",
		Query: "SELECT * FROM users",
	}

	suite.mock.ExpectQuery("SELECT \\* FROM users").
		WillReturnError(errors.New("query error"))

	result, err := suite.txImpl.Query(query)
	suite.Error(err)
	suite.Nil(result)
	suite.Equal("query error", err.Error())
	suite.NoError(suite.mock.ExpectationsWereMet())
}
