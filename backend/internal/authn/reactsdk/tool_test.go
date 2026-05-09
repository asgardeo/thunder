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

package reactsdk

import (
	"context"
	"reflect"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ReactSDKToolsTestSuite struct {
	suite.Suite
}

func TestReactSDKToolsTestSuite(t *testing.T) {
	suite.Run(t, new(ReactSDKToolsTestSuite))
}

func (suite *ReactSDKToolsTestSuite) SetupTest() {

}

func (suite *ReactSDKToolsTestSuite) TestIntegrateReactSDK() {
	input := integrateReactSDKInput{
		ServerURL: "https://thunder.example.com",
	}

	result, output, err := integrateReactSDK(context.Background(), nil, input)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), output.Instructions, "https://thunder.example.com")
	assert.Contains(suite.T(), output.CodeSnippets, "https://thunder.example.com")
}

func (suite *ReactSDKToolsTestSuite) TestIntegrateReactSDK_Defaults() {
	// Empty input to trigger defaults
	input := integrateReactSDKInput{}

	result, output, err := integrateReactSDK(context.Background(), nil, input)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	// Should contain the default URL
	assert.Contains(suite.T(), output.Instructions, "https://localhost:8090")
	assert.Contains(suite.T(), output.CodeSnippets, "https://localhost:8090")
}

func (suite *ReactSDKToolsTestSuite) TestRegisterTools() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}, nil)

	RegisterTools(server)

	toolsField := reflect.ValueOf(server).Elem().FieldByName("tools")
	if !toolsField.IsValid() {
		suite.T().Fatal("tools field not found in mcp.Server")
	}

	featuresField := toolsField.Elem().FieldByName("features")
	if !featuresField.IsValid() {
		suite.T().Fatal("features field not found in featureSet")
	}

	// Check if our tool is in the map
	found := false
	iter := featuresField.MapRange()
	for iter.Next() {
		if iter.Key().String() == "thunderid_integrate_react_sdk" {
			found = true
			break
		}
	}
	assert.True(suite.T(), found, "thunderid_integrate_react_sdk tool should be registered")
}
