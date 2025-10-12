/**
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
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import Box from '@oxygen-ui/react/src/components/Box/Box';
import Typography from '@oxygen-ui/react/src/components/Typography/Typography';
import Link from 'next/link';
import { ReactElement } from 'react';

export default function NotFound(): ReactElement {
  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 2 }}>
        404 - Page Not Found
      </Typography>
      <Typography variant="body1">Sorry, the page you are looking for does not exist.</Typography>
      <Link href="/" className="text-blue-600 hover:underline">
        Go back home
      </Link>
    </Box>
  );
}
