/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import React from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import {useDocsVersion} from '@docusaurus/plugin-content-docs/client';
import ApiReference from './ApiReference';

/**
 * Renders the API reference for the currently active Docusaurus doc version,
 * along with a toolbar containing a download link for the Postman collection.
 *
 * The combined OpenAPI spec is expected to live at:
 *   static/api/<versionPath>/combined.yaml
 *
 * The Postman collection is expected to live at:
 *   static/api/<versionPath>/postman/thunder.json
 *
 * The version path follows the convention:
 *   - Docusaurus "current" version (labeled "Next") → 'next'
 *   - Any other version (e.g. '1.1.0') → the version name as-is
 *
 * This matches both the `path` values in docusaurus.config.ts `versions` config
 * and the directory names under static/api/.
 */
export default function ApiVersionReference() {
  const {siteConfig} = useDocusaurusContext();
  const {version} = useDocsVersion();

  // Map the Docusaurus internal version name to its URL path segment.
  // 'current' is the unreleased (Next) version, served under the 'next' path.
  const versionPath = version === 'current' ? 'next' : version;
  const specUrl = `${siteConfig.baseUrl}api/${versionPath}/combined.yaml`;
  const postmanCollectionUrl = `${siteConfig.baseUrl}api/${versionPath}/postman/thunder.json`;

  return (
    <div style={{position: 'relative', height: '100%'}}>
      <div
        style={{
          position: 'fixed',
          top: 'calc(var(--ifm-navbar-height) + var(--docusaurus-announcement-bar-height) + 12px)',
          right: '16px',
          zIndex: 100,
          display: 'flex',
          alignItems: 'center',
          gap: '8px',
        }}
      >
        <a
          href={postmanCollectionUrl}
          download="thunder.json"
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '6px',
            padding: '7px 14px',
            borderRadius: '6px',
            fontSize: '13px',
            fontWeight: 600,
            lineHeight: 1,
            textDecoration: 'none',
            color: '#ffffff',
            background: '#ff6c37',
            border: 'none',
            cursor: 'pointer',
            whiteSpace: 'nowrap',
            boxShadow: '0 1px 3px rgba(0,0,0,0.18)',
            transition: 'background 0.15s ease',
          }}
          onMouseEnter={(e) => {
            (e.currentTarget as HTMLAnchorElement).style.background = '#e5562a';
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLAnchorElement).style.background = '#ff6c37';
          }}
        >
          <svg
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="currentColor"
            aria-hidden="true"
          >
            <path d="M12 16l-6-6h4V4h4v6h4l-6 6zm6 2H6v2h12v-2z" />
          </svg>
          Download Postman Collection
        </a>
      </div>
      <ApiReference specUrl={specUrl} />
    </div>
  );
}
