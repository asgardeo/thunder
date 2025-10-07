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

// Package common contains the common models and constants for notification package.
package common

import "github.com/asgardeo/thunder/internal/system/cmodels"

// SMSData represents the data structure for a SMS message.
type SMSData struct {
	To   string `json:"to"`
	Body string `json:"body"`
}

// OTP represents the data structure for an OTP (One-Time Password).
type OTP struct {
	Value                  string `json:"value"`
	GeneratedTimeInMillis  int64  `json:"generated_time_in_millis"`
	ValidityPeriodInMillis int64  `json:"validity_period_in_millis"`
	ExpiryTimeInMillis     int64  `json:"expiry_time_in_millis"`
	AttemptCount           int    `json:"attempt_count"`
}

// NotificationSenderDTO represents the data transfer object for a notification sender.
type NotificationSenderDTO struct {
	ID          string
	Name        string
	Description string
	Type        NotificationSenderType
	Provider    MessageProviderType
	Properties  []cmodels.Property
}

// NotificationSenderRequest represents the request structure for creating or updating a notification sender.
type NotificationSenderRequest struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Provider    string                `json:"provider"`
	Properties  []cmodels.PropertyDTO `json:"properties"`
}

// NotificationSenderResponse represents the response structure for a notification sender.
type NotificationSenderResponse struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Provider    MessageProviderType   `json:"provider"`
	Properties  []cmodels.PropertyDTO `json:"properties"`
}

// SendOTPRequest represents the request structure for sending an OTP.
type SendOTPRequest struct {
	Recipient string `json:"recipient"`
	SenderID  string `json:"sender_id"`
	Channel   string `json:"channel"`
}

// SendOTPResponse represents the response structure for OTP send request.
type SendOTPResponse struct {
	SessionToken string `json:"session_token"`
	Status       string `json:"status"`
}

// VerifyOTPRequest represents the request structure for verifying an OTP.
type VerifyOTPRequest struct {
	SessionToken string `json:"session_token"`
	OTPCode      string `json:"otp_code"`
}

// VerifyOTPResponse represents the response structure for OTP verification.
type VerifyOTPResponse struct {
	Status string `json:"status"`
}

// SendOTPDTO represents the service layer data structure for sending an OTP.
type SendOTPDTO struct {
	Recipient string
	SenderID  string
	Channel   string
}

// SendOTPResultDTO represents the service layer result for OTP send operation.
type SendOTPResultDTO struct {
	SessionToken string
}

// VerifyOTPDTO represents the service layer data structure for verifying an OTP.
type VerifyOTPDTO struct {
	SessionToken string
	OTPCode      string
}

// VerifyOTPResultDTO represents the service layer result for OTP verify operation.
type VerifyOTPResultDTO struct {
	Status    OTPVerifyStatus
	Recipient string
}

// OTPSessionData represents the data stored in the OTP session token.
type OTPSessionData struct {
	Recipient  string `json:"recipient"`
	Channel    string `json:"channel"`
	SenderID   string `json:"sender_id"`
	OTPValue   string `json:"otp_value"`
	ExpiryTime int64  `json:"expiry_time"`
}
