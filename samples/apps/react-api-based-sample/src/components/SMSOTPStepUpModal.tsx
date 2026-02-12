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

import { useState, useEffect, useRef, useCallback } from "react";
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Dialog,
  DialogContent,
  DialogTitle,
  Stack,
  TextField,
  Typography,
} from "@wso2/oxygen-ui";
import { sendSMSOTP, verifySMSOTP } from "../utils/api";
import { getConfig } from "../config";

interface SMSOTPStepUpModalProps {
  open: boolean;
  mobileNumber: string;
  existingAssertion: string;
  onSuccess: (enrichedAssertion: string) => void;
  onClose: () => void;
  criticalOperationMessage?: string;
}

function SMSOTPStepUpModal({
  open,
  mobileNumber,
  existingAssertion,
  onSuccess,
  onClose,
  criticalOperationMessage,
}: SMSOTPStepUpModalProps) {
  const [otp, setOtp] = useState("");
  const [loading, setLoading] = useState(false);
  const [sendingOTP, setSendingOTP] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [sessionToken, setSessionToken] = useState<string | null>(null);
  const [otpSent, setOtpSent] = useState(false);
  const [resendCooldown, setResendCooldown] = useState(0);
  // Use refs to prevent double sends in React StrictMode
  const hasSentOTPRef = useRef(false);
  const isSendingRef = useRef(false);
  const sendPromiseRef = useRef<Promise<void> | null>(null);
  const cooldownIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const handleSendOTP = useCallback(async (): Promise<void> => {
    setSendingOTP(true);
    setError(null);
    setOtpSent(false);

    try {
      const { notificationSenderId } = getConfig();
      if (!notificationSenderId) {
        throw new Error("Notification sender ID not configured");
      }

      const response = await sendSMSOTP(notificationSenderId, mobileNumber);
      
      setSessionToken(response.session_token);
      setOtpSent(true);
      hasSentOTPRef.current = true;
      
      // Clear any existing cooldown interval
      if (cooldownIntervalRef.current) {
        clearInterval(cooldownIntervalRef.current);
      }

      // Set cooldown for resend (60 seconds)
      setResendCooldown(60);
      cooldownIntervalRef.current = setInterval(() => {
        setResendCooldown((prev) => {
          if (prev <= 1) {
            if (cooldownIntervalRef.current) {
              clearInterval(cooldownIntervalRef.current);
              cooldownIntervalRef.current = null;
            }
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to send OTP. Please try again."
      );
      // Reset ref on error so user can retry
      hasSentOTPRef.current = false;
      isSendingRef.current = false;
    } finally {
      setSendingOTP(false);
    }
  }, [mobileNumber]);

  useEffect(() => {
    // Only send OTP once when modal opens for the first time
    if (open) {
      // Check if we've already sent or are sending
      if (hasSentOTPRef.current || isSendingRef.current || sendPromiseRef.current) {
        return;
      }
      
      // Set refs immediately to prevent duplicate sends (even in StrictMode)
      isSendingRef.current = true;
      hasSentOTPRef.current = true;
      
      // Store the promise to prevent duplicate sends
      const sendPromise = handleSendOTP();
      sendPromiseRef.current = sendPromise;
      
      sendPromise.finally(() => {
        sendPromiseRef.current = null;
        isSendingRef.current = false;
      }).catch(() => {
        // Error already handled in handleSendOTP
      });
    } else {
      // Reset state when modal closes
      setOtp("");
      setError(null);
      setSessionToken(null);
      setOtpSent(false);
      setResendCooldown(0);
      hasSentOTPRef.current = false;
      isSendingRef.current = false;
      sendPromiseRef.current = null;
    }
    
    // Cleanup function to prevent sends if component unmounts
    return () => {
      if (!open) {
        hasSentOTPRef.current = false;
        isSendingRef.current = false;
        sendPromiseRef.current = null;
      }
      // Clear cooldown interval on cleanup
      if (cooldownIntervalRef.current) {
        clearInterval(cooldownIntervalRef.current);
        cooldownIntervalRef.current = null;
      }
    };
  }, [open, handleSendOTP]);

  const handleVerifyOTP = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!sessionToken) {
      setError("Session token not available. Please try again.");
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await verifySMSOTP(
        sessionToken,
        otp,
        existingAssertion
      );

      if (response.assertion) {
        onSuccess(response.assertion);
      } else {
        throw new Error("No assertion token received");
      }
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "OTP verification failed. Please try again."
      );
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    // Prevent closing during loading/verification
    if (loading) {
      return;
    }
    setOtp("");
    setError(null);
    setSessionToken(null);
    setOtpSent(false);
    setResendCooldown(0);
    hasSentOTPRef.current = false;
    isSendingRef.current = false;
    sendPromiseRef.current = null;
    // Clear cooldown interval
    if (cooldownIntervalRef.current) {
      clearInterval(cooldownIntervalRef.current);
      cooldownIntervalRef.current = null;
    }
    onClose();
  };

  // Mask mobile number for display
  const maskedMobile = mobileNumber
    ? mobileNumber.length > 4
      ? `${mobileNumber.slice(0, -4)}****`
      : "****"
    : "";

  return (
    <Dialog
      open={open}
      onClose={handleClose}
      maxWidth="sm"
      fullWidth
      disableEscapeKeyDown={loading}
    >
      <DialogTitle>
        <Typography variant="h6" component="div">
          Step-Up Authentication
        </Typography>
      </DialogTitle>
      <DialogContent>
        <Box sx={{ mt: 2 }}>
          {criticalOperationMessage && (
            <Alert severity="warning" sx={{ mb: 2 }}>
              <Typography variant="body2" fontWeight={600}>
                Critical Operation
              </Typography>
              <Typography variant="body2">
                {criticalOperationMessage}
              </Typography>
            </Alert>
          )}
          <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
            For enhanced security, please verify your identity with an OTP sent
            to your mobile number.
          </Typography>

          {sendingOTP && !otpSent && (
            <Alert severity="info" sx={{ mb: 2 }}>
              Sending OTP to {maskedMobile}...
            </Alert>
          )}

          {otpSent && (
            <Alert severity="success" sx={{ mb: 2 }}>
              OTP sent successfully to {maskedMobile}
            </Alert>
          )}

          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}

          <Box component="form" onSubmit={handleVerifyOTP}>
            <Stack spacing={2}>
              <TextField
                fullWidth
                label="Enter OTP"
                name="otp"
                type="text"
                value={otp}
                onChange={(e) => {
                  // Filter out non-numeric characters
                  const numericValue = e.target.value.replace(/[^0-9]/g, "");
                  setOtp(numericValue);
                }}
                placeholder="123456"
                required
                disabled={loading || !otpSent || sendingOTP}
                inputProps={{
                  maxLength: 6,
                  inputMode: "numeric",
                  pattern: "[0-9]*",
                  autoComplete: "one-time-code",
                }}
                helperText="Enter the 6-digit code sent to your mobile number"
              />

              <Stack direction="row" spacing={2}>
                <Button
                  type="submit"
                  variant="contained"
                  color="primary"
                  fullWidth
                  disabled={loading || !otpSent || sendingOTP || !otp}
                >
                  {loading ? (
                    <CircularProgress size={24} color="inherit" />
                  ) : (
                    "Verify OTP"
                  )}
                </Button>
              </Stack>

              <Button
                variant="text"
                size="small"
                onClick={handleSendOTP}
                disabled={sendingOTP || loading || resendCooldown > 0}
                sx={{ mt: 1 }}
              >
                {sendingOTP
                  ? "Sending..."
                  : resendCooldown > 0
                  ? `Resend OTP (${resendCooldown}s)`
                  : "Resend OTP"}
              </Button>
            </Stack>
          </Box>
        </Box>
      </DialogContent>
    </Dialog>
  );
}

export default SMSOTPStepUpModal;
