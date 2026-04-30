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

import React, {JSX} from 'react';
import Link from '@docusaurus/Link';
import {Box, Typography, Container, Card} from '@wso2/oxygen-ui';

function SDK({icon, to = '#', name}: {icon: string; name: string; to?: string}): JSX.Element {
  return (
    <Card
      component={Link}
      to={to}
      sx={{
        display: 'flex',
        alignItems: 'center',
        p: 1.5,
        border: '1px solid',
        borderColor: 'divider',
        textDecoration: 'none',
        cursor: 'pointer',
        transition: 'all 0.3s ease',
        color: 'inherit',
        '&:hover': {
          borderColor: 'primary.main',
          color: 'primary.main',
          textDecoration: 'none',
        },
      }}
    >
      <Box component="img" src={icon} alt={name} sx={{mr: 1.5, height: 28, width: 28}} />
      <Typography variant="body2" fontWeight={500}>
        {name}
      </Typography>
    </Card>
  );
}

export default function SDKs() {
  return (
    <Container
      maxWidth="lg"
      sx={{
        mb: 16,
        px: {xs: 2, sm: 4},
      }}
    >
      <Typography variant="overline" color="text.secondary" sx={{mb: 1, display: 'block', letterSpacing: 1.5}}>
        Developer Resources
      </Typography>

      <Typography variant="h3" fontWeight={600} sx={{mb: 6}}>
        Build with Thunder in your favorite stack
      </Typography>

      <Box sx={{mb: 5}}>
        <Typography variant="h5" fontWeight={600} sx={{mb: 1}}>
          Client Libraries
        </Typography>

        <Typography variant="body2" color="text.secondary" sx={{mb: 3}}>
          Official SDKs and libraries for integrating Thunder authentication into your applications.
        </Typography>

        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: {
              xs: '1fr',
              sm: 'repeat(2, 1fr)',
              md: 'repeat(4, 1fr)',
            },
            gap: 2,
          }}
        >
          <SDK name="JavaScript" to="/guides/setup/javascript" icon="/static/landing-page/sdk-icons/js.png" />
          <SDK name="React" to="/guides/setup/react" icon="/static/landing-page/sdk-icons/react.png" />
          <SDK name="Node.js" to="/guides/setup/nodejs" icon="/static/landing-page/sdk-icons/js.png" />
          <SDK name="Python" to="/guides/setup/python" icon="/static/landing-page/sdk-icons/js.png" />
          <SDK name="Java" to="/guides/setup/java" icon="/static/landing-page/sdk-icons/kotlin.png" />
          <SDK name="Go" to="/guides/setup/go" icon="/static/landing-page/sdk-icons/js.png" />
          <SDK name=".NET" to="/guides/setup/dotnet" icon="/static/landing-page/sdk-icons/js.png" />
          <SDK name="PHP" to="/guides/setup/php" icon="/static/landing-page/sdk-icons/js.png" />
        </Box>
      </Box>

      <Box>
        <Typography variant="h5" fontWeight={600} sx={{mb: 1}}>
          Mobile SDKs
        </Typography>

        <Typography variant="body2" color="text.secondary" sx={{mb: 3}}>
          Native authentication SDKs for mobile platforms with biometric and device security support.
        </Typography>

        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: {
              xs: '1fr',
              sm: 'repeat(2, 1fr)',
              md: 'repeat(4, 1fr)',
            },
            gap: 2,
          }}
        >
          <SDK name="iOS" to="/guides/setup/ios" icon="/static/landing-page/sdk-icons/swift.png" />
          <SDK name="Android" to="/guides/setup/android" icon="/static/landing-page/sdk-icons/kotlin.png" />
          <SDK name="React Native" to="/guides/setup/react-native" icon="/static/landing-page/sdk-icons/react.png" />
          <SDK name="Flutter" to="/guides/setup/flutter" icon="/static/landing-page/sdk-icons/flutter.png" />
        </Box>
      </Box>
    </Container>
  );
}
