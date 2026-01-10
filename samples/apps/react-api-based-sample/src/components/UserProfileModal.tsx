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

import { useMemo, useState } from "react";
import {
  Avatar,
  Box,
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  IconButton,
  Stack,
  Tooltip,
  Typography,
} from "@wso2/oxygen-ui";
import { decodeJwt } from "../utils/jwt";

interface TokenPayload {
  sub?: string;
  userType?: string;
  ouName?: string;
  iss?: string;
  username?: string;
  [key: string]: unknown;
}

interface UserProfileModalProps {
  open: boolean;
  onClose: () => void;
}

function UserProfileModal({ open, onClose }: UserProfileModalProps) {
  const [copied, setCopied] = useState(false);

  // Read assertion when modal opens
  const assertion = useMemo(() => {
    if (!open) return null;
    return sessionStorage.getItem("assertion");
  }, [open]);

  const decodedToken = useMemo(() => {
    if (!assertion) return null;
    return decodeJwt(assertion);
  }, [assertion]);

  if (!open || !assertion || !decodedToken) {
    return null;
  }

  const payload = decodedToken.payload as TokenPayload;

  const getUserInitials = () => {
    if (payload.username) {
      return String(payload.username).substring(0, 2).toUpperCase();
    }
    if (payload.sub) {
      return String(payload.sub).substring(0, 2).toUpperCase();
    }
    return "U";
  };

  const handleCopyToken = async () => {
    await navigator.clipboard.writeText(assertion);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogContent sx={{ p: 0 }}>
        <Box
          sx={{
            bgcolor: "primary.main",
            px: 3,
            py: 4,
            display: "flex",
            alignItems: "center",
            gap: 2,
          }}
        >
          <Avatar
            sx={{
              width: 64,
              height: 64,
              bgcolor: "primary.dark",
              fontSize: "1.5rem",
              fontWeight: 600,
            }}
          >
            {getUserInitials()}
          </Avatar>
          <Box sx={{ flex: 1 }}>
            <Typography
              variant="h6"
              color="primary.contrastText"
              fontWeight={600}
            >
              {payload.username || payload.sub || "User"}
            </Typography>
            <Box
              sx={{ display: "flex", alignItems: "center", gap: 1, mt: 0.5 }}
            >
              {payload.userType && (
                <Chip
                  label={payload.userType}
                  size="small"
                  sx={{
                    bgcolor: "rgba(255,255,255,0.2)",
                    color: "primary.contrastText",
                    textTransform: "capitalize",
                    fontWeight: 500,
                  }}
                />
              )}
              {payload.ouName && (
                <Typography
                  variant="body2"
                  color="primary.contrastText"
                  sx={{ opacity: 0.8 }}
                >
                  {payload.ouName}
                </Typography>
              )}
            </Box>
          </Box>
        </Box>

        <Box sx={{ p: 3 }}>
          <Stack spacing={3}>
            <Box>
              <Typography variant="subtitle2" fontWeight={600} sx={{ mb: 2 }}>
                User Details
              </Typography>
              <Stack spacing={0}>
                <InfoRow label="User ID" value={payload.sub} />
                <InfoRow label="User Type" value={payload.userType} />
                <InfoRow label="Organization" value={payload.ouName} />
                <InfoRow label="Issuer" value={payload.iss} />
              </Stack>
            </Box>

            <Box>
              <Box
                sx={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                  mb: 1,
                }}
              >
                <Typography variant="subtitle2" fontWeight={600}>
                  Assertion Token
                </Typography>
                <Tooltip title={copied ? "Copied!" : "Copy token"}>
                  <IconButton size="small" onClick={handleCopyToken}>
                    <Box component="span" sx={{ fontSize: "0.875rem" }}>
                      {copied ? "✓" : "⧉"}
                    </Box>
                  </IconButton>
                </Tooltip>
              </Box>
              <Box
                component="pre"
                sx={{
                  bgcolor: "action.hover",
                  color: "text.primary",
                  p: 2,
                  borderRadius: 1,
                  overflow: "auto",
                  fontSize: "0.7rem",
                  fontFamily: "monospace",
                  whiteSpace: "pre-wrap",
                  wordBreak: "break-all",
                  maxHeight: 100,
                  m: 0,
                  border: 1,
                  borderColor: "divider",
                }}
              >
                {assertion}
              </Box>
            </Box>

            <Box>
              <Typography variant="subtitle2" fontWeight={600} sx={{ mb: 1 }}>
                Decoded Payload
              </Typography>
              <Box
                component="pre"
                sx={{
                  bgcolor: "action.hover",
                  color: "text.primary",
                  p: 2,
                  borderRadius: 1,
                  overflow: "auto",
                  fontSize: "0.7rem",
                  fontFamily: "monospace",
                  maxHeight: 150,
                  m: 0,
                  border: 1,
                  borderColor: "divider",
                }}
              >
                {JSON.stringify(payload, null, 2)}
              </Box>
            </Box>
          </Stack>
        </Box>
      </DialogContent>
      <DialogActions
        sx={{ px: 3, py: 2, borderTop: 1, borderColor: "divider" }}
      >
        <Button onClick={onClose} variant="outlined">
          Close
        </Button>
      </DialogActions>
    </Dialog>
  );
}

function InfoRow({ label, value }: { label: string; value?: string }) {
  if (!value) return null;

  return (
    <Box
      sx={{
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
        py: 1.5,
        borderBottom: 1,
        borderColor: "divider",
        "&:last-child": {
          borderBottom: 0,
        },
      }}
    >
      <Typography variant="body2" color="text.secondary">
        {label}
      </Typography>
      <Typography
        variant="body2"
        fontWeight={500}
        sx={{
          maxWidth: "60%",
          overflow: "hidden",
          textOverflow: "ellipsis",
          whiteSpace: "nowrap",
        }}
      >
        {value}
      </Typography>
    </Box>
  );
}

export default UserProfileModal;
