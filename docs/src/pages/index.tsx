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

import type {ReactNode} from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import React, {useEffect} from 'react';
import Head from '@docusaurus/Head';
import HeroSection from '@site/src/components/HomePage/HeroSection';
import FeatureCardsSection from '@site/src/components/HomePage/FeatureCardsSection';
import GetStartedSection from '@site/src/components/HomePage/GetStartedSection';

import FlowsSection from '@site/src/components/HomePage/FlowsSection';
import UIComponentsSection from '@site/src/components/HomePage/UIComponentsSection';
import APIReferenceSection from '@site/src/components/HomePage/APIReferenceSection';
import CommunitySection from '@site/src/components/HomePage/CommunitySection';
import HomeFooter from '@site/src/components/Footer';

export default function Homepage(): ReactNode {
  const {siteConfig} = useDocusaurusContext();

  useEffect(() => {
    document.documentElement.setAttribute('data-page', 'home');

    return () => {
      document.documentElement.removeAttribute('data-page');
    };
  }, []);

  return (
    <Layout title={siteConfig.tagline} noFooter>
      <Head>
        <link rel="prefetch" href="/assets/css/elements.min.css" />
        <html data-page="home" />
      </Head>
      <div>
        <HeroSection />
        <FeatureCardsSection />
        <GetStartedSection />
        <FlowsSection />
        <UIComponentsSection />
        <APIReferenceSection />
        <CommunitySection />
        <HomeFooter />
      </div>
    </Layout>
  );
}
