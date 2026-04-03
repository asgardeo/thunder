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

package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/transaction"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// oAuthConfig is the internal structure for marshaling/unmarshaling the OAUTH_CONFIG JSON column.
type oAuthConfig struct {
	RedirectURIs            []string                      `json:"redirect_uris"`
	GrantTypes              []string                      `json:"grant_types"`
	ResponseTypes           []string                      `json:"response_types"`
	TokenEndpointAuthMethod string                        `json:"token_endpoint_auth_method"`
	PKCERequired            bool                          `json:"pkce_required"`
	PublicClient            bool                          `json:"public_client"`
	Token                   *oAuthTokenConfig             `json:"token,omitempty"`
	Scopes                  []string                      `json:"scopes,omitempty"`
	UserInfo                *userInfoConfig               `json:"user_info,omitempty"`
	ScopeClaims             map[string][]string           `json:"scope_claims,omitempty"`
	Certificate             *model.ApplicationCertificate `json:"certificate,omitempty"`
}

// oAuthTokenConfig represents the OAuth token configuration for JSON marshaling.
type oAuthTokenConfig struct {
	AccessToken *accessTokenConfig `json:"access_token,omitempty"`
	IDToken     *idTokenConfig     `json:"id_token,omitempty"`
}

// accessTokenConfig represents access token configuration for JSON marshaling.
type accessTokenConfig struct {
	ValidityPeriod int64    `json:"validity_period,omitempty"`
	UserAttributes []string `json:"user_attributes,omitempty"`
}

// idTokenConfig represents ID token configuration for JSON marshaling.
type idTokenConfig struct {
	ValidityPeriod int64    `json:"validity_period,omitempty"`
	UserAttributes []string `json:"user_attributes,omitempty"`
}

// userInfoConfig represents user info endpoint configuration for JSON marshaling.
type userInfoConfig struct {
	ResponseType   model.UserInfoResponseType `json:"response_type,omitempty"`
	UserAttributes []string                   `json:"user_attributes,omitempty"`
}

// applicationConfigDAO represents the gateway configuration for an application (no identity data).
// Identity data (name, description, clientId, credentials) is in the ENTITY table.
type applicationConfigDAO struct {
	ID                        string
	AuthFlowID                string
	RegistrationFlowID        string
	IsRegistrationFlowEnabled bool
	ThemeID                   string
	LayoutID                  string
	Assertion                 *model.AssertionConfig
	LoginConsent              *model.LoginConsentConfig
	AllowedEntityTypes        []string
	Properties                map[string]interface{}
	IsReadOnly                bool
}

// oauthConfigDAO represents the OAuth configuration for an application, keyed by entity ID.
type oauthConfigDAO struct {
	AppID       string
	OAuthConfig *oAuthConfig
}

// appJSON is the internal structure for marshaling/unmarshaling the APP_JSON column.
type appJSON struct {
	Assertion          *model.AssertionConfig    `json:"assertion,omitempty"`
	LoginConsent       *model.LoginConsentConfig `json:"login_consent,omitempty"`
	AllowedEntityTypes []string                  `json:"allowed_entity_types,omitempty"`
	Properties         map[string]interface{}    `json:"properties,omitempty"`
}

// applicationStoreInterface defines the interface for application gateway config persistence.
// Identity data (name, description, clientId, credentials) is NOT handled here — that's in the entity layer.
type applicationStoreInterface interface {
	CreateApplication(ctx context.Context, app applicationConfigDAO) error
	CreateOAuthConfig(ctx context.Context, entityID string, oauthConfigJSON json.RawMessage) error
	GetApplicationByID(ctx context.Context, id string) (*applicationConfigDAO, error)
	GetOAuthConfigByAppID(ctx context.Context, entityID string) (*oauthConfigDAO, error)
	GetApplicationList(ctx context.Context) ([]applicationConfigDAO, error)
	GetTotalApplicationCount(ctx context.Context) (int, error)
	UpdateApplication(ctx context.Context, app applicationConfigDAO) error
	UpdateOAuthConfig(ctx context.Context, entityID string, oauthConfigJSON json.RawMessage) error
	DeleteApplication(ctx context.Context, id string) error
	DeleteOAuthConfig(ctx context.Context, entityID string) error
	IsApplicationExists(ctx context.Context, id string) (bool, error)
	IsApplicationDeclarative(ctx context.Context, id string) bool
}

// applicationStore implements applicationStoreInterface for database persistence.
type applicationStore struct {
	dbProvider   provider.DBProviderInterface
	deploymentID string
}

var getDBProvider = provider.GetDBProvider

// newApplicationStore creates a new database-backed application store.
func newApplicationStore() (applicationStoreInterface, transaction.Transactioner, error) {
	dbProvider := getDBProvider()
	client, err := dbProvider.GetConfigDBClient()
	if err != nil {
		return nil, nil, err
	}

	transactioner, err := dbProvider.GetConfigDBTransactioner()
	if err != nil {
		return nil, nil, err
	}

	deploymentID := config.GetThunderRuntime().Config.Server.Identifier
	if _, err := client.QueryContext(context.Background(), queryGetApplicationCount, deploymentID); err != nil {
		return nil, nil, fmt.Errorf("failed to verify application table: %w", err)
	}

	return &applicationStore{
		dbProvider:   dbProvider,
		deploymentID: deploymentID,
	}, transactioner, nil
}

// marshalApplicationConfigDAO marshals the JSON fields of an applicationConfigDAO and returns
// the prepared values ready for a SQL statement.
func marshalApplicationConfigDAO(app applicationConfigDAO) (
	appJSONBytes interface{},
	isRegistrationEnabledStr string,
	themeID, layoutID interface{},
	err error,
) {
	aj := appJSON{
		Assertion:          app.Assertion,
		LoginConsent:       app.LoginConsent,
		AllowedEntityTypes: app.AllowedEntityTypes,
		Properties:         app.Properties,
	}
	appJSONBytes, err = marshalNullableJSON(aj)
	if err != nil {
		return nil, "", nil, nil, fmt.Errorf("failed to marshal app json: %w", err)
	}

	isRegistrationEnabledStr = utils.BoolToNumString(app.IsRegistrationFlowEnabled)

	if app.ThemeID != "" {
		themeID = app.ThemeID
	}
	if app.LayoutID != "" {
		layoutID = app.LayoutID
	}

	return appJSONBytes, isRegistrationEnabledStr, themeID, layoutID, nil
}

// CreateApplication creates a new application gateway config entry.
func (st *applicationStore) CreateApplication(ctx context.Context, app applicationConfigDAO) error {
	dbClient, err := st.dbProvider.GetConfigDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	appJSONBytes, isRegistrationEnabledStr, themeID, layoutID, marshalErr := marshalApplicationConfigDAO(app)
	if marshalErr != nil {
		return marshalErr
	}

	_, err = dbClient.ExecuteContext(ctx, queryCreateApplication,
		app.ID, app.AuthFlowID, app.RegistrationFlowID, isRegistrationEnabledStr,
		themeID, layoutID, appJSONBytes, st.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to insert application: %w", err)
	}
	return nil
}

// CreateOAuthConfig creates a new OAuth config entry for an application.
func (st *applicationStore) CreateOAuthConfig(ctx context.Context, entityID string,
	oauthConfigJSON json.RawMessage) error {
	dbClient, err := st.dbProvider.GetConfigDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	_, err = dbClient.ExecuteContext(ctx, queryCreateOAuthApplication, entityID, oauthConfigJSON, st.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to insert OAuth config: %w", err)
	}
	return nil
}

// GetApplicationByID retrieves application gateway config by entity ID.
func (st *applicationStore) GetApplicationByID(ctx context.Context, id string) (*applicationConfigDAO, error) {
	dbClient, err := st.dbProvider.GetConfigDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.QueryContext(ctx, queryGetApplicationByID, id, st.deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	if len(results) == 0 {
		return nil, model.ApplicationNotFoundError
	}
	return buildAppConfigFromRow(results[0])
}

// GetOAuthConfigByAppID retrieves OAuth config by entity ID.
func (st *applicationStore) GetOAuthConfigByAppID(ctx context.Context,
	entityID string) (*oauthConfigDAO, error) {
	dbClient, err := st.dbProvider.GetConfigDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.QueryContext(ctx, queryGetOAuthConfigByAppID, entityID, st.deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	if len(results) == 0 {
		return nil, model.ApplicationNotFoundError
	}
	return buildOAuthConfigFromRow(results[0])
}

// GetApplicationList retrieves all application gateway configs.
func (st *applicationStore) GetApplicationList(ctx context.Context) ([]applicationConfigDAO, error) {
	dbClient, err := st.dbProvider.GetConfigDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.QueryContext(ctx, queryGetApplicationList, st.deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	apps := make([]applicationConfigDAO, 0, len(results))
	for _, row := range results {
		app, err := buildAppConfigFromRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to build application from result row: %w", err)
		}
		apps = append(apps, *app)
	}
	return apps, nil
}

// GetTotalApplicationCount retrieves the total count of applications.
func (st *applicationStore) GetTotalApplicationCount(ctx context.Context) (int, error) {
	dbClient, err := st.dbProvider.GetConfigDBClient()
	if err != nil {
		return 0, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.QueryContext(ctx, queryGetApplicationCount, st.deploymentID)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) > 0 {
		if total, ok := results[0]["total"].(int64); ok {
			return int(total), nil
		}
		return 0, fmt.Errorf("failed to parse total count")
	}
	return 0, nil
}

// UpdateApplication updates an application's gateway config.
func (st *applicationStore) UpdateApplication(ctx context.Context, app applicationConfigDAO) error {
	dbClient, err := st.dbProvider.GetConfigDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	appJSONBytes, isRegistrationEnabledStr, themeID, layoutID, marshalErr := marshalApplicationConfigDAO(app)
	if marshalErr != nil {
		return marshalErr
	}

	_, err = dbClient.ExecuteContext(ctx, queryUpdateApplicationByID,
		app.ID, app.AuthFlowID, app.RegistrationFlowID, isRegistrationEnabledStr,
		themeID, layoutID, appJSONBytes, st.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to update application: %w", err)
	}
	return nil
}

// UpdateOAuthConfig updates OAuth config for an application.
func (st *applicationStore) UpdateOAuthConfig(ctx context.Context, entityID string,
	oauthConfigJSON json.RawMessage) error {
	dbClient, err := st.dbProvider.GetConfigDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	_, err = dbClient.ExecuteContext(ctx, queryUpdateOAuthConfigByAppID,
		entityID, oauthConfigJSON, st.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to update OAuth config: %w", err)
	}
	return nil
}

// DeleteApplication deletes an application by entity ID. Cascades to OAuth config.
func (st *applicationStore) DeleteApplication(ctx context.Context, id string) error {
	dbClient, err := st.dbProvider.GetConfigDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	_, err = dbClient.ExecuteContext(ctx, queryDeleteApplicationByID, id, st.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}
	return nil
}

// DeleteOAuthConfig deletes OAuth config for an application.
func (st *applicationStore) DeleteOAuthConfig(ctx context.Context, entityID string) error {
	dbClient, err := st.dbProvider.GetConfigDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	_, err = dbClient.ExecuteContext(ctx, queryDeleteOAuthConfigByAppID, entityID, st.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to delete OAuth config: %w", err)
	}
	return nil
}

// IsApplicationExists checks if an application exists by entity ID.
func (st *applicationStore) IsApplicationExists(ctx context.Context, id string) (bool, error) {
	dbClient, err := st.dbProvider.GetConfigDBClient()
	if err != nil {
		return false, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.QueryContext(ctx, queryCheckApplicationExistsByID, id, st.deploymentID)
	if err != nil {
		return false, fmt.Errorf("failed to execute existence check query: %w", err)
	}

	return parseBoolFromCount(results)
}

// IsApplicationDeclarative returns false for database store (all database applications are mutable).
func (st *applicationStore) IsApplicationDeclarative(_ context.Context, _ string) bool {
	return false
}

// --- Helper functions ---
// buildAppConfigFromRow constructs an applicationConfigDAO from a database result row.
func buildAppConfigFromRow(row map[string]interface{}) (*applicationConfigDAO, error) {
	appID, ok := row["id"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to parse id as string")
	}

	authFlowID := parseStringColumn(row, "auth_flow_id")
	regFlowID := parseStringColumn(row, "registration_flow_id")
	themeID := parseStringColumn(row, "theme_id")
	layoutID := parseStringColumn(row, "layout_id")

	isRegistrationFlowEnabled := false
	if val := parseStringOrBytesColumn(row, "is_registration_flow_enabled"); val != "" {
		isRegistrationFlowEnabled = utils.NumStringToBool(val)
	}

	app := &applicationConfigDAO{
		ID:                        appID,
		AuthFlowID:                authFlowID,
		RegistrationFlowID:        regFlowID,
		IsRegistrationFlowEnabled: isRegistrationFlowEnabled,
		ThemeID:                   themeID,
		LayoutID:                  layoutID,
	}

	// Parse APP_JSON column containing assertion, login consent, allowed entity types, and properties.
	if appJSONStr := parseJSONColumnString(row, "app_json"); appJSONStr != "" {
		var aj appJSON
		if err := json.Unmarshal([]byte(appJSONStr), &aj); err != nil {
			log.GetLogger().Debug("Failed to unmarshal app_json", log.Error(err))
		} else {
			app.Assertion = aj.Assertion
			app.LoginConsent = aj.LoginConsent
			app.AllowedEntityTypes = aj.AllowedEntityTypes
			app.Properties = aj.Properties
		}
	}

	return app, nil
}

// buildOAuthConfigFromRow constructs an oauthConfigDAO from a database result row.
func buildOAuthConfigFromRow(row map[string]interface{}) (*oauthConfigDAO, error) {
	appID, ok := row["app_id"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to parse app_id as string")
	}

	dao := &oauthConfigDAO{AppID: appID}

	if configStr := parseJSONColumnString(row, "oauth_config"); configStr != "" {
		var cfg oAuthConfig
		if err := json.Unmarshal([]byte(configStr), &cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal OAuth config JSON: %w", err)
		}
		dao.OAuthConfig = &cfg
	}

	return dao, nil
}

// extractAssertionConfigFromJSON extracts assertion configuration from JSON data.
func extractStringMapFromJSON(data map[string]interface{}, key string) map[string]string {
	val, exists := data[key]
	if !exists || val == nil {
		return nil
	}
	raw, ok := val.(map[string]interface{})
	if !ok {
		return nil
	}
	result := make(map[string]string, len(raw))
	for k, v := range raw {
		if s, ok := v.(string); ok {
			result[k] = s
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// marshalNullableJSON marshals a value to JSON, returning nil for nil/empty input.
func marshalNullableJSON(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if string(data) == "null" {
		return nil, nil
	}
	return data, nil
}

// parseStringColumn safely extracts a string from a result row, returning "" for nil.
func parseStringColumn(row map[string]interface{}, key string) string {
	if row[key] == nil {
		return ""
	}
	if s, ok := row[key].(string); ok {
		return s
	}
	return ""
}

// parseStringOrBytesColumn handles columns that may come as string or []byte.
func parseStringOrBytesColumn(row map[string]interface{}, key string) string {
	if row[key] == nil {
		return ""
	}
	switch v := row[key].(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}

// parseJSONColumnString extracts a JSON column value as a string.
func parseJSONColumnString(row map[string]interface{}, column string) string {
	val, exists := row[column]
	if !exists || val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}

// parseBoolFromCount parses a boolean from a COUNT(*) query result.
func parseBoolFromCount(results []map[string]interface{}) (bool, error) {
	if len(results) == 0 {
		return false, nil
	}
	count, ok := results[0]["count"].(int64)
	if !ok {
		return false, fmt.Errorf("failed to parse count from query result")
	}
	return count > 0, nil
}

// getOAuthConfigJSONBytes serializes an InboundAuthConfigProcessedDTO to the OAUTH_CONFIG JSON format.
func getOAuthConfigJSONBytes(inboundAuth model.InboundAuthConfigProcessedDTO) (json.RawMessage, error) {
	if inboundAuth.OAuthAppConfig == nil {
		return nil, nil
	}
	oa := inboundAuth.OAuthAppConfig
	cfg := oAuthConfig{
		RedirectURIs:            oa.RedirectURIs,
		GrantTypes:              utils.ConvertToStringSlice(oa.GrantTypes),
		ResponseTypes:           utils.ConvertToStringSlice(oa.ResponseTypes),
		TokenEndpointAuthMethod: string(oa.TokenEndpointAuthMethod),
		PKCERequired:            oa.PKCERequired,
		PublicClient:            oa.PublicClient,
		Scopes:                  oa.Scopes,
		ScopeClaims:             oa.ScopeClaims,
		Certificate:             oa.Certificate,
	}
	if oa.Token != nil {
		cfg.Token = &oAuthTokenConfig{}
		if oa.Token.AccessToken != nil {
			cfg.Token.AccessToken = &accessTokenConfig{
				ValidityPeriod: oa.Token.AccessToken.ValidityPeriod,
				UserAttributes: oa.Token.AccessToken.UserAttributes,
			}
		}
		if oa.Token.IDToken != nil {
			cfg.Token.IDToken = &idTokenConfig{
				ValidityPeriod: oa.Token.IDToken.ValidityPeriod,
				UserAttributes: oa.Token.IDToken.UserAttributes,
			}
		}
	}
	if oa.UserInfo != nil {
		cfg.UserInfo = &userInfoConfig{
			ResponseType:   oa.UserInfo.ResponseType,
			UserAttributes: oa.UserInfo.UserAttributes,
		}
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OAuth config JSON: %w", err)
	}
	return data, nil
}
