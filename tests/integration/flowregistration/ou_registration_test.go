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

package flowregistration

import (
	"fmt"
	"testing"
	"time"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	mockNotificationServerPortOU = 8098
)

var (
	ouRegTestOU = testutils.OrganizationUnit{
		Handle:      "ou-reg-flow-test-ou",
		Name:        "OU Registration Flow Test Organization Unit",
		Description: "Organization unit for OU registration flow testing",
		Parent:      nil,
	}

	dynamicUserSchema = testutils.UserSchema{
		Name: "dynamic-user-type",
		Schema: map[string]interface{}{
			"username": map[string]interface{}{
				"type": "string",
			},
			"password": map[string]interface{}{
				"type": "string",
			},
			"email": map[string]interface{}{
				"type": "string",
			},
			"firstName": map[string]interface{}{
				"type": "string",
			},
			"lastName": map[string]interface{}{
				"type": "string",
			},
			"mobileNumber": map[string]interface{}{
				"type": "string",
			},
		},
	}
)

type OURegistrationFlowTestSuite struct {
	suite.Suite
	config             *TestSuiteConfig
	mockServer         *testutils.MockNotificationServer
	createdOUs         []string
	basicFlowTestAppID string
	basicFlowTestOUID  string
	smsFlowTestAppID   string
	smsFlowTestOUID    string
	userSchemaID       string
}

func TestOURegistrationFlowTestSuite(t *testing.T) {
	suite.Run(t, new(OURegistrationFlowTestSuite))
}

func (ts *OURegistrationFlowTestSuite) SetupSuite() {
	ts.config = &TestSuiteConfig{}
	ts.createdOUs = []string{}

	ouID, err := testutils.CreateOrganizationUnit(ouRegTestOU)
	if err != nil {
		ts.T().Fatalf("Failed to create test organization unit during setup: %v", err)
	}
	ts.basicFlowTestOUID = ouID

	// Create dynamic user schema
	dynamicUserSchema.OrganizationUnitId = ts.basicFlowTestOUID
	dynamicUserSchema.AllowSelfRegistration = true
	schemaID, err := testutils.CreateUserType(dynamicUserSchema)
	if err != nil {
		ts.T().Fatalf("Failed to create dynamic user schema during setup: %v", err)
	}
	ts.userSchemaID = schemaID

	// Create test application with allowed user types
	ouRegTestApp := testutils.Application{
		Name:                      "OU Registration Flow Test Application",
		Description:               "Application for testing OU registration flows",
		IsRegistrationFlowEnabled: true,
		AuthFlowGraphID:           "auth_flow_config_basic",
		RegistrationFlowGraphID:   "registration_flow_config_basic_with_ou",
		ClientID:                  "ou_reg_flow_test_client",
		ClientSecret:              "ou_reg_flow_test_secret",
		RedirectURIs:              []string{"http://localhost:3000/callback"},
		AllowedUserTypes:          []string{dynamicUserSchema.Name},
	}

	appID, err := testutils.CreateApplication(ouRegTestApp)
	if err != nil {
		ts.T().Fatalf("Failed to create test application during setup: %v", err)
	}
	ts.basicFlowTestAppID = appID

	smsApp := testutils.Application{
		Name:                      "OU SMS Registration Flow Test Application",
		Description:               "Application for testing OU SMS registration flows",
		IsRegistrationFlowEnabled: true,
		AuthFlowGraphID:           "auth_flow_config_sms",
		RegistrationFlowGraphID:   "registration_flow_config_sms_with_ou",
		ClientID:                  "ou_sms_reg_flow_test_client",
		ClientSecret:              "ou_sms_reg_flow_test_secret",
		RedirectURIs:              []string{"http://localhost:3000/callback"},
		AllowedUserTypes:          []string{dynamicUserSchema.Name},
	}

	smsAppID, err := testutils.CreateApplication(smsApp)
	if err != nil {
		ts.T().Fatalf("Failed to create SMS test application during setup: %v", err)
	}
	ts.smsFlowTestAppID = smsAppID

	smsOU := testutils.OrganizationUnit{
		Handle:      "ou-sms-reg-flow-test-ou",
		Name:        "OU SMS Registration Flow Test Organization Unit",
		Description: "Organization unit for OU SMS registration flow testing",
		Parent:      nil,
	}

	smsOUID, err := testutils.CreateOrganizationUnit(smsOU)
	if err != nil {
		ts.T().Fatalf("Failed to create SMS test organization unit during setup: %v", err)
	}
	ts.smsFlowTestOUID = smsOUID

	ts.mockServer = testutils.NewMockNotificationServer(mockNotificationServerPortOU)
	err = ts.mockServer.Start()
	if err != nil {
		ts.T().Fatalf("Failed to start mock notification server: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
}

func (ts *OURegistrationFlowTestSuite) TearDownSuite() {
	if err := testutils.CleanupUsers(ts.config.CreatedUserIDs); err != nil {
		ts.T().Logf("Failed to cleanup users during teardown: %v", err)
	}
	for _, ouID := range ts.createdOUs {
		if err := testutils.DeleteOrganizationUnit(ouID); err != nil {
			ts.T().Logf("Failed to delete created OU %s during teardown: %v", ouID, err)
		}
	}
	if ts.mockServer != nil {
		err := ts.mockServer.Stop()
		if err != nil {
			ts.T().Logf("Failed to stop mock notification server during teardown: %v", err)
		}
	}
	if ts.basicFlowTestAppID != "" {
		if err := testutils.DeleteApplication(ts.basicFlowTestAppID); err != nil {
			ts.T().Logf("Failed to delete test application during teardown: %v", err)
		}
	}
	if ts.smsFlowTestAppID != "" {
		if err := testutils.DeleteApplication(ts.smsFlowTestAppID); err != nil {
			ts.T().Logf("Failed to delete SMS test application during teardown: %v", err)
		}
	}
	if ts.basicFlowTestOUID != "" {
		if err := testutils.DeleteOrganizationUnit(ts.basicFlowTestOUID); err != nil {
			ts.T().Logf("Failed to delete test organization unit during teardown: %v", err)
		}
	}
	if ts.smsFlowTestOUID != "" {
		if err := testutils.DeleteOrganizationUnit(ts.smsFlowTestOUID); err != nil {
			ts.T().Logf("Failed to delete SMS test organization unit during teardown: %v", err)
		}
	}

	if ts.userSchemaID != "" {
		if err := testutils.DeleteUserType(ts.userSchemaID); err != nil {
			ts.T().Logf("Failed to delete dynamic user schema during teardown: %v", err)
		}
	}
}

func (ts *OURegistrationFlowTestSuite) TestBasicRegistrationFlowWithOU() {
	testCases := []struct {
		name          string
		ouName        string
		ouHandle      string
		ouDescription string
	}{
		{
			name:          "SuccessWithDescription",
			ouName:        "Test OU With Desc",
			ouHandle:      generateUniqueHandle("ou-desc"),
			ouDescription: "Test OU created with description",
		},
		{
			name:          "SuccessWithoutDescription",
			ouName:        "Test OU No Desc",
			ouHandle:      generateUniqueHandle("ou-nodesc"),
			ouDescription: "",
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			username := generateUniqueUsername("ouuser")
			inputs := map[string]string{
				"username":      username,
				"password":      "testpassword123",
				"ouName":        tc.ouName,
				"ouHandle":      tc.ouHandle,
				"ouDescription": tc.ouDescription,
				"firstName":     "Test",
				"lastName":      "User",
				"email":         username + "@example.com",
			}

			flowStep, err := initiateRegistrationFlow(ts.basicFlowTestAppID, false, inputs, "")
			ts.Require().NoError(err)
			ts.Require().Equal("COMPLETE", flowStep.FlowStatus)
			ts.Require().NotEmpty(flowStep.Assertion)

			jwtClaims, err := testutils.DecodeJWT(flowStep.Assertion)
			ts.Require().NoError(err)
			ts.Require().Equal(dynamicUserSchema.Name, jwtClaims.UserType)
			ts.Require().NotEmpty(jwtClaims.OuID)

			user, err := testutils.FindUserByAttribute("username", username)
			ts.Require().NoError(err)
			ts.Require().NotNil(user)

			if user != nil {
				ts.Require().Equal(jwtClaims.OuID, user.OrganizationUnit)
				ts.config.CreatedUserIDs = append(ts.config.CreatedUserIDs, user.ID)
			}

			ou, err := testutils.GetOrganizationUnit(jwtClaims.OuID)
			ts.Require().NoError(err)
			ts.Require().Equal(tc.ouName, ou.Name)
			ts.Require().Equal(tc.ouHandle, ou.Handle)
			ts.Require().NotNil(ou.Parent, "Created OU should have a parent")
			ts.Require().Equal(ts.basicFlowTestOUID, *ou.Parent,
				"Created OU should be a child of the user schema's OU")

			if tc.ouDescription != "" {
				ts.Require().Equal(tc.ouDescription, ou.Description)
			}

			ts.createdOUs = append(ts.createdOUs, jwtClaims.OuID)
		})
	}
}

func (ts *OURegistrationFlowTestSuite) TestBasicRegistrationFlowWithOUCreationDuplicateError() {
	testCases := []struct {
		name                string
		existingOUName      string
		existingOUHandle    string
		newOUName           string
		newOUHandle         string
		expectedErrorSubstr string
	}{
		{
			name:                "DuplicateOUName",
			existingOUName:      "Duplicate OU Name",
			existingOUHandle:    generateUniqueHandle("duplicate-name"),
			newOUName:           "Duplicate OU Name",
			newOUHandle:         generateUniqueHandle("new-handle"),
			expectedErrorSubstr: "An organization unit with the same name already exists",
		},
		{
			name:                "DuplicateOUHandle",
			existingOUName:      "Existing OU Handle",
			existingOUHandle:    generateUniqueHandle("duplicate-handle"),
			newOUName:           "New OU Name",
			newOUHandle:         "",
			expectedErrorSubstr: "An organization unit with the same handle already exists",
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			existingOU := testutils.OrganizationUnit{
				Handle:      tc.existingOUHandle,
				Name:        tc.existingOUName,
				Description: "Existing OU",
				Parent:      &ts.basicFlowTestOUID,
			}

			existingOUID, err := testutils.CreateOrganizationUnit(existingOU)
			ts.Require().NoError(err)
			ts.createdOUs = append(ts.createdOUs, existingOUID)

			username := generateUniqueUsername("dupou")
			newHandle := tc.newOUHandle
			if newHandle == "" {
				newHandle = tc.existingOUHandle
			}

			inputs := map[string]string{
				"username":      username,
				"password":      "testpassword123",
				"ouName":        tc.newOUName,
				"ouHandle":      newHandle,
				"ouDescription": "Should fail due to duplicate",
				"firstName":     "Test",
				"lastName":      "User",
				"email":         username + "@example.com",
			}

			flowStep, err := initiateRegistrationFlow(ts.basicFlowTestAppID, false, inputs, "")
			ts.Require().NoError(err)
			ts.Require().Equal("ERROR", flowStep.FlowStatus)
			ts.Require().Empty(flowStep.Assertion)
			ts.Require().Contains(flowStep.FailureReason, tc.expectedErrorSubstr)
		})
	}
}

func (ts *OURegistrationFlowTestSuite) TestSMSRegistrationFlowWithOUCreation() {
	testCases := []struct {
		name          string
		ouName        string
		ouHandle      string
		ouDescription string
	}{
		{
			name:          "SuccessWithDescription",
			ouName:        "Test SMS OU With Desc",
			ouHandle:      generateUniqueHandle("sms-ou-desc"),
			ouDescription: "Test SMS OU created with description",
		},
		{
			name:          "SuccessWithoutDescription",
			ouName:        "Test SMS OU No Desc",
			ouHandle:      generateUniqueHandle("sms-ou-nodesc"),
			ouDescription: "",
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			ts.mockServer.ClearMessages()

			mobileNumber := generateUniqueMobileNumber()
			inputs := map[string]string{
				"mobileNumber": mobileNumber,
			}

			flowStep, err := initiateRegistrationFlow(ts.smsFlowTestAppID, false, inputs, "")
			ts.Require().NoError(err)
			ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus)

			time.Sleep(500 * time.Millisecond)

			lastMessage := ts.mockServer.GetLastMessage()
			ts.Require().NotNil(lastMessage)
			ts.Require().NotEmpty(lastMessage.OTP)

			inputs = map[string]string{
				"otp":           lastMessage.OTP,
				"ouName":        tc.ouName,
				"ouHandle":      tc.ouHandle,
				"ouDescription": tc.ouDescription,
				"firstName":     "Test",
				"lastName":      "User",
				"email":         mobileNumber + "@example.com",
			}

			flowStep, err = completeRegistrationFlow(flowStep.FlowID, inputs, "")
			ts.Require().NoError(err)
			ts.Require().Equal("COMPLETE", flowStep.FlowStatus)
			ts.Require().NotEmpty(flowStep.Assertion)

			jwtClaims, err := testutils.DecodeJWT(flowStep.Assertion)
			ts.Require().NoError(err)
			ts.Require().Equal(dynamicUserSchema.Name, jwtClaims.UserType)
			ts.Require().NotEmpty(jwtClaims.OuID)

			user, err := testutils.FindUserByAttribute("mobileNumber", mobileNumber)
			ts.Require().NoError(err)
			ts.Require().NotNil(user)

			if user != nil {
				ts.Require().Equal(jwtClaims.OuID, user.OrganizationUnit)
				ts.config.CreatedUserIDs = append(ts.config.CreatedUserIDs, user.ID)
			}

			ou, err := testutils.GetOrganizationUnit(jwtClaims.OuID)
			ts.Require().NoError(err)
			ts.Require().Equal(tc.ouName, ou.Name)
			ts.Require().Equal(tc.ouHandle, ou.Handle)
			ts.Require().NotNil(ou.Parent, "Created OU should have a parent")
			ts.Require().Equal(ts.basicFlowTestOUID, *ou.Parent,
				"Created OU should be a child of the user schema's OU")

			if tc.ouDescription != "" {
				ts.Require().Equal(tc.ouDescription, ou.Description)
			}

			ts.createdOUs = append(ts.createdOUs, jwtClaims.OuID)
		})
	}
}

func (ts *OURegistrationFlowTestSuite) TestSMSRegistrationFlowWithOUCreationDuplicateError() {
	testCases := []struct {
		name                string
		existingOUName      string
		existingOUHandle    string
		newOUName           string
		newOUHandle         string
		expectedErrorSubstr string
	}{
		{
			name:                "DuplicateOUName",
			existingOUName:      "SMS Duplicate OU Name",
			existingOUHandle:    generateUniqueHandle("sms-duplicate-name"),
			newOUName:           "SMS Duplicate OU Name",
			newOUHandle:         generateUniqueHandle("new-sms-handle"),
			expectedErrorSubstr: "An organization unit with the same name already exists",
		},
		{
			name:                "DuplicateOUHandle",
			existingOUName:      "SMS Existing OU Handle",
			existingOUHandle:    generateUniqueHandle("sms-duplicate-handle"),
			newOUName:           "SMS New OU Name",
			newOUHandle:         "",
			expectedErrorSubstr: "An organization unit with the same handle already exists",
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			existingOU := testutils.OrganizationUnit{
				Handle:      tc.existingOUHandle,
				Name:        tc.existingOUName,
				Description: "Existing OU",
				Parent:      &ts.basicFlowTestOUID,
			}

			existingOUID, err := testutils.CreateOrganizationUnit(existingOU)
			ts.Require().NoError(err)
			ts.createdOUs = append(ts.createdOUs, existingOUID)

			ts.mockServer.ClearMessages()

			mobileNumber := generateUniqueMobileNumber()
			inputs := map[string]string{
				"mobileNumber": mobileNumber,
			}

			flowStep, err := initiateRegistrationFlow(ts.smsFlowTestAppID, false, inputs, "")
			ts.Require().NoError(err)

			time.Sleep(500 * time.Millisecond)

			lastMessage := ts.mockServer.GetLastMessage()
			ts.Require().NotNil(lastMessage)

			newHandle := tc.newOUHandle
			if newHandle == "" {
				newHandle = tc.existingOUHandle
			}

			inputs = map[string]string{
				"otp":           lastMessage.OTP,
				"ouName":        tc.newOUName,
				"ouHandle":      newHandle,
				"ouDescription": "Should fail due to duplicate",
				"firstName":     "Test",
				"lastName":      "User",
				"email":         mobileNumber + "@example.com",
			}

			flowStep, err = completeRegistrationFlow(flowStep.FlowID, inputs, "")
			ts.Require().NoError(err)
			ts.Require().Equal("ERROR", flowStep.FlowStatus)
			ts.Require().Empty(flowStep.Assertion)
			ts.Require().Contains(flowStep.FailureReason, tc.expectedErrorSubstr)
		})
	}
}

func generateUniqueHandle(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()%1000000)
}
