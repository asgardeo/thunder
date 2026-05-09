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

// integrateReactSDKInput represents the input for the integrate_react_sdk tool.
type integrateReactSDKInput struct {
	ServerURL string `json:"server_url,omitempty" jsonschema:"URL of the ThunderID server which is the baseurl for SDK"`
}

// integrateReactSDKOutput represents the output for the integrate_react_sdk tool.
type integrateReactSDKOutput struct {
	Instructions string `json:"instructions"`
	CodeSnippets string `json:"code_snippets"`
}
