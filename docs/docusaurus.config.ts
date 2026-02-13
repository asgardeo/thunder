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

import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';
import fs from 'fs';
import webpackPlugin from './plugins/webpackPlugin';
import thunderConfig from './docusaurus.thunder.config';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

const resourcesHTML = fs.readFileSync('./src/snippets/resources.html', 'utf-8');

const config: Config = {
  title: thunderConfig.project.name,
  tagline: thunderConfig.project.description,
  favicon: 'assets/images/favicon.ico',

  // Prevent search engine indexing
  // TODO: Remove this flag when the docs are ready for public access
  // Tracker: https://github.com/asgardeo/thunder/issues/1209
  noIndex: true,

  // Future flags, see https://docusaurus.io/docs/api/docusaurus-config#future
  future: {
    v4: true, // Improve compatibility with the upcoming Docusaurus v4
  },

  url: thunderConfig.documentation.deployment.production.url,
  // Since we use GitHub pages, the base URL is the repository name.
  baseUrl: `/${thunderConfig.documentation.deployment.production.baseUrl}/`,

  // GitHub pages deployment config.
  organizationName: thunderConfig.project.source.github.owner.name, // Usually your GitHub org/user name.
  projectName: thunderConfig.project.source.github.name, // Usually your repo name.

  onBrokenLinks: 'log',

  // Useful metadata like html lang for internationalization.
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  plugins: [webpackPlugin],

  presets: [
    [
      'classic',
      {
        docs: {
          path: 'content',
          sidebarPath: './sidebars.ts',
          // Edit URL for the "edit this page" feature.
          editUrl: thunderConfig.project.source.github.editUrls.content,
        },
        blog: {
          showReadingTime: true,
          feedOptions: {
            type: ['rss', 'atom'],
            xslt: true,
          },
          // Blog edit URL.
          editUrl: thunderConfig.project.source.github.editUrls.blog,
          // Useful options to enforce blogging best practices
          onInlineTags: 'warn',
          onInlineAuthors: 'warn',
          onUntruncatedBlogPosts: 'warn',
        },
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    announcementBar: {
      id: 'docs_wip',
      content: '🚧 WIP: Docs are under active development and may change frequently.',
      isCloseable: false,
    },
    image: 'assets/images/thunder-social-card.png',
    colorMode: {
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: '',
      logo: {
        href: '/',
        src: '/assets/images/logo.svg',
        srcDark: '/assets/images/logo-inverted.svg',
        alt: `${thunderConfig.project.name} Logo`,
        height: '40px',
        width: '101px',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docsSidebar',
          position: 'left',
          label: 'Docs',
          className: 'navbar__link--docs',
        },
        {
          to: '/apis',
          position: 'left',
          label: 'APIs',
        },
        {
          to: '/docs/sdks/overview',
          position: 'left',
          label: 'SDKs',
        },
        {
          label: 'Resources',
          type: 'dropdown',
          className: 'navbar__link--dropdown',
          items: [
            {
              type: 'html',
              value: resourcesHTML
                .replace('{{ISSUES_URL}}', thunderConfig.project.source.github.issuesUrl)
                .replace('{{DISCUSSIONS_URL}}', thunderConfig.project.source.github.discussionsUrl),
              className: 'navbar__link--dropdown-item',
            },
          ],
        },
        {
          type: 'docSidebar',
          sidebarId: 'communitySidebar',
          position: 'left',
          label: 'Community',
        },
        {
          href: `https://github.com/${thunderConfig.project.source.github.fullName}`,
          position: 'right',
          className: 'navbar__github--link',
          'aria-label': 'GitHub repository',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [],
      copyright: `Copyright © ${new Date().getFullYear()} ${thunderConfig.project.name}.`,
    },
    prism: {
      theme: prismThemes.nightOwlLight,
      darkTheme: prismThemes.nightOwl,
    },
  } satisfies Preset.ThemeConfig,

  /* -------------------------------- Thunder Config ------------------------------- */
  customFields: {
    thunder: thunderConfig,
  },
};

export default config;
