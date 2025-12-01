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

package cert

import (
	"errors"
	"fmt"

	"github.com/asgardeo/thunder/internal/system/config"
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	dbprovider "github.com/asgardeo/thunder/internal/system/database/provider"
)

// certificateStoreInterface defines the methods for certificate storage operations.
type certificateStoreInterface interface {
	GetCertificateByID(id string) (*Certificate, error)
	GetCertificateByReference(refType CertificateReferenceType, refID string) (*Certificate, error)
	CreateCertificate(cert *Certificate) error
	UpdateCertificateByID(existingCert, updatedCert *Certificate) error
	UpdateCertificateByReference(existingCert, updatedCert *Certificate) error
	DeleteCertificateByID(id string) error
	DeleteCertificateByReference(refType CertificateReferenceType, refID string) error
}

// certificateStore implements the certificateStoreInterface for managing certificates.
type certificateStore struct {
	dbProvider   dbprovider.DBProviderInterface
	deploymentID string
}

// NewCertificateStore creates a new instance of CertificateStore.
func newCertificateStore() certificateStoreInterface {
	return &certificateStore{
		dbProvider:   dbprovider.GetDBProvider(),
		deploymentID: config.GetThunderRuntime().Config.Server.Identifier,
	}
}

// GetCertificateByID retrieves a certificate by its ID.
func (s *certificateStore) GetCertificateByID(id string) (*Certificate, error) {
	return s.getCertificate(queryGetCertificateByID, id, s.deploymentID)
}

// GetCertificateByReference retrieves a certificate by its reference type and ID.
func (s *certificateStore) GetCertificateByReference(refType CertificateReferenceType, refID string) (
	*Certificate, error) {
	return s.getCertificate(queryGetCertificateByReference, refType, refID, s.deploymentID)
}

// getCertificate retrieves a certificate based on a query and its arguments.
func (s *certificateStore) getCertificate(query dbmodel.DBQuery, args ...interface{}) (*Certificate, error) {
	dbClient, err := s.dbProvider.GetConfigDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	if len(results) == 0 {
		return nil, ErrCertificateNotFound
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
func (s *certificateStore) buildCertificateFromResultRow(row map[string]interface{}) (*Certificate, error) {
	certID, ok := row["cert_id"].(string)
	if !ok {
		return nil, errors.New("failed to parse cert_id as string")
	}

	refTypeStr, ok := row["ref_type"].(string)
	if !ok {
		return nil, errors.New("failed to parse ref_type as string")
	}
	refType := CertificateReferenceType(refTypeStr)

	refID, ok := row["ref_id"].(string)
	if !ok {
		return nil, errors.New("failed to parse ref_id as string")
	}

	typeStr, ok := row["type"].(string)
	if !ok {
		return nil, errors.New("failed to parse type as string")
	}
	certType := CertificateType(typeStr)

	value, ok := row["value"].(string)
	if !ok {
		return nil, errors.New("failed to parse value as string")
	}

	return &Certificate{
		ID:      certID,
		RefType: refType,
		RefID:   refID,
		Type:    certType,
		Value:   value,
	}, nil
}

// CreateCertificate creates a new certificate in the database.
func (s *certificateStore) CreateCertificate(cert *Certificate) error {
	dbClient, err := s.dbProvider.GetConfigDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	rows, err := dbClient.Execute(queryInsertCertificate, cert.ID, cert.RefType, cert.RefID, cert.Type, cert.Value,
		s.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to insert certificate: %w", err)
	}
	if rows == 0 {
		return errors.New("no rows affected, certificate creation failed")
	}

	return nil
}

// UpdateCertificateByID updates a certificate by its ID.
func (s *certificateStore) UpdateCertificateByID(existingCert, updatedCert *Certificate) error {
	return s.updateCertificate(queryUpdateCertificateByID, existingCert.ID, updatedCert.Type, updatedCert.Value,
		s.deploymentID)
}

// UpdateCertificateByReference updates a certificate by its reference type and ID.
func (s *certificateStore) UpdateCertificateByReference(existingCert, updatedCert *Certificate) error {
	return s.updateCertificate(queryUpdateCertificateByReference, existingCert.RefType, existingCert.RefID,
		updatedCert.Type, updatedCert.Value, s.deploymentID)
}

// updateCertificate updates a certificate based on a query and its arguments.
func (s *certificateStore) updateCertificate(query dbmodel.DBQuery, args ...interface{}) error {
	dbClient, err := s.dbProvider.GetConfigDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

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
func (s *certificateStore) DeleteCertificateByID(id string) error {
	return s.deleteCertificate(queryDeleteCertificateByID, id, s.deploymentID)
}

// DeleteCertificateByReference deletes a certificate by its reference type and ID.
func (s *certificateStore) DeleteCertificateByReference(refType CertificateReferenceType,
	refID string) error {
	return s.deleteCertificate(queryDeleteCertificateByReference, refType, refID, s.deploymentID)
}

// deleteCertificate deletes a certificate based on a query and its arguments.
func (s *certificateStore) deleteCertificate(query dbmodel.DBQuery, args ...interface{}) error {
	dbClient, err := s.dbProvider.GetConfigDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	_, err = dbClient.Execute(query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %w", err)
	}

	return nil
}
