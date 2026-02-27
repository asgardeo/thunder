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
import {Box, Container, Typography, Button} from '@wso2/oxygen-ui';

export default function CTASection(): JSX.Element {
  return (
    <Box sx={{py: {xs: 10, lg: 14}}}>
      <Container maxWidth="lg" sx={{px: {xs: 2, sm: 4}}}>
        <Box sx={{textAlign: 'center'}}>
          <Typography
            variant="h3"
            sx={{
              mb: 3,
              fontSize: {xs: '1.75rem', sm: '2.25rem', md: '2.75rem'},
              fontWeight: 700,
              color: '#ffffff',
            }}
          >
            Ready to start building?
          </Typography>
          <Typography
            variant="body1"
            sx={{
              mb: 5,
              maxWidth: '550px',
              mx: 'auto',
              fontSize: {xs: '0.95rem', sm: '1.05rem'},
              lineHeight: 1.7,
              color: 'rgba(255, 255, 255, 0.6)',
            }}
          >
            Launch your next project on Thunder&apos;s generous free tier. No credit card required.
          </Typography>
          <Button
            component={Link}
            href="/docs/guides/introduction"
            variant="contained"
            color="primary"
            size="large"
            sx={{
              px: 5,
              py: 1.5,
              fontWeight: 600,
              textTransform: 'none',
              fontSize: '1rem',
              borderRadius: 2,
              background: 'linear-gradient(135deg, #FF6B00 0%, #FF8C00 100%)',
              '&:hover': {
                background: 'linear-gradient(135deg, #e65e00 0%, #e67d00 100%)',
              },
            }}
          >
            Start building for <Box component="span" sx={{fontWeight: 800, ml: 0.5}}>FREE</Box>
          </Button>
        </Box>
      </Container>
    </Box>
  );
}
