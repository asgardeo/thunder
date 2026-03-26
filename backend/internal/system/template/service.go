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

package template

import (
	"context"
	"errors"
	"regexp"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

var ctxPlaceholderRegex = regexp.MustCompile(`\{\{ctx\((\w+)\)}}`)

// templateService implements TemplateServiceInterface using a templateStoreInterface.
type templateService struct {
	store  templateStoreInterface
	logger *log.Logger
}

// newTemplateService creates a new template service with the provided store.
func newTemplateService(store templateStoreInterface) TemplateServiceInterface {
	return &templateService{
		store:  store,
		logger: log.GetLogger().With(log.String(log.LoggerKeyComponentName, "TemplateService")),
	}
}

// GetTemplateByScenario retrieves a template for the specified scenario.
func (s *templateService) GetTemplateByScenario(
	ctx context.Context, scenario ScenarioType) (*TemplateDTO, *serviceerror.I18nServiceError) {
	s.logger.Debug("Retrieving template by scenario", log.String("scenario", string(scenario)))
	tmpl, err := s.store.GetTemplateByScenario(ctx, scenario)
	if err != nil {
		if errors.Is(err, errTemplateNotFound) {
			return nil, &ErrorTemplateNotFound
		}
		s.logger.Error("Failed to retrieve template by scenario",
			log.String("scenario", string(scenario)),
			log.Error(err))
		return nil, &serviceerror.InternalServerErrorWithI18n
	}

	return tmpl, nil
}

// Render renders a template for the specified scenario using the provided data.
func (s *templateService) Render(
	ctx context.Context, scenario ScenarioType, data TemplateData) (*RenderedTemplate, *serviceerror.I18nServiceError) {
	s.logger.Debug("Rendering template", log.String("scenario", string(scenario)))
	tmpl, svcErr := s.GetTemplateByScenario(ctx, scenario)
	if svcErr != nil {
		return nil, svcErr
	}

	replacePlaceholders := func(s string) string {
		return ctxPlaceholderRegex.ReplaceAllStringFunc(s, func(match string) string {
			// Extract the key from {{ctx(key)}}
			submatches := ctxPlaceholderRegex.FindStringSubmatch(match)
			if len(submatches) < 2 {
				return match
			}
			key := submatches[1]
			if val, ok := data[key]; ok {
				return val
			}
			return match
		})
	}

	s.logger.Debug("Template rendered successfully",
		log.String("scenario", string(scenario)),
		log.String("templateID", tmpl.ID))

	return &RenderedTemplate{
		Subject: replacePlaceholders(tmpl.Subject),
		Body:    replacePlaceholders(tmpl.Body),
		IsHTML:  tmpl.ContentType == "text/html",
	}, nil
}
