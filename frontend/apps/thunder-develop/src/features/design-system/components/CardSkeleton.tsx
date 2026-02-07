/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

import type {JSX} from 'react';
import {Card, CardContent, Grid, Skeleton, Stack} from '@mui/material';

/**
 * Props for the CardSkeleton component.
 */
export interface CardSkeletonProps {
  /** Number of skeleton cards to render */
  count?: number;
  /** Height of each skeleton card in pixels */
  height?: number;
}

/**
 * Loading skeleton component that matches the card grid layout.
 * Shows animated placeholder cards while content is loading.
 *
 * @param props - Component props
 * @returns CardSkeleton component
 *
 * @example
 * ```tsx
 * {isLoading ? (
 *   <CardSkeleton count={8} height={280} />
 * ) : (
 *   <Grid container spacing={3}>
 *     {items.map(item => <Card key={item.id}>...</Card>)}
 *   </Grid>
 * )}
 * ```
 */
export default function CardSkeleton({count = 8, height = 280}: CardSkeletonProps): JSX.Element {
  return (
    <Grid container spacing={3}>
      {Array.from({length: count}).map((_, index) => (
        <Grid item xs={12} sm={6} md={4} lg={3} key={index}>
          <Card sx={{height}}>
            <CardContent>
              <Stack spacing={2}>
                {/* Preview area */}
                <Skeleton variant="rectangular" height={height * 0.6} sx={{borderRadius: 2}} />

                {/* Title */}
                <Skeleton variant="text" width="80%" height={24} />

                {/* Metadata */}
                <Stack direction="row" spacing={1}>
                  <Skeleton variant="rectangular" width={60} height={24} sx={{borderRadius: 1}} />
                  <Skeleton variant="text" width={100} height={20} />
                </Stack>
              </Stack>
            </CardContent>
          </Card>
        </Grid>
      ))}
    </Grid>
  );
}
