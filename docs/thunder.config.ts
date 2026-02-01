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

const thunderConfig = {
  project: {
    name: 'Thunder',
    description:
      'Thunder is a modern, open-source identity management service designed for teams building secure, customizable authentication experiences across applications, services, and AI agents.',
    documentation: {
      deployment: {
        production: {
          baseUrl: 'thunder',
          // TODO: Docusaurus doesn't seem to allow subpaths in the URL yet.
          // Can't use the GitHub pages URL until then.
          url: 'https://thunder.dev',
        },
      },
    },
    source: {
      github: {
        name: 'thunder',
        fullName: 'asgardeo/thunder',
        url: 'https://github.com/asgardeo/thunder',
        discussionsUrl: 'https://github.com/asgardeo/thunder/discussions',
        issuesUrl: 'https://github.com/asgardeo/thunder/issues',
        releasesUrl: 'https://github.com/asgardeo/thunder/releases',
        editUrls: {
          blog: 'https://github.com/asgardeo/thunder/tree/main/blog/',
          content: 'https://github.com/asgardeo/thunder/tree/main/content/',
        },
        owner: {
          name: 'asgardeo',
        },
      },
    },
  },
};

export default thunderConfig;
