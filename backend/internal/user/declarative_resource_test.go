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

package user

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/declarative_resource/entity"
	"github.com/asgardeo/thunder/internal/system/log"
)

// DeclarativeResourceTestSuite tests user declarative resource parsing and export.
type DeclarativeResourceTestSuite struct {
	suite.Suite
}

// TestDeclarativeResourceTestSuite runs the test suite.
func TestDeclarativeResourceTestSuite(t *testing.T) {
	suite.Run(t, new(DeclarativeResourceTestSuite))
}

// SetupTest initializes runtime config required for hashing.
func (suite *DeclarativeResourceTestSuite) SetupTest() {
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("test", &config.Config{
		Crypto: config.CryptoConfig{
			PasswordHashing: config.PasswordHashingConfig{
				Algorithm: string(hash.SHA256),
				Parameters: config.PasswordHashingParamsConfig{
					SaltSize: 16,
				},
			},
		},
	})
	suite.Require().NoError(err)
}

func (suite *DeclarativeResourceTestSuite) TestParseCredentials_SimpleFormatHashes() {
	credentials, err := parseCredentials(map[string]interface{}{
		"password": "secret",
	})

	suite.NoError(err)
	suite.Contains(credentials, CredentialType("password"))
	suite.Len(credentials["password"], 1)
	suite.Equal("hash", credentials["password"][0].StorageType)
	suite.NotEqual("secret", credentials["password"][0].Value)
	suite.NotEmpty(credentials["password"][0].StorageAlgo)
}

func (suite *DeclarativeResourceTestSuite) TestParseCredentials_SystemManagedPreserves() {
	credentials, err := parseCredentials(map[string]interface{}{
		string(CredentialTypePasskey): "raw-value",
	})

	suite.NoError(err)
	suite.Contains(credentials, CredentialTypePasskey)
	suite.Len(credentials[CredentialTypePasskey], 1)
	suite.Equal("raw-value", credentials[CredentialTypePasskey][0].Value)
}

func (suite *DeclarativeResourceTestSuite) TestParseCredentials_FullFormatPreserves() {
	credentials, err := parseCredentials(map[string]interface{}{
		"password": []interface{}{
			map[string]interface{}{
				"storageType": "hash",
				"storageAlgo": "argon2",
				"storageAlgoParams": map[string]interface{}{
					"iterations": 1,
					"keySize":    32,
					"salt":       "salt",
				},
				"value": "hashed-value",
			},
		},
	})

	suite.NoError(err)
	suite.Len(credentials["password"], 1)
	suite.Equal("hash", credentials["password"][0].StorageType)
	suite.Equal("hashed-value", credentials["password"][0].Value)
}

func (suite *DeclarativeResourceTestSuite) TestParseCredentials_InvalidFormat() {
	_, err := parseCredentials(map[string]interface{}{
		"password": 123,
	})
	suite.Error(err)
}

func (suite *DeclarativeResourceTestSuite) TestParseCredentialObject_HashesWhenNoStorageType() {
	hashService := hash.Initialize()
	cred, err := parseCredentialObject(map[string]interface{}{
		"value": "secret",
	}, hashService, CredentialType("password"))

	suite.NoError(err)
	suite.Equal("hash", cred.StorageType)
	suite.NotEqual("secret", cred.Value)
}

func (suite *DeclarativeResourceTestSuite) TestParseCredentialObject_SystemManagedMarker() {
	hashService := hash.Initialize()
	cred, err := parseCredentialObject(map[string]interface{}{
		"value":             "raw",
		"systemManaged":     true,
		"storageType":       "system",
		"storageAlgo":       "",
		"storageAlgoParams": map[string]interface{}{},
	}, hashService, CredentialTypePasskey)

	suite.NoError(err)
	suite.Equal("system", cred.StorageType)
	suite.Equal("raw", cred.Value)
}

func (suite *DeclarativeResourceTestSuite) TestParseToUser_HashesCredentials() {
	yamlData := []byte("" +
		"id: user-1\n" +
		"type: person\n" +
		"ou_id: ou-1\n" +
		"attributes:\n" +
		"  username: alice\n" +
		"  email: alice@example.com\n" +
		"credentials:\n" +
		"  password: \"secret\"\n")

	resource, err := parseToUser(yamlData)
	suite.NoError(err)

	passwordCreds := resource.Credentials["password"]
	suite.Len(passwordCreds, 1)
	suite.NotEqual("secret", passwordCreds[0].Value)
}

func (suite *DeclarativeResourceTestSuite) TestParseToUserWrapper() {
	yamlData := []byte("" +
		"id: user-1\n" +
		"type: person\n" +
		"ou_id: ou-1\n" +
		"attributes:\n" +
		"  username: alice\n" +
		"  email: alice@example.com\n")

	resource, err := parseToUserWrapper(yamlData)
	suite.NoError(err)
	_, ok := resource.(*userResource)
	suite.True(ok)
}

func (suite *DeclarativeResourceTestSuite) TestUserExporter_GetResourceByID() {
	mockSvc := NewUserServiceInterfaceMock(suite.T())
	exporter := newUserExporter(mockSvc)

	attrs := json.RawMessage(`{"username":"alice"}`)
	mockSvc.On("GetUser", context.Background(), "user-1").
		Return(&User{ID: "user-1", Type: "person", OrganizationUnit: "ou-1", Attributes: attrs}, nil)

	resource, name, err := exporter.GetResourceByID(context.Background(), "user-1")
	suite.Nil(err)
	suite.Equal("alice", name)

	userResource, ok := resource.(*userDeclarativeResource)
	suite.True(ok)
	suite.Empty(userResource.Credentials)
}

func (suite *DeclarativeResourceTestSuite) TestUserExporter_Metadata() {
	exporter := newUserExporter(NewUserServiceInterfaceMock(suite.T()))

	suite.Equal(resourceTypeUser, exporter.GetResourceType())
	suite.Equal(paramTypeUser, exporter.GetParameterizerType())
}

func (suite *DeclarativeResourceTestSuite) TestUserExporter_GetAllResourceIDs() {
	ctx := context.Background()
	mockSvc := NewUserServiceInterfaceMock(suite.T())
	exporter := newUserExporter(mockSvc)

	users := []User{{ID: "user-1"}, {ID: "user-2"}}
	mockSvc.On("GetUserList", ctx, serverconst.MaxPageSize, 0, mock.Anything).
		Return(&UserListResponse{Users: users}, nil)
	mockSvc.On("IsUserDeclarative", ctx, "user-1").Return(true, nil)
	mockSvc.On("IsUserDeclarative", ctx, "user-2").Return(false, nil)
	mockSvc.On("GetUserList", ctx, serverconst.MaxPageSize, 2, mock.Anything).
		Return(&UserListResponse{Users: []User{}}, nil)

	ids, err := exporter.GetAllResourceIDs(ctx)
	suite.Nil(err)
	suite.Equal([]string{"user-2"}, ids)
}

func (suite *DeclarativeResourceTestSuite) TestLoadDeclarativeResources() {
	tempDir := suite.T().TempDir()
	usersDir := filepath.Join(tempDir, "repository", "resources", "users")
	suite.NoError(os.MkdirAll(usersDir, 0o750))

	userYAML := "" +
		"id: user-1\n" +
		"type: person\n" +
		"ou_id: ou-1\n" +
		"attributes:\n" +
		"  username: alice\n" +
		"  email: alice@example.com\n" +
		"credentials:\n" +
		"  password: \"secret\"\n"

	filePath := filepath.Join(usersDir, "user-1.yaml")
	suite.NoError(os.WriteFile(filePath, []byte(userYAML), 0o600))

	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime(tempDir, &config.Config{
		Crypto: config.CryptoConfig{
			PasswordHashing: config.PasswordHashingConfig{
				Algorithm: string(hash.SHA256),
				Parameters: config.PasswordHashingParamsConfig{
					SaltSize: 16,
				},
			},
		},
	})
	suite.Require().NoError(err)

	fileStore := &userFileBasedStore{
		GenericFileBasedStore: declarativeresource.NewGenericFileBasedStoreForTest(entity.KeyTypeUser),
	}

	err = loadDeclarativeResources(fileStore, nil)
	suite.NoError(err)

	user, err := fileStore.GetUser(context.Background(), "user-1")
	suite.NoError(err)
	suite.Equal("user-1", user.ID)
}

func (suite *DeclarativeResourceTestSuite) TestGetResourceRules_IncludesCredentials() {
	exporter := newUserExporter(NewUserServiceInterfaceMock(suite.T()))

	rules := exporter.GetResourceRules()
	suite.Contains(rules.DynamicPropertyFields, "Credentials")
}

func (suite *DeclarativeResourceTestSuite) TestValidateResource_MissingUsername() {
	exporter := newUserExporter(NewUserServiceInterfaceMock(suite.T()))

	resource := &userDeclarativeResource{
		ID:               "user-1",
		Type:             "person",
		OrganizationUnit: "ou-1",
		Attributes:       map[string]interface{}{},
	}

	_, err := exporter.ValidateResource(resource, "user-1", log.GetLogger())
	suite.NotNil(err)
}

func (suite *DeclarativeResourceTestSuite) TestValidateUserWrapper_Success() {
	fileStore := &userFileBasedStore{
		GenericFileBasedStore: declarativeresource.NewGenericFileBasedStoreForTest(entity.KeyTypeUser),
	}
	user := User{ID: "user-1", Type: "person", OrganizationUnit: "ou-1"}
	attrs, err := json.Marshal(map[string]interface{}{"username": "alice"})
	suite.NoError(err)
	user.Attributes = attrs

	resource := &userResource{User: user}
	storeMock := newUserStoreInterfaceMock(suite.T())
	storeMock.On("GetUser", context.Background(), "user-1").Return(User{}, ErrUserNotFound)

	err = validateUserWrapper(resource, fileStore, storeMock)
	suite.NoError(err)
}

func (suite *DeclarativeResourceTestSuite) TestValidateUserWrapper_DuplicateFileStore() {
	fileStore := &userFileBasedStore{
		GenericFileBasedStore: declarativeresource.NewGenericFileBasedStoreForTest(entity.KeyTypeUser),
	}
	user := User{ID: "user-1", Type: "person", OrganizationUnit: "ou-1"}
	attrs, err := json.Marshal(map[string]interface{}{"username": "alice"})
	suite.NoError(err)
	user.Attributes = attrs

	resource := &userResource{User: user}
	suite.NoError(fileStore.GenericFileBasedStore.Create("user-1", resource))

	err = validateUserWrapper(resource, fileStore, nil)
	suite.Error(err)
	suite.Contains(err.Error(), "duplicate user ID")
}

func (suite *DeclarativeResourceTestSuite) TestValidateUserWrapper_DBError() {
	fileStore := &userFileBasedStore{
		GenericFileBasedStore: declarativeresource.NewGenericFileBasedStoreForTest(entity.KeyTypeUser),
	}
	user := User{ID: "user-1", Type: "person", OrganizationUnit: "ou-1"}
	attrs, err := json.Marshal(map[string]interface{}{"username": "alice"})
	suite.NoError(err)
	user.Attributes = attrs

	resource := &userResource{User: user}
	storeMock := newUserStoreInterfaceMock(suite.T())
	storeMock.On("GetUser", context.Background(), "user-1").Return(User{}, errors.New("db error"))

	err = validateUserWrapper(resource, fileStore, storeMock)
	suite.Error(err)
	suite.Contains(err.Error(), "checking user existence")
}

func (suite *DeclarativeResourceTestSuite) TestParseCredentials_YAMLMapInterfaceFormat() {
	// Simulate YAML map[interface{}]interface{} for credentials
	credMap := map[interface{}]interface{}{
		"storageType": "hash",
		"storageAlgo": "argon2",
		"storageAlgoParams": map[interface{}]interface{}{
			"iterations": 2,
			"keySize":    64,
			"salt":       "salty",
		},
		"value": "hashed-value",
	}
	creds := map[string]interface{}{
		"password": []interface{}{credMap},
	}
	parsed, err := parseCredentials(creds)
	suite.NoError(err)
	suite.Len(parsed["password"], 1)
	suite.Equal("hash", parsed["password"][0].StorageType)
	suite.Equal("hashed-value", parsed["password"][0].Value)
	suite.Equal("argon2", string(parsed["password"][0].StorageAlgo))
	suite.Equal(2, parsed["password"][0].StorageAlgoParams.Iterations)
	suite.Equal(64, parsed["password"][0].StorageAlgoParams.KeySize)
	suite.Equal("salty", parsed["password"][0].StorageAlgoParams.Salt)
}

func (suite *DeclarativeResourceTestSuite) TestParseCredentialObject_YAMLMapInterfaceParams() {
	hashService := hash.Initialize()
	credMap := map[string]interface{}{
		"value":       "hashed-value",
		"storageType": "hash",
		"storageAlgo": "argon2",
		"storageAlgoParams": map[interface{}]interface{}{
			"iterations": 3,
			"keySize":    128,
			"salt":       "pepper",
		},
	}
	cred, err := parseCredentialObject(credMap, hashService, CredentialType("password"))
	suite.NoError(err)
	suite.Equal("hash", cred.StorageType)
	suite.Equal("hashed-value", cred.Value)
	suite.Equal("argon2", string(cred.StorageAlgo))
	suite.Equal(3, cred.StorageAlgoParams.Iterations)
	suite.Equal(128, cred.StorageAlgoParams.KeySize)
	suite.Equal("pepper", cred.StorageAlgoParams.Salt)
}

func (suite *DeclarativeResourceTestSuite) TestParseCredentials_InvalidCredentialMapType() {
	creds := map[string]interface{}{
		"password": []interface{}{123}, // not a map
	}
	_, err := parseCredentials(creds)
	suite.Error(err)
}
