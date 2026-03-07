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
	"net/mail"
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

// newSMTPClient creates a new instance of smtpClient.
// It validates the configuration at creation time to avoid runtime errors.
func newSMTPClient(config smtpConfig) (EmailClientInterface, error) {
	if config.from == "" {
		return nil, ErrInvalidSender
	}
	if _, err := mail.ParseAddress(config.from); err != nil {
		return nil, ErrInvalidSender
	}
	if strings.TrimSpace(config.host) == "" {
		return nil, ErrInvalidHost
	}
	if config.port <= 0 {
		return nil, ErrInvalidPort
	}
	if config.enableAuthentication {
		if strings.TrimSpace(config.username) == "" || strings.TrimSpace(config.password) == "" {
			return nil, ErrInvalidCredentials
		}
	}
	return &smtpClient{
		config: config,
	}, nil
}

// NewSMTPClientFromConfig creates a new smtpClient using the global Thunder configuration.
// It reads the email.smtp section from the Thunder runtime config.
// Returns an error if the configuration is invalid (e.g. missing sender address)
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

func (c *smtpClient) Send(email EmailData) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, smtpLoggerComponentName))

	if err := c.validateEmail(email); err != nil {
		return err
	}

	logger.Debug("Sending email via SMTP",
		log.String("from", log.MaskString(c.config.from)),
		log.Int("recipientCount", len(email.To)))

	addr := fmt.Sprintf("%s:%d", c.config.host, c.config.port)
	allRecipients := c.getRecipients(email)
	message := c.buildMessage(email)

	if err := c.sendViaSMTP(addr, allRecipients, message); err != nil {
		return err
	}

	logger.Debug("Email sent successfully")
	return nil
}

func (c *smtpClient) validateEmail(email EmailData) error {
	hasRecipient := false
	allAddresses := append(append(email.To, email.CC...), email.BCC...)
	for _, r := range allAddresses {
		trimmed := strings.TrimSpace(r)
		if trimmed == "" {
			return fmt.Errorf("%w: recipient address cannot be empty", ErrorInvalidRecipient)
		}
		hasRecipient = true
		if _, err := mail.ParseAddress(trimmed); err != nil {
			return fmt.Errorf("%w: invalid recipient address '%s': %w", ErrorInvalidRecipient, trimmed, err)
		}
	}

	if !hasRecipient {
		return ErrorInvalidRecipient
	}
	// Reject CR/LF in Subject to prevent header injection.
	if strings.ContainsAny(email.Subject, "\r\n") {
		return ErrInvalidSubject
	}
	return nil
}

func (c *smtpClient) getRecipients(email EmailData) []string {
	recipients := make([]string, 0, len(email.To)+len(email.CC)+len(email.BCC))
	for _, group := range [][]string{email.To, email.CC, email.BCC} {
		for _, r := range group {
			if trimmed := strings.TrimSpace(r); trimmed != "" {
				recipients = append(recipients, trimmed)
			}
		}
	}
	return recipients
}

func (c *smtpClient) buildMessage(email EmailData) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("From: %s\r\n", c.config.from))

	if len(email.To) > 0 {
		builder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ", ")))
	} else {
		builder.WriteString("To: undisclosed-recipients:;\r\n")
	}

	if len(email.CC) > 0 {
		builder.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(email.CC, ", ")))
	}

	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.QEncoding.Encode("utf-8", email.Subject)))
	builder.WriteString("MIME-Version: 1.0\r\n")

	if email.IsHTML {
		builder.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	} else {
		builder.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	}

	builder.WriteString("\r\n")
	builder.WriteString(email.Body)

	return builder.String()
}

func (c *smtpClient) sendViaSMTP(addr string, recipients []string, message string) error {
	conn, err := net.DialTimeout("tcp", addr, smtpDialTimeout)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSMTPConnection, err)
	}

	client, err := smtp.NewClient(conn, c.config.host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("%w: %w", ErrSMTPConnection, err)
	}
	defer func() {
		_ = client.Close()
	}()

	if c.config.useTLS {
		ok, _ := client.Extension("STARTTLS")
		if !ok {
			return fmt.Errorf("%w: STARTTLS not supported by server", ErrSMTPConnection)
		}
		tlsConfig := &tls.Config{
			ServerName: c.config.host,
			MinVersion: tls.VersionTLS12,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("%w: %w", ErrSMTPConnection, err)
		}
	}

	if c.config.enableAuthentication && c.config.username != "" && c.config.password != "" {
		if err := client.Auth(smtp.PlainAuth("", c.config.username, c.config.password, c.config.host)); err != nil {
			return fmt.Errorf("%w: %w", ErrSMTPAuth, err)
		}
	}

	if err := client.Mail(c.config.from); err != nil {
		return fmt.Errorf("%w: %w", ErrorEmailSendFailed, err)
	}

	for _, rcpt := range recipients {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("%w: %w", ErrorEmailSendFailed, err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrorEmailSendFailed, err)
	}
	if _, err := writer.Write([]byte(message)); err != nil {
		return fmt.Errorf("%w: %w", ErrorEmailSendFailed, err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("%w: %w", ErrorEmailSendFailed, err)
	}

	if err := client.Quit(); err != nil {
		log.GetLogger().With(log.String(log.LoggerKeyComponentName, smtpLoggerComponentName)).
			Error("Failed to gracefully close SMTP client", log.Error(err))
	}

	return nil
}
