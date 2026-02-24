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

import {Box, Card, Typography} from '@wso2/oxygen-ui';
import {ArrowUpRight} from '@wso2/oxygen-ui-icons-react';
import type {JSX, ReactNode} from 'react';

export interface ItemCardProps {
  thumbnail: ReactNode;
  name: string;
  onClick: () => void;
}

export default function ItemCard({thumbnail, name, onClick}: ItemCardProps): JSX.Element {
  return (
    <Card
      onClick={onClick}
      sx={{
        cursor: 'pointer',
        '&:hover': {
          borderColor: 'primary.main',
          boxShadow: '0 4px 20px rgba(0,0,0,0.1)',
          transform: 'translateY(-2px)',
          '& .card-overlay': {opacity: 1},
        },
      }}
    >
      <Box sx={{aspectRatio: '4/3', overflow: 'hidden', position: 'relative'}}>
        {thumbnail}
        {/* Hover overlay */}
        <Box
          className="card-overlay"
          sx={{
            position: 'absolute',
            inset: 0,
            bgcolor: 'rgba(0,0,0,0.35)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            opacity: 0,
            transition: 'opacity 0.18s ease',
            backdropFilter: 'blur(2px)',
          }}
        >
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              gap: 0.5,
              bgcolor: 'common.background',
              borderRadius: 2,
              px: 1.5,
              py: 0.75,
            }}
          >
            <ArrowUpRight size={13} />
            <Typography variant="caption" sx={{fontWeight: 600, fontSize: '0.75rem'}}>
              Open in builder
            </Typography>
          </Box>
        </Box>
      </Box>
      <Box sx={{px: 1.5, py: 1, borderTop: '1px solid', borderColor: 'divider'}}>
        <Typography
          variant="body2"
          sx={{
            fontWeight: 500,
            fontSize: '0.8125rem',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
        >
          {name}
        </Typography>
      </Box>
    </Card>
  );
}
