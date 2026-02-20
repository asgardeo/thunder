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
import React from 'react';
import Head from '@docusaurus/Head';
import HeroSection from '@site/src/components/HomePage/HeroSection';
import SocialProofSection from '@site/src/components/HomePage/SocialProofSection';
import FeatureCardsSection from '@site/src/components/HomePage/FeatureCardsSection';
import GetStartedSection from '@site/src/components/HomePage/GetStartedSection';
import AIFlowsSection from '@site/src/components/HomePage/AIFlowsSection';
import UIComponentsSection from '@site/src/components/HomePage/UIComponentsSection';
import UseCasesSection from '@site/src/components/HomePage/UseCasesSection';
import CTASection from '@site/src/components/HomePage/CTASection';
import HomeFooter from '@site/src/components/Footer';

export default function Homepage(): ReactNode {
  const {siteConfig} = useDocusaurusContext();

  return (
    <Layout title={siteConfig.tagline} noFooter>
      <Head>
        <link rel="prefetch" href="/assets/css/elements.min.css" />
      </Head>
      <HeroSection />
      <SocialProofSection />
      <FeatureCardsSection />
      <GetStartedSection />
      <AIFlowsSection />
      <UIComponentsSection />
      <UseCasesSection />
      <CTASection />
      <HomeFooter />
    </Layout>
  );
}
