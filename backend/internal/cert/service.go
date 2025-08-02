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

// Package cert provides the implementation for managing certificates in the system.
package cert

import (
	"errors"

	"github.com/asgardeo/thunder/internal/cert/constants"
	"github.com/asgardeo/thunder/internal/cert/model"
	"github.com/asgardeo/thunder/internal/cert/store"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

const loggerComponentName = "CertificateService"

// CertificateServiceInterface defines the methods for certificate service operations.
type CertificateServiceInterface interface {
	GetCertificateByID(id string) (*model.Certificate, *serviceerror.ServiceError)
	GetCertificateByReference(refType constants.CertificateReferenceType, refID string) (
		*model.Certificate, *serviceerror.ServiceError)
	CreateCertificate(cert *model.Certificate) (*model.Certificate, *serviceerror.ServiceError)
	UpdateCertificateByID(id string, cert *model.Certificate) (*model.Certificate, *serviceerror.ServiceError)
	UpdateCertificateByReference(refType constants.CertificateReferenceType, refID string, cert *model.Certificate) (
		*model.Certificate, *serviceerror.ServiceError)
	DeleteCertificateByID(id string) *serviceerror.ServiceError
	DeleteCertificateByReference(refType constants.CertificateReferenceType, refID string) *serviceerror.ServiceError
}

// CertificateService implements the CertificateServiceInterface for managing certificates.
type CertificateService struct {
	Store store.CertificateStoreInterface
}

// NewCertificateService creates a new instance of CertificateService.
func NewCertificateService() CertificateServiceInterface {
	return &CertificateService{
		Store: store.NewCertificateStore(),
	}
}

// GetCertificateByID retrieves a certificate by its ID.
func (s *CertificateService) GetCertificateByID(id string) (*model.Certificate, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if id == "" {
		return nil, &constants.ErrorInvalidCertificateID
	}

	certObj, err := s.Store.GetCertificateByID(id)
	if err != nil {
		if errors.Is(err, constants.ErrCertificateNotFound) {
			return nil, &constants.ErrorCertificateNotFound
		}
		logger.Error("Failed to get certificate by ID", log.String("id", id), log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	return certObj, nil
}

// GetCertificateByReference retrieves a certificate by its reference type and ID.
func (s *CertificateService) GetCertificateByReference(refType constants.CertificateReferenceType,
	refID string) (*model.Certificate, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if !isValidReferenceType(refType) {
		return nil, &constants.ErrorInvalidReferenceType
	}
	if refID == "" {
		return nil, &constants.ErrorInvalidReferenceID
	}

	certObj, err := s.Store.GetCertificateByReference(refType, refID)
	if err != nil {
		if errors.Is(err, constants.ErrCertificateNotFound) {
			return nil, &constants.ErrorCertificateNotFound
		}
		logger.Error("Failed to get certificate by reference", log.String("refType", string(refType)),
			log.String("refID", refID), log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	return certObj, nil
}

// CreateCertificate creates a new certificate.
func (s *CertificateService) CreateCertificate(cert *model.Certificate) (*model.Certificate,
	*serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := validateCertificateForCreation(cert); err != nil {
		return nil, err
	}

	cert.ID = sysutils.GenerateUUID()
	err := s.Store.CreateCertificate(cert)
	if err != nil {
		logger.Error("Failed to create certificate", log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	return cert, nil
}

// UpdateCertificateByID updates an existing certificate by its ID.
func (s *CertificateService) UpdateCertificateByID(id string, cert *model.Certificate) (
	*model.Certificate, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if id == "" {
		return nil, &constants.ErrorInvalidCertificateID
	}
	if err := validateCertificate(cert); err != nil {
		return nil, err
	}

	err := s.Store.UpdateCertificateByID(id, cert)
	if err != nil {
		if errors.Is(err, constants.ErrCertificateNotFound) {
			return nil, &constants.ErrorCertificateNotFound
		}
		logger.Error("Failed to update certificate by ID", log.String("id", id), log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	return cert, nil
}

// UpdateCertificateByReference updates an existing certificate by its reference type and ID.
func (s *CertificateService) UpdateCertificateByReference(refType constants.CertificateReferenceType,
	refID string, cert *model.Certificate) (*model.Certificate, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if !isValidReferenceType(refType) {
		return nil, &constants.ErrorInvalidReferenceType
	}
	if refID == "" {
		return nil, &constants.ErrorInvalidReferenceID
	}
	if err := validateCertificate(cert); err != nil {
		return nil, err
	}

	err := s.Store.UpdateCertificateByReference(refType, refID, cert)
	if err != nil {
		if errors.Is(err, constants.ErrCertificateNotFound) {
			return nil, &constants.ErrorCertificateNotFound
		}
		logger.Error("Failed to update certificate by reference", log.String("refType", string(refType)),
			log.String("refID", refID), log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	return cert, nil
}

// DeleteCertificateByID deletes a certificate by its ID.
func (s *CertificateService) DeleteCertificateByID(id string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if id == "" {
		return &constants.ErrorInvalidCertificateID
	}

	err := s.Store.DeleteCertificateByID(id)
	if err != nil {
		logger.Error("Failed to delete certificate by ID", log.String("id", id), log.Error(err))
		return &constants.ErrorInternalServerError
	}

	return nil
}

// DeleteCertificateByReference deletes a certificate by its reference type and ID.
func (s *CertificateService) DeleteCertificateByReference(refType constants.CertificateReferenceType,
	refID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if !isValidReferenceType(refType) {
		return &constants.ErrorInvalidReferenceType
	}
	if refID == "" {
		return &constants.ErrorInvalidReferenceID
	}

	err := s.Store.DeleteCertificateByReference(refType, refID)
	if err != nil {
		logger.Error("Failed to delete certificate by reference", log.String("refType", string(refType)),
			log.String("refID", refID), log.Error(err))
		return &constants.ErrorInternalServerError
	}

	return nil
}

// isValidReferenceType checks if the provided reference type is valid.
func isValidReferenceType(refType constants.CertificateReferenceType) bool {
	switch refType {
	case constants.CertificateReferenceTypeApplication, constants.CertificateReferenceTypeIDP:
		return true
	default:
		return false
	}
}

// isValidCertificateType checks if the provided certificate type is valid.
func isValidCertificateType(certType constants.CertificateType) bool {
	switch certType {
	case constants.CertificateTypeNone, constants.CertificateTypeJWKS, constants.CertificateTypeJWKSURI:
		return true
	default:
		return false
	}
}

// validateCertificate checks if the provided certificate is valid.
func validateCertificate(cert *model.Certificate) *serviceerror.ServiceError {
	if cert == nil {
		return &constants.ErrorInvalidCertificateValue
	}
	if cert.ID == "" {
		return &constants.ErrorInvalidCertificateID
	}
	if cert.RefID == "" {
		return &constants.ErrorInvalidReferenceID
	}
	if !isValidReferenceType(cert.RefType) {
		return &constants.ErrorInvalidReferenceType
	}
	if !isValidCertificateType(cert.Type) {
		return &constants.ErrorInvalidCertificateType
	}
	if len(cert.Value) < 10 || len(cert.Value) > 4096 {
		return &constants.ErrorInvalidCertificateValue
	}
	return nil
}

// validateCertificateForCreation checks if the provided certificate is valid for creation.
func validateCertificateForCreation(cert *model.Certificate) *serviceerror.ServiceError {
	if cert == nil {
		return &constants.ErrorInvalidCertificateValue
	}
	if cert.RefID == "" {
		return &constants.ErrorInvalidReferenceID
	}
	if !isValidReferenceType(cert.RefType) {
		return &constants.ErrorInvalidReferenceType
	}
	if !isValidCertificateType(cert.Type) {
		return &constants.ErrorInvalidCertificateType
	}
	if len(cert.Value) < 10 || len(cert.Value) > 4096 {
		return &constants.ErrorInvalidCertificateValue
	}
	return nil
}
