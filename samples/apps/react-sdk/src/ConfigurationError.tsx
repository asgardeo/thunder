/*
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
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import {
  Box,
  Card,
  CardContent,
  Typography,
  List,
  ListItem,
  Alert,
  AlertTitle,
} from "@wso2/oxygen-ui";
import type { ReactElement } from "react";

interface ConfigurationErrorProps {
  missingConfig: string[];
}

export const ConfigurationError = ({
  missingConfig,
}: ConfigurationErrorProps): ReactElement => {
  return (
    <Card
      sx={{
        maxWidth: 600,
      }}
    >
      <CardContent sx={{ p: 4 }}>
        <Alert severity="info" sx={{ mb: 3 }}>
          <AlertTitle>Configuration Required</AlertTitle>
          Please configure the following environment variables to get started:
        </Alert>

        <List sx={{ bgcolor: "info.light", borderRadius: 1, mb: 2, py: 1 }}>
          {missingConfig.map((key) => (
            <ListItem key={key}>
              <Typography
                component="code"
                sx={{
                  fontFamily: "monospace",
                  bgcolor: "background.paper",
                  px: 1,
                  py: 0.5,
                  borderRadius: 0.5,
                  fontSize: "0.875rem",
                }}
              >
                {key}
              </Typography>
            </ListItem>
          ))}
        </List>

        <Typography variant="body1" color="text.secondary" sx={{ mb: 3 }}>
          Please create a <code>.env</code> file in the root directory and
          configure these variables. See <code>.env.example</code> for
          reference.
        </Typography>

        <Box
          sx={{
            bgcolor: "grey.50",
            borderLeft: 4,
            borderColor: "primary.main",
            borderRadius: 1,
            p: 2,
          }}
        >
          <Typography variant="subtitle2" sx={{ mb: 1, fontWeight: "bold" }}>
            Example:
          </Typography>
          <Box
            component="pre"
            sx={{
              m: 0,
              p: 1.5,
              bgcolor: "background.paper",
              borderRadius: 1,
              overflowX: "auto",
              fontSize: "0.875rem",
              fontFamily: "monospace",
            }}
          >
            {`VITE_REACT_APP_CLIENT_ID=your-client-id
VITE_THUNDER_BASE_URL=https://localhost:8090`}
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
};
