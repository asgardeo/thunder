/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

// Package model defines the data structures for the application module.
package model

// Application represents an application with its details.
type Application struct {
	ID                      string   `json:"id,omitempty"`
	OrgID                   string   `json:"org_id"`
	Name                    string   `json:"name"`
	Description             string   `json:"description"`
	ClientID                string   `json:"client_id"`
	ClientSecret            string   `json:"client_secret"`
	CallbackURLs            []string `json:"callback_url"`
	SupportedGrantTypes     []string `json:"supported_grant_types"`
	AuthFlowGraphID         string   `json:"auth_flow_graph_id,omitempty"`
	RegistrationFlowGraphID string   `json:"registration_flow_graph_id,omitempty"`
}
