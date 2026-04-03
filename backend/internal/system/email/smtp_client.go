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

package email

import (
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
)

const (
	smtpLoggerComponentName = "SMTPEmailClient"
	smtpDialTimeout         = 30 * time.Second
)

// The newSMTPClient creates a new instance of smtpClient.
// It validates the configuration at creation time to avoid runtime errors.
func newSMTPClient(config smtpConfig) (EmailClientInterface, error) {
	if config.from == "" {
		return nil, ErrorInvalidSender
	}
	if !IsValidEmail(config.from) {
		return nil, ErrorInvalidSender
	}
	if strings.TrimSpace(config.host) == "" {
		return nil, ErrorInvalidHost
	}
	if config.port <= 0 {
		return nil, ErrorInvalidPort
	}
	if config.enableAuthentication {
		if strings.TrimSpace(config.username) == "" || strings.TrimSpace(config.password) == "" {
			return nil, ErrorInvalidCredentials
		}
	}
	return &smtpClient{
		config: config,
	}, nil
}

// NewSMTPClientFromConfig creates a new smtpClient using the global Thunder configuration.
// It reads the email.smtp section from the Thunder runtime config.
// Returns an error if the configuration is invalid (e.g., missing sender address)
// or if the runtime is not initialized.
func NewSMTPClientFromConfig() (EmailClientInterface, error) {
	emailConfig := config.GetThunderRuntime().Config.Email.SMTP

	enableStartTLS := true
	if emailConfig.EnableStartTLS != nil {
		enableStartTLS = *emailConfig.EnableStartTLS
	}

	enableAuth := true
	if emailConfig.EnableAuthentication != nil {
		enableAuth = *emailConfig.EnableAuthentication
	}

	return newSMTPClient(smtpConfig{
		host:                 emailConfig.Host,
		port:                 emailConfig.Port,
		username:             emailConfig.Username,
		password:             emailConfig.Password,
		from:                 emailConfig.FromAddress,
		useTLS:               enableStartTLS,
		enableAuthentication: enableAuth,
	})
}

func (c *smtpClient) Send(emailData EmailData) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, smtpLoggerComponentName))

	// 1. Validate, sanitize in place, and extract the flat envelope list
	allRecipients, error := c.validateAndProcessRecipients(&emailData)
	if error != nil {
		return error
	}

	logger.Debug("Sending email via SMTP",
		log.String("from", log.MaskString(c.config.from)),
		log.Int("recipientCount", len(emailData.To)))

	serverAddress := fmt.Sprintf("%s:%d", c.config.host, c.config.port)

	// 2. Build the message headers (now using the trimmed emailData.To and emailData.CC arrays)
	message := c.buildMessage(emailData)

	// 3. Send via SMTP
	if error := c.sendViaSMTP(serverAddress, allRecipients, message); error != nil {
		return error
	}

	logger.Debug("Email sent successfully")
	return nil
}

func (c *smtpClient) validateAndProcessRecipients(emailData *EmailData) ([]string, error) {
	var allRecipients []string
	hasRecipient := false

	// Inline helper to validate and clean a specific group of addresses
	processGroup := func(addresses []string) ([]string, error) {
		var cleaned []string
		for _, address := range addresses {
			trimmed := strings.TrimSpace(address)
			if trimmed == "" {
				return nil, fmt.Errorf("%w: recipient address cannot be empty", ErrorInvalidRecipient)
			}
			if !IsValidEmail(trimmed) {
				return nil, fmt.Errorf("%w: invalid recipient address '%s'", ErrorInvalidRecipient, trimmed)
			}
			cleaned = append(cleaned, trimmed)
			allRecipients = append(allRecipients, trimmed)
			hasRecipient = true
		}
		return cleaned, nil
	}

	var error error
	if emailData.To, error = processGroup(emailData.To); error != nil {
		return nil, error
	}
	if emailData.CC, error = processGroup(emailData.CC); error != nil {
		return nil, error
	}
	if emailData.BCC, error = processGroup(emailData.BCC); error != nil {
		return nil, error
	}

	if !hasRecipient {
		return nil, ErrorInvalidRecipient
	}

	// Reject CR/LF in Subject to prevent header injection.
	if strings.ContainsAny(emailData.Subject, "\r\n") {
		return nil, ErrorInvalidSubject
	}

	return allRecipients, nil
}

func (c *smtpClient) buildMessage(emailData EmailData) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("From: %s\r\n", c.config.from))

	if len(emailData.To) > 0 {
		builder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(emailData.To, ", ")))
	} else {
		builder.WriteString("To: undisclosed-recipients:;\r\n")
	}

	if len(emailData.CC) > 0 {
		builder.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(emailData.CC, ", ")))
	}

	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.QEncoding.Encode("utf-8", emailData.Subject)))
	builder.WriteString("MIME-Version: 1.0\r\n")

	if emailData.IsHTML {
		builder.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	} else {
		builder.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	}

	builder.WriteString("\r\n")
	builder.WriteString(emailData.Body)

	return builder.String()
}

func (c *smtpClient) sendViaSMTP(serverAddress string, recipients []string, message string) error {
	conn, error := net.DialTimeout("tcp", serverAddress, smtpDialTimeout)
	if error != nil {
		return fmt.Errorf("%w: %w", ErrorSMTPConnection, error)
	}

	client, error := smtp.NewClient(conn, c.config.host)
	if error != nil {
		_ = conn.Close()
		return fmt.Errorf("%w: %w", ErrorSMTPConnection, error)
	}
	defer func() {
		_ = client.Close()
	}()

	if c.config.useTLS {
		ok, _ := client.Extension("STARTTLS")
		if !ok {
			return fmt.Errorf("%w: STARTTLS not supported by server", ErrorSMTPConnection)
		}
		tlsConfig := &tls.Config{
			ServerName: c.config.host,
			MinVersion: tls.VersionTLS12,
		}
		if error := client.StartTLS(tlsConfig); error != nil {
			return fmt.Errorf("%w: %w", ErrorSMTPConnection, error)
		}
	}

	if c.config.enableAuthentication && c.config.username != "" && c.config.password != "" {
		if error := client.Auth(smtp.PlainAuth("", c.config.username, c.config.password, c.config.host)); error != nil {
			return fmt.Errorf("%w: %w", ErrorSMTPAuth, error)
		}
	}

	if error := client.Mail(c.config.from); error != nil {
		return fmt.Errorf("%w: %w", ErrorEmailSendFailed, error)
	}

	for _, recipient := range recipients {
		if error := client.Rcpt(recipient); error != nil {
			return fmt.Errorf("%w: %w", ErrorEmailSendFailed, error)
		}
	}

	writer, error := client.Data()
	if error != nil {
		return fmt.Errorf("%w: %w", ErrorEmailSendFailed, error)
	}
	if _, error := writer.Write([]byte(message)); error != nil {
		return fmt.Errorf("%w: %w", ErrorEmailSendFailed, error)
	}
	if error := writer.Close(); error != nil {
		return fmt.Errorf("%w: %w", ErrorEmailSendFailed, error)
	}

	if error := client.Quit(); error != nil {
		log.GetLogger().With(log.String(log.LoggerKeyComponentName, smtpLoggerComponentName)).
			Error("Failed to gracefully close SMTP client", log.Error(error))
	}

	return nil
}
