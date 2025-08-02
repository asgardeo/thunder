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

// Package store defines the storage interface and implementation for managing certificates.
package store

import (
	"errors"
	"fmt"

	"github.com/asgardeo/thunder/internal/cert/constants"
	"github.com/asgardeo/thunder/internal/cert/model"
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	dbprovider "github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

const loggerComponentName = "CertificateStore"

// CertificateStoreInterface defines the methods for certificate storage operations.
type CertificateStoreInterface interface {
	GetCertificateByID(id string) (*model.Certificate, error)
	GetCertificateByReference(refType constants.CertificateReferenceType, refID string) (*model.Certificate, error)
	CreateCertificate(cert *model.Certificate) error
	UpdateCertificateByID(id string, cert *model.Certificate) error
	UpdateCertificateByReference(refType constants.CertificateReferenceType, refID string,
		cert *model.Certificate) error
	DeleteCertificateByID(id string) error
	DeleteCertificateByReference(refType constants.CertificateReferenceType, refID string) error
}

// CertificateStore implements the CertificateStoreInterface for managing certificates.
type CertificateStore struct {
	DBProvider dbprovider.DBProviderInterface
}

// NewCertificateStore creates a new instance of CertificateStore.
func NewCertificateStore() CertificateStoreInterface {
	return &CertificateStore{
		DBProvider: dbprovider.NewDBProvider(),
	}
}

// GetCertificateByID retrieves a certificate by its ID.
func (s *CertificateStore) GetCertificateByID(id string) (*model.Certificate, error) {
	return s.getCertificate(QueryGetCertificateByID, id)
}

// GetCertificateByReference retrieves a certificate by its reference type and ID.
func (s *CertificateStore) GetCertificateByReference(refType constants.CertificateReferenceType, refID string) (
	*model.Certificate, error) {
	return s.getCertificate(QueryGetCertificateByReference, refType, refID)
}

// getCertificate retrieves a certificate based on a query and its arguments.
func (s *CertificateStore) getCertificate(query dbmodel.DBQuery, args ...interface{}) (*model.Certificate, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := s.DBProvider.GetDBClient("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	results, err := dbClient.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	if len(results) == 0 {
		return nil, constants.ErrCertificateNotFound
	}
	if len(results) > 1 {
		return nil, errors.New("multiple certificates found")
	}

	cert, err := s.buildCertificateFromResultRow(results[0])
	if err != nil {
		return nil, fmt.Errorf("failed to build certificate from result row: %w", err)
	}
	return cert, nil
}

// buildCertificateFromResultRow builds a Certificate object from a database result row.
func (s *CertificateStore) buildCertificateFromResultRow(row map[string]interface{}) (*model.Certificate, error) {
	certID, ok := row["cert_id"].(string)
	if !ok {
		return nil, errors.New("failed to parse cert_id as string")
	}

	refTypeStr, ok := row["ref_type"].(string)
	if !ok {
		return nil, errors.New("failed to parse ref_type as string")
	}
	refType := constants.CertificateReferenceType(refTypeStr)

	refID, ok := row["ref_id"].(string)
	if !ok {
		return nil, errors.New("failed to parse ref_id as string")
	}

	typeStr, ok := row["type"].(string)
	if !ok {
		return nil, errors.New("failed to parse type as string")
	}
	certType := constants.CertificateType(typeStr)

	value, ok := row["value"].(string)
	if !ok {
		return nil, errors.New("failed to parse value as string")
	}

	return &model.Certificate{
		ID:      certID,
		RefType: refType,
		RefID:   refID,
		Type:    certType,
		Value:   value,
	}, nil
}

// CreateCertificate creates a new certificate in the database.
func (s *CertificateStore) CreateCertificate(cert *model.Certificate) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := s.DBProvider.GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	rows, err := dbClient.Execute(QueryInsertCertificate, cert.ID, cert.RefType, cert.RefID, cert.Type, cert.Value)
	if err != nil {
		return fmt.Errorf("failed to insert certificate: %w", err)
	}
	if rows == 0 {
		return errors.New("no rows affected, certificate creation failed")
	}

	return nil
}

// UpdateCertificateByID updates a certificate by its ID.
func (s *CertificateStore) UpdateCertificateByID(id string, cert *model.Certificate) error {
	return s.updateCertificate(QueryUpdateCertificateByID, id, cert.Type, cert.Value)
}

// UpdateCertificateByReference updates a certificate by its reference type and ID.
func (s *CertificateStore) UpdateCertificateByReference(refType constants.CertificateReferenceType,
	refID string, cert *model.Certificate) error {
	return s.updateCertificate(QueryUpdateCertificateByReference, refType, refID, cert.Type, cert.Value)
}

// updateCertificate updates a certificate based on a query and its arguments.
func (s *CertificateStore) updateCertificate(query dbmodel.DBQuery, args ...interface{}) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := s.DBProvider.GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	rows, err := dbClient.Execute(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update certificate: %w", err)
	}
	if rows == 0 {
		return errors.New("no rows affected, certificate update failed")
	}

	return nil
}

// DeleteCertificateByID deletes a certificate by its ID.
func (s *CertificateStore) DeleteCertificateByID(id string) error {
	return s.deleteCertificate(QueryDeleteCertificateByID, id)
}

// DeleteCertificateByReference deletes a certificate by its reference type and ID.
func (s *CertificateStore) DeleteCertificateByReference(refType constants.CertificateReferenceType,
	refID string) error {
	return s.deleteCertificate(QueryDeleteCertificateByReference, refType, refID)
}

// deleteCertificate deletes a certificate based on a query and its arguments.
func (s *CertificateStore) deleteCertificate(query dbmodel.DBQuery, args ...interface{}) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := s.DBProvider.GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	_, err = dbClient.Execute(query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %w", err)
	}

	return nil
}
