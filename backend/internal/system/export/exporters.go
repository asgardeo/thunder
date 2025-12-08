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

package export

import (
	"github.com/asgardeo/thunder/internal/application"
	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/notification"
	"github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userschema"
)

const (
	paramTypApplication        = "Application"
	paramTypIdentityProvider   = "IdentityProvider"
	paramTypNotificationSender = "NotificationSender"
	paramTypUserSchema         = "UserSchema"
)

// ApplicationExporter implements ResourceExporter for applications.
type ApplicationExporter struct {
	service application.ApplicationServiceInterface
}

// NewApplicationExporter creates a new application exporter.
func NewApplicationExporter(service application.ApplicationServiceInterface) *ApplicationExporter {
	return &ApplicationExporter{service: service}
}

// GetResourceType returns the resource type for applications.
func (e *ApplicationExporter) GetResourceType() string {
	return resourceTypeApplication
}

// GetParameterizerType returns the parameterizer type for applications.
func (e *ApplicationExporter) GetParameterizerType() string {
	return paramTypApplication
}

// GetAllResourceIDs retrieves all application IDs.
func (e *ApplicationExporter) GetAllResourceIDs() ([]string, *serviceerror.ServiceError) {
	apps, err := e.service.GetApplicationList()
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(apps.Applications))
	for _, app := range apps.Applications {
		ids = append(ids, app.ID)
	}
	return ids, nil
}

// GetResourceByID retrieves an application by its ID.
func (e *ApplicationExporter) GetResourceByID(id string) (interface{}, string, *serviceerror.ServiceError) {
	app, err := e.service.GetApplication(id)
	if err != nil {
		return nil, "", err
	}
	return app, app.Name, nil
}

// ValidateResource validates an application resource.
func (e *ApplicationExporter) ValidateResource(
	resource interface{}, id string, logger *log.Logger,
) (string, *ExportError) {
	app, ok := resource.(*appmodel.Application)
	if !ok {
		return "", createTypeError(resourceTypeApplication, id)
	}

	if err := validateResourceName(app.Name, resourceTypeApplication, id, "APP_VALIDATION_ERROR", logger); err != nil {
		return "", err
	}

	return app.Name, nil
}

// IDPExporter implements ResourceExporter for identity providers.
type IDPExporter struct {
	service idp.IDPServiceInterface
}

// NewIDPExporter creates a new IDP exporter.
func NewIDPExporter(service idp.IDPServiceInterface) *IDPExporter {
	return &IDPExporter{service: service}
}

// GetResourceType returns the resource type for identity providers.
func (e *IDPExporter) GetResourceType() string {
	return resourceTypeIdentityProvider
}

// GetParameterizerType returns the parameterizer type for identity providers.
func (e *IDPExporter) GetParameterizerType() string {
	return paramTypIdentityProvider
}

// GetAllResourceIDs retrieves all identity provider IDs.
func (e *IDPExporter) GetAllResourceIDs() ([]string, *serviceerror.ServiceError) {
	idps, err := e.service.GetIdentityProviderList()
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(idps))
	for _, idp := range idps {
		ids = append(ids, idp.ID)
	}
	return ids, nil
}

// GetResourceByID retrieves an identity provider by its ID.
func (e *IDPExporter) GetResourceByID(id string) (interface{}, string, *serviceerror.ServiceError) {
	idpDTO, err := e.service.GetIdentityProvider(id)
	if err != nil {
		return nil, "", err
	}
	return idpDTO, idpDTO.Name, nil
}

// ValidateResource validates an identity provider resource.
func (e *IDPExporter) ValidateResource(resource interface{}, id string, logger *log.Logger) (string, *ExportError) {
	idpDTO, ok := resource.(*idp.IDPDTO)
	if !ok {
		return "", createTypeError(resourceTypeIdentityProvider, id)
	}

	err := validateResourceName(
		idpDTO.Name, resourceTypeIdentityProvider, id, "IDP_VALIDATION_ERROR", logger,
	)
	if err != nil {
		return "", err
	}

	if len(idpDTO.Properties) == 0 {
		logger.Warn("Identity provider has no properties",
			log.String("idpID", id), log.String("name", idpDTO.Name))
	}

	return idpDTO.Name, nil
}

// NotificationSenderExporter implements ResourceExporter for notification senders.
type NotificationSenderExporter struct {
	service notification.NotificationSenderMgtSvcInterface
}

// NewNotificationSenderExporter creates a new notification sender exporter.
func NewNotificationSenderExporter(service notification.NotificationSenderMgtSvcInterface) *NotificationSenderExporter {
	return &NotificationSenderExporter{service: service}
}

// GetResourceType returns the resource type for notification senders.
func (e *NotificationSenderExporter) GetResourceType() string {
	return resourceTypeNotificationSender
}

// GetParameterizerType returns the parameterizer type for notification senders.
func (e *NotificationSenderExporter) GetParameterizerType() string {
	return paramTypNotificationSender
}

// GetAllResourceIDs retrieves all notification sender IDs.
func (e *NotificationSenderExporter) GetAllResourceIDs() ([]string, *serviceerror.ServiceError) {
	senders, err := e.service.ListSenders()
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(senders))
	for _, sender := range senders {
		ids = append(ids, sender.ID)
	}
	return ids, nil
}

// GetResourceByID retrieves a notification sender by its ID.
func (e *NotificationSenderExporter) GetResourceByID(id string) (interface{}, string, *serviceerror.ServiceError) {
	sender, err := e.service.GetSender(id)
	if err != nil {
		return nil, "", err
	}
	return sender, sender.Name, nil
}

// ValidateResource validates a notification sender resource.
func (e *NotificationSenderExporter) ValidateResource(
	resource interface{}, id string, logger *log.Logger,
) (string, *ExportError) {
	sender, ok := resource.(*common.NotificationSenderDTO)
	if !ok {
		return "", createTypeError(resourceTypeNotificationSender, id)
	}

	err := validateResourceName(
		sender.Name, resourceTypeNotificationSender, id, "SENDER_VALIDATION_ERROR", logger,
	)
	if err != nil {
		return "", err
	}

	if len(sender.Properties) == 0 {
		logger.Warn("Notification sender has no properties",
			log.String("senderID", id), log.String("name", sender.Name))
	}

	return sender.Name, nil
}

// UserSchemaExporter implements ResourceExporter for user schemas.
type UserSchemaExporter struct {
	service userschema.UserSchemaServiceInterface
}

// NewUserSchemaExporter creates a new user schema exporter.
func NewUserSchemaExporter(service userschema.UserSchemaServiceInterface) *UserSchemaExporter {
	return &UserSchemaExporter{service: service}
}

// GetResourceType returns the resource type for user schemas.
func (e *UserSchemaExporter) GetResourceType() string {
	return resourceTypeUserSchema
}

// GetParameterizerType returns the parameterizer type for user schemas.
func (e *UserSchemaExporter) GetParameterizerType() string {
	return paramTypUserSchema
}

// GetAllResourceIDs retrieves all user schema IDs.
func (e *UserSchemaExporter) GetAllResourceIDs() ([]string, *serviceerror.ServiceError) {
	response, err := e.service.GetUserSchemaList(0, 1000)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(response.Schemas))
	for _, schema := range response.Schemas {
		ids = append(ids, schema.ID)
	}
	return ids, nil
}

// GetResourceByID retrieves a user schema by its ID.
func (e *UserSchemaExporter) GetResourceByID(id string) (interface{}, string, *serviceerror.ServiceError) {
	schema, err := e.service.GetUserSchema(id)
	if err != nil {
		return nil, "", err
	}
	return schema, schema.Name, nil
}

// ValidateResource validates a user schema resource.
func (e *UserSchemaExporter) ValidateResource(
	resource interface{}, id string, logger *log.Logger,
) (string, *ExportError) {
	schema, ok := resource.(*userschema.UserSchema)
	if !ok {
		return "", createTypeError(resourceTypeUserSchema, id)
	}

	err := validateResourceName(
		schema.Name, resourceTypeUserSchema, id, "SCHEMA_VALIDATION_ERROR", logger,
	)
	if err != nil {
		return "", err
	}

	if len(schema.Schema) == 0 {
		logger.Warn("User schema has no schema definition",
			log.String("schemaID", id), log.String("name", schema.Name))
	}

	return schema.Name, nil
}
