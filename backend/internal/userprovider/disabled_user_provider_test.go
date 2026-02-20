/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

package userprovider

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
)

type DisabledUserProviderTestSuite struct {
	suite.Suite
	provider UserProviderInterface
}

func (suite *DisabledUserProviderTestSuite) SetupTest() {
	suite.provider = newDisabledUserProvider()
}

func TestDisabledUserProviderTestSuite(t *testing.T) {
	suite.Run(t, new(DisabledUserProviderTestSuite))
}

func (suite *DisabledUserProviderTestSuite) TestIdentifyUser() {
	userID, err := suite.provider.IdentifyUser(map[string]interface{}{})
	suite.Nil(userID)
	suite.Equal(errNotImplemented, err)
}

func (suite *DisabledUserProviderTestSuite) TestGetUser() {
	user, err := suite.provider.GetUser("user-id")
	suite.Nil(user)
	suite.Equal(errNotImplemented, err)
}

func (suite *DisabledUserProviderTestSuite) TestGetUserGroups() {
	groups, err := suite.provider.GetUserGroups("user-id", 10, 0)
	suite.Nil(groups)
	suite.Equal(errNotImplemented, err)
}

func (suite *DisabledUserProviderTestSuite) TestUpdateUser() {
	user, err := suite.provider.UpdateUser("user-id", &User{})
	suite.Nil(user)
	suite.Equal(errNotImplemented, err)
}

func (suite *DisabledUserProviderTestSuite) TestCreateUser() {
	user, err := suite.provider.CreateUser(&User{})
	suite.Nil(user)
	suite.Equal(errNotImplemented, err)
}

func (suite *DisabledUserProviderTestSuite) TestUpdateUserCredentials() {
	err := suite.provider.UpdateUserCredentials("user-id", json.RawMessage{})
	suite.Equal(errNotImplemented, err)
}

func (suite *DisabledUserProviderTestSuite) TestDeleteUser() {
	err := suite.provider.DeleteUser("user-id")
	suite.Equal(errNotImplemented, err)
}
