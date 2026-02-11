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
	"context"
	"errors"

	"github.com/asgardeo/thunder/internal/system/cache"
	"github.com/asgardeo/thunder/internal/system/log"
)

const cacheBackedStoreLoggerComponentName = "CacheBackedCertificateStore"

// cacheBackedStore is the implementation of CertificateStoreInterface that uses caching.
type cacheBackedStore struct {
	certByIDCache        cache.CacheInterface[*Certificate]
	certByReferenceCache cache.CacheInterface[*Certificate]
	store                certificateStoreInterface
}

// NewCachedBackedCertificateStore creates a new instance of CachedBackedCertificateStore.
func newCachedBackedCertificateStore() certificateStoreInterface {
	return &cacheBackedStore{
		certByIDCache:        cache.GetCache[*Certificate]("CertificateByIDCache"),
		certByReferenceCache: cache.GetCache[*Certificate]("CertificateByReferenceCache"),
		store:                newCertificateStore(),
	}
}

// GetCertificateByID retrieves a certificate by its ID, using cache if available.
func (s *cacheBackedStore) GetCertificateByID(ctx context.Context, id string) (*Certificate, error) {
	cacheKey := cache.CacheKey{
		Key: id,
	}
	cachedCert, ok := s.certByIDCache.Get(cacheKey)
	if ok {
		return cachedCert, nil
	}

	cert, err := s.store.GetCertificateByID(ctx, id)
	if err != nil || cert == nil {
		return cert, err
	}
	s.cacheCertificate(cert)

	return cert, nil
}

// GetCertificateByReference retrieves a certificate by its reference type and ID, using cache if available.
func (s *cacheBackedStore) GetCertificateByReference(ctx context.Context, refType CertificateReferenceType,
	refID string) (*Certificate, error) {
	cacheKey := getCertByReferenceCacheKey(refType, refID)
	cachedCert, ok := s.certByReferenceCache.Get(cacheKey)
	if ok {
		return cachedCert, nil
	}

	cert, err := s.store.GetCertificateByReference(ctx, refType, refID)
	if err != nil || cert == nil {
		return cert, err
	}
	s.cacheCertificate(cert)

	return cert, nil
}

// CreateCertificate creates a new certificate and caches it.
func (s *cacheBackedStore) CreateCertificate(ctx context.Context, cert *Certificate) error {
	if err := s.store.CreateCertificate(ctx, cert); err != nil {
		return err
	}
	s.cacheCertificate(cert)
	return nil
}

// UpdateCertificateByID updates an existing certificate by its ID and refreshes the cache.
func (s *cacheBackedStore) UpdateCertificateByID(ctx context.Context, existingCert, updatedCert *Certificate) error {
	if err := s.store.UpdateCertificateByID(ctx, existingCert, updatedCert); err != nil {
		return err
	}

	// Invalidate old caches and cache the updated certificate
	s.invalidateCertificateCache(existingCert.ID, existingCert.RefType, existingCert.RefID)
	s.cacheCertificate(updatedCert)

	return nil
}

// UpdateCertificateByReference updates an existing certificate by its reference type and ID and refreshes the cache.
func (s *cacheBackedStore) UpdateCertificateByReference(ctx context.Context, existingCert,
	updatedCert *Certificate) error {
	if err := s.store.UpdateCertificateByReference(ctx, existingCert, updatedCert); err != nil {
		return err
	}

	// Invalidate old caches and cache the updated certificate
	s.invalidateCertificateCache(existingCert.ID, existingCert.RefType, existingCert.RefID)
	s.cacheCertificate(updatedCert)

	return nil
}

// DeleteCertificateByID deletes a certificate by its ID and invalidates the caches.
func (s *cacheBackedStore) DeleteCertificateByID(ctx context.Context, id string) error {
	cacheKey := cache.CacheKey{
		Key: id,
	}
	existingCert, ok := s.certByIDCache.Get(cacheKey)
	if !ok {
		var err error
		existingCert, err = s.store.GetCertificateByID(ctx, id)
		if err != nil {
			if errors.Is(err, ErrCertificateNotFound) {
				return nil
			}
			return err
		}
	}
	if existingCert == nil {
		return nil
	}

	if err := s.store.DeleteCertificateByID(ctx, id); err != nil {
		return err
	}
	s.invalidateCertificateCache(existingCert.ID, existingCert.RefType, existingCert.RefID)

	return nil
}

// DeleteCertificateByReference deletes a certificate by its reference type and ID and invalidates the caches.
func (s *cacheBackedStore) DeleteCertificateByReference(ctx context.Context, refType CertificateReferenceType,
	refID string) error {
	cacheKey := getCertByReferenceCacheKey(refType, refID)
	existingCert, ok := s.certByReferenceCache.Get(cacheKey)
	if !ok {
		var err error
		existingCert, err = s.store.GetCertificateByReference(ctx, refType, refID)
		if err != nil {
			if errors.Is(err, ErrCertificateNotFound) {
				return nil
			}
			return err
		}
	}
	if existingCert == nil {
		return nil
	}

	if err := s.store.DeleteCertificateByReference(ctx, refType, refID); err != nil {
		return err
	}
	s.invalidateCertificateCache(existingCert.ID, existingCert.RefType, existingCert.RefID)

	return nil
}

// cacheCertificate caches the certificate by ID and reference.
func (s *cacheBackedStore) cacheCertificate(cert *Certificate) {
	if cert == nil {
		return
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, cacheBackedStoreLoggerComponentName))

	// Cache by ID
	if cert.ID != "" {
		idCacheKey := cache.CacheKey{
			Key: cert.ID,
		}
		if err := s.certByIDCache.Set(idCacheKey, cert); err != nil {
			logger.Error("Failed to cache certificate by ID", log.Error(err),
				log.String("certID", cert.ID))
		} else {
			logger.Debug("Certificate cached by ID", log.String("certID", cert.ID))
		}
	}

	// Cache by reference type and ID
	if cert.RefType != "" && cert.RefID != "" {
		refCacheKey := getCertByReferenceCacheKey(cert.RefType, cert.RefID)
		if err := s.certByReferenceCache.Set(refCacheKey, cert); err != nil {
			logger.Error("Failed to cache certificate by reference", log.Error(err),
				log.String("refType", string(cert.RefType)), log.String("refID", cert.RefID))
		} else {
			logger.Debug("Certificate cached by reference", log.String("refType", string(cert.RefType)),
				log.String("refID", cert.RefID))
		}
	}
}

// invalidateCertificateCache invalidates all certificate caches for the given ID and reference.
func (s *cacheBackedStore) invalidateCertificateCache(id string,
	refType CertificateReferenceType, refID string) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, cacheBackedStoreLoggerComponentName))

	// Invalidate ID cache
	if id != "" {
		idCacheKey := cache.CacheKey{
			Key: id,
		}
		if err := s.certByIDCache.Delete(idCacheKey); err != nil {
			logger.Error("Failed to invalidate certificate cache by ID", log.Error(err),
				log.String("certID", id))
		} else {
			logger.Debug("Certificate cache invalidated by ID", log.String("certID", id))
		}
	}

	// Invalidate reference cache
	if refType != "" && refID != "" {
		refCacheKey := getCertByReferenceCacheKey(refType, refID)
		if err := s.certByReferenceCache.Delete(refCacheKey); err != nil {
			logger.Error("Failed to invalidate certificate cache by reference", log.Error(err),
				log.String("refType", string(refType)), log.String("refID", refID))
		} else {
			logger.Debug("Certificate cache invalidated by reference", log.String("refType", string(refType)),
				log.String("refID", refID))
		}
	}
}

// getCertByReferenceCacheKey generates a cache key for a certificate based on its reference type and ID.
func getCertByReferenceCacheKey(refType CertificateReferenceType, refID string) cache.CacheKey {
	return cache.CacheKey{
		Key: string(refType) + ":" + refID,
	}
}
