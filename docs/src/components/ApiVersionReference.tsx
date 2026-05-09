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

import BrowserOnly from '@docusaurus/BrowserOnly';
import {useDocsVersion} from '@docusaurus/plugin-content-docs/client';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import {useEffect, useState} from 'react';
import {createPortal} from 'react-dom';
import ApiReference from './ApiReference';
import PostmanButton from './PostmanButton';

/**
 * Renders the API reference for the currently active Docusaurus doc version,
 * along with a sticky panel containing a Postman dropdown button.
 *
 * The combined OpenAPI spec is expected to live at:
 *   static/api/<versionPath>/combined.yaml
 *
 * The Postman collection is expected to live at:
 *   static/api/<versionPath>/postman/thunderid-api-postman-collection.json
 *
 * The version path follows the convention:
 *   - Docusaurus "current" version (labeled "Next") → 'next'
 *   - Any other version (e.g. '1.1.0') → the version name as-is
 *
 * This matches both the `path` values in docusaurus.config.ts `versions` config
 * and the directory names under static/api/.
 */

// Approximate height of Scalar's own toolbar row (Developer Tools / Configure / Share / Deploy).
const SCALAR_TOOLBAR_HEIGHT = 52;

export default function ApiVersionReference() {
  const {siteConfig} = useDocusaurusContext();
  const {version} = useDocsVersion();
  const [scalarScrolled, setScalarScrolled] = useState(false);
  const [clientPanelOpen, setClientPanelOpen] = useState(false);

  // Detect scroll inside the Scalar viewer to know when its toolbar is hidden.
  useEffect(() => {
    let scalarContainer: Element | null = null;
    let handleScroll: (() => void) | null = null;

    const timer = setTimeout(() => {
      scalarContainer = document.querySelector('.apis-page');
      if (!scalarContainer) return;

      handleScroll = () => setScalarScrolled(scalarContainer!.scrollTop > 10);
      scalarContainer.addEventListener('scroll', handleScroll, {passive: true});
    }, 300);

    return () => {
      clearTimeout(timer);
      if (scalarContainer && handleScroll) {
        scalarContainer.removeEventListener('scroll', handleScroll);
      }
    };
  }, []);

  // Use IntersectionObserver to detect when Scalar's Test Request panel is visible.
  // #scalar-client is always in the DOM but only intersects the viewport when open.
  useEffect(() => {
    let observer: IntersectionObserver | null = null;

    const timer = setTimeout(() => {
      const clientEl = document.getElementById('scalar-client');
      if (!clientEl) return;

      observer = new IntersectionObserver(([entry]) => setClientPanelOpen(entry.isIntersecting), {threshold: 0.1});

      observer.observe(clientEl);
    }, 500);

    return () => {
      clearTimeout(timer);
      observer?.disconnect();
    };
  }, []);

  const versionPath = version === 'current' ? 'next' : version;
  const productConfig = siteConfig.customFields?.product as {postman: {collection: {output: string}}};
  const specUrl = `${siteConfig.baseUrl}api/${versionPath}/combined.yaml`;
  const postmanCollectionUrl = `${siteConfig.baseUrl}api/${versionPath}/postman/collections/${productConfig.postman.collection.output}`;

  const topOffset = scalarScrolled ? 8 : SCALAR_TOOLBAR_HEIGHT + 8;

  return (
    <>
      <BrowserOnly>
        {() =>
          createPortal(
            <div
              style={{
                position: 'fixed',
                top: `calc(var(--ifm-navbar-height) + var(--docusaurus-announcement-bar-height) + ${topOffset}px)`,
                right: '40px',
                zIndex: 9999,
                transition: 'top 0.2s ease, opacity 0.15s ease',
                opacity: clientPanelOpen ? 0 : 1,
                pointerEvents: clientPanelOpen ? 'none' : 'auto',
              }}
            >
              <PostmanButton
                collectionUrl={postmanCollectionUrl}
                downloadFileName={productConfig.postman.collection.output}
              />
            </div>,
            document.body,
          )
        }
      </BrowserOnly>
      <ApiReference specUrl={specUrl} />
    </>
  );
}
