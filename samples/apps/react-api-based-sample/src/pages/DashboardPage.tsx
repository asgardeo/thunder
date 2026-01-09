/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import { useMemo } from "react";
import { Box, Container, Typography } from "@wso2/oxygen-ui";
import { useNavigate } from "@tanstack/react-router";
import { decodeJwt } from "../utils/jwt";
import UserTable from "../components/UserTable";

interface TokenPayload {
  sub?: string;
  userType?: string;
  username?: string;
  ouName?: string;
  iss?: string;
  [key: string]: unknown;
}

function DashboardPage() {
  const navigate = useNavigate();

  const assertion = useMemo(() => {
    return sessionStorage.getItem("assertion");
  }, []);

  const decodedToken = useMemo(() => {
    if (!assertion) return null;
    return decodeJwt(assertion);
  }, [assertion]);

  if (!assertion || !decodedToken) {
    navigate({ to: "/" });
    return null;
  }

  const payload = decodedToken.payload as TokenPayload;
  const displayName = payload.username || payload.sub || "User";

  return (
    <Container maxWidth="lg" sx={{ px: { xs: 2, sm: 3 } }}>
      <Box sx={{ my: 4 }}>
        <Typography variant="h4" component="h1" fontWeight={600}>
          Dashboard
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Welcome, {displayName}! You are successfully authenticated.
        </Typography>
      </Box>

      <UserTable />
    </Container>
  );
}

export default DashboardPage;
