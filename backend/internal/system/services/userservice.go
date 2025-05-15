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

package services

import (
	"github.com/asgardeo/thunder/internal/user/handler"
	"net/http"
)

type UserService struct {
	userHandler *handler.UserHandler
}

func NewUserService(mux *http.ServeMux) *UserService {

	instance := &UserService{
		userHandler: handler.NewUserHandler(),
	}
	instance.RegisterRoutes(mux)

	return instance
}

func (s *UserService) RegisterRoutes(mux *http.ServeMux) {

	mux.HandleFunc("POST /users", s.userHandler.HandleUserPostRequest)
	mux.HandleFunc("GET /users", s.userHandler.HandleUserListRequest)
	mux.HandleFunc("GET /users/", s.userHandler.HandleUserGetRequest)
	mux.HandleFunc("PUT /users/", s.userHandler.HandleUserPutRequest)
	mux.HandleFunc("DELETE /users/", s.userHandler.HandleUserDeleteRequest)
}
