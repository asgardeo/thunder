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
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
)

const (
	cmdData = "DATA"
	cmdQuit = "QUIT"
)

type SMTPClientTestSuite struct {
	suite.Suite
}

func TestSMTPClientTestSuite(t *testing.T) {
	suite.Run(t, new(SMTPClientTestSuite))
}

func (suite *SMTPClientTestSuite) SetupSuite() {
	testConfig := &config.Config{}
	error := config.InitializeThunderRuntime("", testConfig)
	if error != nil {
		suite.T().Fatalf("Failed to initialize ThunderRuntime: %v", error)
	}
}

func (suite *SMTPClientTestSuite) getValidSMTPConfig(host string, port int) smtpConfig {
	return smtpConfig{
		host:                 host,
		port:                 port,
		username:             "testuser",
		password:             "testpass",
		from:                 "sender@example.com",
		useTLS:               false,
		enableAuthentication: true,
	}
}

// waitForDone waits for the mock server to finish with a timeout to avoid deadlocks.
func (suite *SMTPClientTestSuite) waitForDone(done <-chan bool) {
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		suite.T().Fatal("timed out waiting for mock SMTP server to finish")
	}
}

func (suite *SMTPClientTestSuite) TestNewSMTPClient_Success() {
	config := suite.getValidSMTPConfig("localhost", 25)

	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	suite.NotNil(client)
}

func (suite *SMTPClientTestSuite) TestNewSMTPClient_EmptyFrom_Error() {
	config := suite.getValidSMTPConfig("localhost", 25)
	config.from = ""

	client, error := newSMTPClient(config)
	suite.Error(error)
	suite.Nil(client)
	suite.True(errors.Is(error, ErrorInvalidSender))
}

func (suite *SMTPClientTestSuite) TestNewSMTPClient_InvalidFrom_Error() {
	config := suite.getValidSMTPConfig("localhost", 25)
	config.from = "invalid-email"

	client, error := newSMTPClient(config)
	suite.Error(error)
	suite.Nil(client)
	suite.True(errors.Is(error, ErrorInvalidSender))
}

func (suite *SMTPClientTestSuite) TestNewSMTPClient_EmptyHost_Error() {
	config := suite.getValidSMTPConfig("localhost", 25)
	config.host = "   "

	client, error := newSMTPClient(config)
	suite.Error(error)
	suite.Nil(client)
	suite.True(errors.Is(error, ErrorInvalidHost))
}

func (suite *SMTPClientTestSuite) TestNewSMTPClient_InvalidPort_Error() {
	config := suite.getValidSMTPConfig("localhost", 25)
	config.port = 0

	client, error := newSMTPClient(config)
	suite.Error(error)
	suite.Nil(client)
	suite.True(errors.Is(error, ErrorInvalidPort))

	config.port = -1
	client, error = newSMTPClient(config)
	suite.Error(error)
	suite.Nil(client)
	suite.True(errors.Is(error, ErrorInvalidPort))
}

func (suite *SMTPClientTestSuite) TestNewSMTPClient_EmptyCredentials_Error() {
	config := suite.getValidSMTPConfig("localhost", 25)
	config.enableAuthentication = true
	config.username = ""

	client, error := newSMTPClient(config)
	suite.Error(error)
	suite.Nil(client)
	suite.True(errors.Is(error, ErrorInvalidCredentials))

	config.username = "user"
	config.password = "  "
	client, error = newSMTPClient(config)
	suite.Error(error)
	suite.Nil(client)
	suite.True(errors.Is(error, ErrorInvalidCredentials))
}

func (suite *SMTPClientTestSuite) TestSendEmail_PlainText_Success() {
	listener, error := net.Listen("tcp", "127.0.0.1:0")
	suite.Require().NoError(error)
	defer func() { _ = listener.Close() }()

	serverAddress := listener.Addr().(*net.TCPAddr)
	config := suite.getValidSMTPConfig("127.0.0.1", serverAddress.Port)
	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	done := make(chan bool, 1)
	go suite.runMockSMTPServer(listener, done)

	emailData := EmailData{
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
		Body:    "Hello, this is a test email.",
		IsHTML:  false,
	}

	error = client.Send(emailData)

	suite.Require().NoError(error)
	suite.waitForDone(done)
}

func (suite *SMTPClientTestSuite) TestSendEmail_HTML_Success() {
	listener, error := net.Listen("tcp", "127.0.0.1:0")
	suite.Require().NoError(error)
	defer func() { _ = listener.Close() }()

	serverAddress := listener.Addr().(*net.TCPAddr)
	config := suite.getValidSMTPConfig("127.0.0.1", serverAddress.Port)
	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	done := make(chan bool, 1)
	go suite.runMockSMTPServer(listener, done)

	emailData := EmailData{
		To:      []string{"recipient@example.com"},
		Subject: "HTML Test",
		Body:    "<h1>Hello</h1><p>This is an HTML email.</p>",
		IsHTML:  true,
	}

	error = client.Send(emailData)

	suite.Require().NoError(error)
	suite.waitForDone(done)
}

func (suite *SMTPClientTestSuite) TestSendEmail_MultipleRecipients_Success() {
	listener, error := net.Listen("tcp", "127.0.0.1:0")
	suite.Require().NoError(error)
	defer func() { _ = listener.Close() }()

	serverAddress := listener.Addr().(*net.TCPAddr)
	config := suite.getValidSMTPConfig("127.0.0.1", serverAddress.Port)
	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	done := make(chan bool, 1)
	go suite.runMockSMTPServer(listener, done)

	emailData := EmailData{
		To:      []string{"to1@example.com", "to2@example.com"},
		CC:      []string{"cc@example.com"},
		BCC:     []string{"bcc@example.com"},
		Subject: "Multi-recipient Test",
		Body:    "Hello everyone!",
		IsHTML:  false,
	}

	error = client.Send(emailData)

	suite.Require().NoError(error)
	suite.waitForDone(done)
}

func (suite *SMTPClientTestSuite) TestSendEmail_EmptyToWithCC_Success() {
	listener, error := net.Listen("tcp", "127.0.0.1:0")
	suite.Require().NoError(error)
	defer func() { _ = listener.Close() }()

	serverAddress := listener.Addr().(*net.TCPAddr)
	config := suite.getValidSMTPConfig("127.0.0.1", serverAddress.Port)
	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	done := make(chan bool, 1)
	go suite.runMockSMTPServer(listener, done)

	emailData := EmailData{
		To:      []string{},
		CC:      []string{"cc@example.com"},
		BCC:     []string{"bcc@example.com"},
		Subject: "Undisclosed Recipients Test",
		Body:    "Hello everyone!",
		IsHTML:  false,
	}

	error = client.Send(emailData)

	suite.Require().NoError(error)
	suite.waitForDone(done)
}

func (suite *SMTPClientTestSuite) TestSendEmail_EmptyRecipients_Error() {
	config := suite.getValidSMTPConfig("localhost", 25)
	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	emailData := EmailData{
		To:      []string{},
		Subject: "Test",
		Body:    "Test body",
	}

	error = client.Send(emailData)

	suite.Error(error)
	suite.True(errors.Is(error, ErrorInvalidRecipient))
}

func (suite *SMTPClientTestSuite) TestSendEmail_EmptyRecipientString_Error() {
	config := suite.getValidSMTPConfig("localhost", 25)
	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	emailData := EmailData{
		To:      []string{"   "}, // empty after trim
		Subject: "Test",
		Body:    "Test body",
	}

	error = client.Send(emailData)

	suite.Error(error)
	suite.True(errors.Is(error, ErrorInvalidRecipient))
	suite.Contains(error.Error(), "recipient address cannot be empty")
}

func (suite *SMTPClientTestSuite) TestSendEmail_ConnectionError() {
	// Allocate an ephemeral port and close it immediately to ensure nothing is listening on it.
	listener, error := net.Listen("tcp", "127.0.0.1:0")
	suite.Require().NoError(error)
	serverAddress := listener.Addr().(*net.TCPAddr)
	port := serverAddress.Port
	error = listener.Close()
	suite.Require().NoError(error)

	config := suite.getValidSMTPConfig("127.0.0.1", port)
	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	emailData := EmailData{
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Test body",
	}

	error = client.Send(emailData)

	suite.Error(error)
	suite.True(errors.Is(error, ErrorSMTPConnection))
}

func (suite *SMTPClientTestSuite) TestSendEmail_CRLFInjection_Error() {
	config := suite.getValidSMTPConfig("localhost", 25)
	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	// Test CR/LF in recipient address.
	emailData := EmailData{
		To:      []string{"recipient@example.com\r\nBcc: evil@example.com"},
		Subject: "Test",
		Body:    "Test body",
	}
	error = client.Send(emailData)
	suite.Error(error)
	suite.True(errors.Is(error, ErrorInvalidRecipient))

	// Test CR/LF in subject.
	emailData = EmailData{
		To:      []string{"recipient@example.com"},
		Subject: "Test\r\nBcc: evil@example.com",
		Body:    "Test body",
	}
	error = client.Send(emailData)
	suite.Error(error)
	suite.Contains(error.Error(), "invalid characters")
	suite.True(errors.Is(error, ErrorInvalidSubject))
}

func (suite *SMTPClientTestSuite) TestBuildMessage_PlainText() {
	config := suite.getValidSMTPConfig("localhost", 25)
	ci, error := newSMTPClient(config)
	suite.Require().NoError(error)
	client := ci.(*smtpClient)

	emailData := EmailData{
		To:      []string{"to@example.com"},
		Subject: "Test Subject",
		Body:    "Hello, world!",
		IsHTML:  false,
	}

	message := client.buildMessage(emailData)

	suite.Contains(message, "From: sender@example.com")
	suite.Contains(message, "To: to@example.com")
	suite.Contains(message, "Content-Type: text/plain; charset=\"utf-8\"")
	suite.Contains(message, "MIME-Version: 1.0")
	suite.Contains(message, "Hello, world!")
}

func (suite *SMTPClientTestSuite) TestBuildMessage_HTMLWithCC() {
	config := suite.getValidSMTPConfig("localhost", 25)
	ci, error := newSMTPClient(config)
	suite.Require().NoError(error)
	client := ci.(*smtpClient)

	emailData := EmailData{
		To:      []string{"to@example.com"},
		CC:      []string{"cc1@example.com", "cc2@example.com"},
		Subject: "HTML Test",
		Body:    "<h1>Hello</h1>",
		IsHTML:  true,
	}

	message := client.buildMessage(emailData)

	suite.Contains(message, "Content-Type: text/html; charset=\"utf-8\"")
	suite.Contains(message, "Cc: cc1@example.com, cc2@example.com")
	suite.Contains(message, "<h1>Hello</h1>")
}

func (suite *SMTPClientTestSuite) TestValidateAndProcessRecipients() {
	config := suite.getValidSMTPConfig("localhost", 25)
	ci, error := newSMTPClient(config)
	suite.Require().NoError(error)
	client := ci.(*smtpClient)

	emailData := EmailData{
		To:      []string{" to@example.com "},
		CC:      []string{"cc@example.com"},
		BCC:     []string{" bcc@example.com "},
		Subject: "Test Subject",
	}

	recipients, error := client.validateAndProcessRecipients(&emailData)
	suite.Require().NoError(error)

	suite.Equal(3, len(recipients))
	suite.Contains(recipients, "to@example.com")
	suite.Contains(recipients, "cc@example.com")
	suite.Contains(recipients, "bcc@example.com")

	// Verify in-place updates
	suite.Equal("to@example.com", emailData.To[0])
	suite.Equal("bcc@example.com", emailData.BCC[0])
}

func (suite *SMTPClientTestSuite) TestValidateAndProcessRecipients_InvalidSubject_Error() {
	config := suite.getValidSMTPConfig("localhost", 25)
	ci, error := newSMTPClient(config)
	suite.Require().NoError(error)
	client := ci.(*smtpClient)

	emailData := EmailData{
		To:      []string{"to@example.com"},
		Subject: "Invalid\nSubject",
	}

	_, error = client.validateAndProcessRecipients(&emailData)
	suite.Error(error)
	suite.True(errors.Is(error, ErrorInvalidSubject))
}

func (suite *SMTPClientTestSuite) TestValidateAndProcessRecipients_InvalidCC_Error() {
	config := suite.getValidSMTPConfig("localhost", 25)
	ci, error := newSMTPClient(config)
	suite.Require().NoError(error)
	client := ci.(*smtpClient)

	emailData := EmailData{
		To: []string{"to@example.com"},
		CC: []string{"invalid-cc"},
	}

	_, error = client.validateAndProcessRecipients(&emailData)
	suite.Error(error)
	suite.True(errors.Is(error, ErrorInvalidRecipient))
}

func (suite *SMTPClientTestSuite) TestValidateAndProcessRecipients_InvalidBCC_Error() {
	config := suite.getValidSMTPConfig("localhost", 25)
	ci, error := newSMTPClient(config)
	suite.Require().NoError(error)
	client := ci.(*smtpClient)

	emailData := EmailData{
		To:  []string{"to@example.com"},
		BCC: []string{"invalid-bcc"},
	}

	_, error = client.validateAndProcessRecipients(&emailData)
	suite.Error(error)
	suite.True(errors.Is(error, ErrorInvalidRecipient))
}

// --- sendViaSMTP error path tests ---

func (suite *SMTPClientTestSuite) TestSendViaSMTP_InvalidGreeting_Error() {
	// A listener that sends an invalid SMTP greeting so smtp.NewClient fails.
	listener, error := net.Listen("tcp", "127.0.0.1:0")
	suite.Require().NoError(error)
	defer func() { _ = listener.Close() }()

	serverAddress := listener.Addr().(*net.TCPAddr)
	config := suite.getValidSMTPConfig("127.0.0.1", serverAddress.Port)
	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	done := make(chan bool, 1)
	go func() {
		defer func() { done <- true }()
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		// Send invalid greeting (not a 220).
		_, _ = fmt.Fprintf(conn, "500 Go away\r\n")
	}()

	emailData := EmailData{
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Test body",
	}

	error = client.Send(emailData)

	suite.Error(error)
	suite.True(errors.Is(error, ErrorSMTPConnection))
	suite.waitForDone(done)
}

func (suite *SMTPClientTestSuite) TestSendViaSMTP_TLSNotSupported_Error() {
	listener, error := net.Listen("tcp", "127.0.0.1:0")
	suite.Require().NoError(error)
	defer func() { _ = listener.Close() }()

	serverAddress := listener.Addr().(*net.TCPAddr)
	config := suite.getValidSMTPConfig("127.0.0.1", serverAddress.Port)
	config.useTLS = true // Enable STARTTLS
	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	done := make(chan bool, 1)
	// Standard mock server does not advertise STARTTLS
	go suite.runMockSMTPServer(listener, done)

	emailData := EmailData{
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Test body",
	}

	error = client.Send(emailData)

	suite.Error(error)
	suite.True(errors.Is(error, ErrorSMTPConnection))
	suite.Contains(error.Error(), "STARTTLS not supported by server")
	suite.waitForDone(done)
}

func (suite *SMTPClientTestSuite) TestSendViaSMTP_TLSError() {
	listener, error := net.Listen("tcp", "127.0.0.1:0")
	suite.Require().NoError(error)
	defer func() { _ = listener.Close() }()

	serverAddress := listener.Addr().(*net.TCPAddr)
	config := suite.getValidSMTPConfig("127.0.0.1", serverAddress.Port)
	config.useTLS = true // Enable STARTTLS
	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	done := make(chan bool, 1)
	go suite.runMockSMTPServerRejectTLS(listener, done)

	emailData := EmailData{
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Test body",
	}

	error = client.Send(emailData)

	suite.Error(error)
	suite.True(errors.Is(error, ErrorSMTPConnection))
	suite.waitForDone(done)
}

func (suite *SMTPClientTestSuite) TestSendViaSMTP_AuthError() {
	listener, error := net.Listen("tcp", "127.0.0.1:0")
	suite.Require().NoError(error)
	defer func() { _ = listener.Close() }()

	serverAddress := listener.Addr().(*net.TCPAddr)
	config := suite.getValidSMTPConfig("127.0.0.1", serverAddress.Port)
	config.enableAuthentication = true
	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	done := make(chan bool, 1)
	go suite.runMockSMTPServerRejectAuth(listener, done)

	emailData := EmailData{
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Test body",
	}

	error = client.Send(emailData)

	suite.Error(error)
	suite.True(errors.Is(error, ErrorSMTPAuth))
	suite.waitForDone(done)
}

func (suite *SMTPClientTestSuite) TestSendViaSMTP_MailFromError() {
	listener, error := net.Listen("tcp", "127.0.0.1:0")
	suite.Require().NoError(error)
	defer func() { _ = listener.Close() }()

	serverAddress := listener.Addr().(*net.TCPAddr)
	config := suite.getValidSMTPConfig("127.0.0.1", serverAddress.Port)
	config.enableAuthentication = false
	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	done := make(chan bool, 1)
	go suite.runMockSMTPServerRejectMailFrom(listener, done)

	emailData := EmailData{
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Test body",
	}

	error = client.Send(emailData)

	suite.Error(error)
	suite.True(errors.Is(error, ErrorEmailSendFailed))
	suite.waitForDone(done)
}

func (suite *SMTPClientTestSuite) TestSendViaSMTP_RcptToError() {
	listener, error := net.Listen("tcp", "127.0.0.1:0")
	suite.Require().NoError(error)
	defer func() { _ = listener.Close() }()

	serverAddress := listener.Addr().(*net.TCPAddr)
	config := suite.getValidSMTPConfig("127.0.0.1", serverAddress.Port)
	config.enableAuthentication = false
	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	done := make(chan bool, 1)
	go suite.runMockSMTPServerRejectRcptTo(listener, done)

	emailData := EmailData{
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Test body",
	}

	error = client.Send(emailData)

	suite.Error(error)
	suite.True(errors.Is(error, ErrorEmailSendFailed))
	suite.waitForDone(done)
}

func (suite *SMTPClientTestSuite) TestSendViaSMTP_DataError() {
	listener, error := net.Listen("tcp", "127.0.0.1:0")
	suite.Require().NoError(error)
	defer func() { _ = listener.Close() }()

	serverAddress := listener.Addr().(*net.TCPAddr)
	config := suite.getValidSMTPConfig("127.0.0.1", serverAddress.Port)
	config.enableAuthentication = false
	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	done := make(chan bool, 1)
	go suite.runMockSMTPServerRejectData(listener, done)

	emailData := EmailData{
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Test body",
	}

	error = client.Send(emailData)

	suite.Error(error)
	suite.True(errors.Is(error, ErrorEmailSendFailed))
	suite.waitForDone(done)
}

func (suite *SMTPClientTestSuite) TestSendViaSMTP_WriteError() {
	listener, error := net.Listen("tcp", "127.0.0.1:0")
	suite.Require().NoError(error)
	defer func() { _ = listener.Close() }()

	serverAddress := listener.Addr().(*net.TCPAddr)
	config := suite.getValidSMTPConfig("127.0.0.1", serverAddress.Port)
	config.enableAuthentication = false
	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	done := make(chan bool, 1)
	go suite.runMockSMTPServerCloseOnData(listener, done)

	// Use a large body (~1MB) to overflow TCP kernel buffers after the server closes
	// the connection. A small body would be silently buffered and the error would only
	// surface on writer.Close(), not writer.Write().
	largeBody := strings.Repeat("X", 1024*1024)

	emailData := EmailData{
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    largeBody,
	}

	error = client.Send(emailData)

	suite.Error(error)
	// On Windows, this is often reported as a connection failure (SYS-EMAIL-5001)
	errStr := error.Error()
	suite.True(errors.Is(error, ErrorEmailSendFailed) || errors.Is(error, ErrorSMTPConnection),
		"Expected error to wrap ErrorEmailSendFailed or ErrorSMTPConnection, but got: %s", errStr)
	suite.waitForDone(done)
}

func (suite *SMTPClientTestSuite) TestSendViaSMTP_DataTerminationError() {
	listener, error := net.Listen("tcp", "127.0.0.1:0")
	suite.Require().NoError(error)
	defer func() { _ = listener.Close() }()

	serverAddress := listener.Addr().(*net.TCPAddr)
	config := suite.getValidSMTPConfig("127.0.0.1", serverAddress.Port)
	config.enableAuthentication = false
	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	done := make(chan bool, 1)
	go suite.runMockSMTPServerRejectDataTermination(listener, done)

	emailData := EmailData{
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Test body",
	}

	error = client.Send(emailData)

	suite.Error(error)
	suite.True(errors.Is(error, ErrorEmailSendFailed))
	suite.waitForDone(done)
}

// --- NewSMTPClientFromConfig tests ---

func (suite *SMTPClientTestSuite) TestNewSMTPClientFromConfig_Defaults() {
	// Reset and set up config with nil boolean pointers (should default to true).
	config.ResetThunderRuntime()
	defer config.ResetThunderRuntime()

	testConfig := &config.Config{
		Email: config.EmailConfig{
			SMTP: config.SMTPEmailConfig{
				Host:                 "smtp.example.com",
				Port:                 587,
				Username:             "user@example.com",
				Password:             "secret",
				FromAddress:          "noreply@example.com",
				EnableStartTLS:       nil, // should default to true
				EnableAuthentication: nil, // should default to true
			},
		},
	}
	error := config.InitializeThunderRuntime("", testConfig)
	suite.Require().NoError(error)

	client, error := NewSMTPClientFromConfig()
	suite.Require().NoError(error)

	suite.NotNil(client)

	// Verify the internal config defaults.
	smtpCl, ok := client.(*smtpClient)
	suite.Require().True(ok)
	suite.True(smtpCl.config.useTLS, "useTLS should default to true when nil")
	suite.True(smtpCl.config.enableAuthentication, "enableAuthentication should default to true when nil")
	suite.Equal("smtp.example.com", smtpCl.config.host)
	suite.Equal(587, smtpCl.config.port)
	suite.Equal("user@example.com", smtpCl.config.username)
	suite.Equal("secret", smtpCl.config.password)
	suite.Equal("noreply@example.com", smtpCl.config.from)
}

func (suite *SMTPClientTestSuite) TestNewSMTPClientFromConfig_ExplicitFalse() {
	config.ResetThunderRuntime()
	defer config.ResetThunderRuntime()

	falseVal := false
	testConfig := &config.Config{
		Email: config.EmailConfig{
			SMTP: config.SMTPEmailConfig{
				Host:                 "smtp.example.com",
				Port:                 25,
				Username:             "",
				Password:             "",
				FromAddress:          "noreply@example.com",
				EnableStartTLS:       &falseVal,
				EnableAuthentication: &falseVal,
			},
		},
	}
	error := config.InitializeThunderRuntime("", testConfig)
	suite.Require().NoError(error)

	client, error := NewSMTPClientFromConfig()
	suite.Require().NoError(error)

	suite.NotNil(client)
	smtpCl, ok := client.(*smtpClient)
	suite.Require().True(ok)
	suite.False(smtpCl.config.useTLS, "useTLS should be false when explicitly set")
	suite.False(smtpCl.config.enableAuthentication, "enableAuthentication should be false when explicitly set")
}

func (suite *SMTPClientTestSuite) TestNewSMTPClientFromConfig_ExplicitTrue() {
	config.ResetThunderRuntime()
	defer config.ResetThunderRuntime()

	trueVal := true
	testConfig := &config.Config{
		Email: config.EmailConfig{
			SMTP: config.SMTPEmailConfig{
				Host:                 "smtp.example.com",
				Port:                 465,
				Username:             "user",
				Password:             "pass",
				FromAddress:          "noreply@example.com",
				EnableStartTLS:       &trueVal,
				EnableAuthentication: &trueVal,
			},
		},
	}
	error := config.InitializeThunderRuntime("", testConfig)
	suite.Require().NoError(error)

	client, error := NewSMTPClientFromConfig()
	suite.Require().NoError(error)

	suite.NotNil(client)
	smtpCl, ok := client.(*smtpClient)
	suite.Require().True(ok)
	suite.True(smtpCl.config.useTLS)
	suite.True(smtpCl.config.enableAuthentication)
}

// --- Test sending without authentication ---

func (suite *SMTPClientTestSuite) TestSendEmail_NoAuth_Success() {
	listener, error := net.Listen("tcp", "127.0.0.1:0")
	suite.Require().NoError(error)
	defer func() { _ = listener.Close() }()

	serverAddress := listener.Addr().(*net.TCPAddr)
	config := suite.getValidSMTPConfig("127.0.0.1", serverAddress.Port)
	config.enableAuthentication = false
	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	done := make(chan bool, 1)
	go suite.runMockSMTPServerNoAuth(listener, done)

	emailData := EmailData{
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
		Body:    "Hello, this is a test email.",
		IsHTML:  false,
	}

	error = client.Send(emailData)

	suite.Require().NoError(error)
	suite.waitForDone(done)
}

// --- Test QUIT error is ignored ---

func (suite *SMTPClientTestSuite) TestSendViaSMTP_QuitError_Ignored() {
	listener, error := net.Listen("tcp", "127.0.0.1:0")
	suite.Require().NoError(error)
	defer func() { _ = listener.Close() }()

	serverAddress := listener.Addr().(*net.TCPAddr)
	config := suite.getValidSMTPConfig("127.0.0.1", serverAddress.Port)
	config.enableAuthentication = false
	client, error := newSMTPClient(config)
	suite.Require().NoError(error)

	done := make(chan bool, 1)
	go suite.runMockSMTPServerRejectQuit(listener, done)

	emailData := EmailData{
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
		Body:    "Hello!",
		IsHTML:  false,
	}

	// QUIT errors should be ignored since the message was already accepted.
	error = client.Send(emailData)
	suite.NoError(error)
	suite.waitForDone(done)
}

// =============================================================================
// Mock SMTP Server Variants
// =============================================================================

// runMockSMTPServer runs a minimal mock SMTP server that accepts one connection.
func (suite *SMTPClientTestSuite) runMockSMTPServer(listener net.Listener, done chan bool) {
	defer func() { done <- true }()

	conn, error := listener.Accept()
	if error != nil {
		return
	}
	defer func() { _ = conn.Close() }()

	reader := bufio.NewReader(conn)

	_, _ = fmt.Fprintf(conn, "220 localhost SMTP Mock Server\r\n")

	for {
		line, error := reader.ReadString('\n')
		if error != nil {
			return
		}
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "EHLO") || strings.HasPrefix(line, "HELO"):
			_, _ = fmt.Fprintf(conn, "250-localhost\r\n")
			_, _ = fmt.Fprintf(conn, "250 AUTH PLAIN LOGIN\r\n")
		case strings.HasPrefix(line, "AUTH"):
			_, _ = fmt.Fprintf(conn, "235 Authentication successful\r\n")
		case strings.HasPrefix(line, "MAIL FROM:"):
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")
		case strings.HasPrefix(line, "RCPT TO:"):
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")
		case line == cmdData:
			_, _ = fmt.Fprintf(conn, "354 Start mail input\r\n")
			for {
				dataLine, dataErr := reader.ReadString('\n')
				if dataErr != nil {
					return
				}
				if strings.TrimSpace(dataLine) == "." {
					break
				}
			}
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")
		case line == cmdQuit:
			_, _ = fmt.Fprintf(conn, "221 Bye\r\n")
			return
		default:
			_, _ = fmt.Fprintf(conn, "500 Unrecognized command\r\n")
		}
	}
}

// runMockSMTPServerRejectTLS accepts EHLO but does not actually support STARTTLS,
// causing the TLS handshake to fail.
func (suite *SMTPClientTestSuite) runMockSMTPServerRejectTLS(listener net.Listener, done chan bool) {
	defer func() { done <- true }()

	conn, error := listener.Accept()
	if error != nil {
		return
	}
	defer func() { _ = conn.Close() }()

	reader := bufio.NewReader(conn)
	_, _ = fmt.Fprintf(conn, "220 localhost SMTP\r\n")

	for {
		line, error := reader.ReadString('\n')
		if error != nil {
			return
		}
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "EHLO") || strings.HasPrefix(line, "HELO"):
			_, _ = fmt.Fprintf(conn, "250-localhost\r\n")
			_, _ = fmt.Fprintf(conn, "250 STARTTLS\r\n")
		case line == "STARTTLS":
			_, _ = fmt.Fprintf(conn, "220 Ready to start TLS\r\n")
			return
		case line == cmdQuit:
			_, _ = fmt.Fprintf(conn, "221 Bye\r\n")
			return
		default:
			_, _ = fmt.Fprintf(conn, "500 Unrecognized command\r\n")
		}
	}
}

func (suite *SMTPClientTestSuite) runMockSMTPServerWithReject(
	listener net.Listener, done chan bool,
	ehloExtra string, rejectCmdPrefix string, rejectResponse string,
) {
	defer func() { done <- true }()

	conn, error := listener.Accept()
	if error != nil {
		return
	}
	defer func() { _ = conn.Close() }()

	reader := bufio.NewReader(conn)
	_, _ = fmt.Fprintf(conn, "220 localhost SMTP\r\n")

	for {
		line, error := reader.ReadString('\n')
		if error != nil {
			return
		}
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "EHLO") || strings.HasPrefix(line, "HELO"):
			_, _ = fmt.Fprintf(conn, "250-localhost\r\n")
			_, _ = fmt.Fprintf(conn, "%s\r\n", ehloExtra)
		case strings.HasPrefix(line, rejectCmdPrefix):
			_, _ = fmt.Fprintf(conn, "%s\r\n", rejectResponse)
			return
		case line == cmdQuit:
			_, _ = fmt.Fprintf(conn, "221 Bye\r\n")
			return
		default:
			_, _ = fmt.Fprintf(conn, "500 Unrecognized command\r\n")
		}
	}
}

// runMockSMTPServerRejectAuth accepts EHLO but rejects AUTH.
func (suite *SMTPClientTestSuite) runMockSMTPServerRejectAuth(listener net.Listener, done chan bool) {
	suite.runMockSMTPServerWithReject(listener, done, "250 AUTH PLAIN LOGIN", "AUTH", "535 Authentication failed")
}

// runMockSMTPServerRejectMailFrom accepts EHLO but rejects MAIL FROM.
func (suite *SMTPClientTestSuite) runMockSMTPServerRejectMailFrom(listener net.Listener, done chan bool) {
	suite.runMockSMTPServerWithReject(listener, done, "250 OK", "MAIL FROM:", "550 Sender rejected")
}

// runMockSMTPServerRejectRcptTo accepts MAIL FROM but rejects RCPT TO.
func (suite *SMTPClientTestSuite) runMockSMTPServerRejectRcptTo(listener net.Listener, done chan bool) {
	defer func() { done <- true }()

	conn, error := listener.Accept()
	if error != nil {
		return
	}
	defer func() { _ = conn.Close() }()

	reader := bufio.NewReader(conn)
	_, _ = fmt.Fprintf(conn, "220 localhost SMTP\r\n")

	for {
		line, error := reader.ReadString('\n')
		if error != nil {
			return
		}
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "EHLO") || strings.HasPrefix(line, "HELO"):
			_, _ = fmt.Fprintf(conn, "250-localhost\r\n")
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")
		case strings.HasPrefix(line, "MAIL FROM:"):
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")
		case strings.HasPrefix(line, "RCPT TO:"):
			_, _ = fmt.Fprintf(conn, "550 Recipient rejected\r\n")
			return
		case line == cmdQuit:
			_, _ = fmt.Fprintf(conn, "221 Bye\r\n")
			return
		default:
			_, _ = fmt.Fprintf(conn, "500 Unrecognized command\r\n")
		}
	}
}

// runMockSMTPServerRejectData accepts RCPT TO but rejects DATA.
func (suite *SMTPClientTestSuite) runMockSMTPServerRejectData(listener net.Listener, done chan bool) {
	defer func() { done <- true }()

	conn, error := listener.Accept()
	if error != nil {
		return
	}
	defer func() { _ = conn.Close() }()

	reader := bufio.NewReader(conn)
	_, _ = fmt.Fprintf(conn, "220 localhost SMTP\r\n")

	for {
		line, error := reader.ReadString('\n')
		if error != nil {
			return
		}
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "EHLO") || strings.HasPrefix(line, "HELO"):
			_, _ = fmt.Fprintf(conn, "250-localhost\r\n")
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")
		case strings.HasPrefix(line, "MAIL FROM:"):
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")
		case strings.HasPrefix(line, "RCPT TO:"):
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")
		case line == cmdData:
			_, _ = fmt.Fprintf(conn, "554 Transaction failed\r\n")
			return
		case line == cmdQuit:
			_, _ = fmt.Fprintf(conn, "221 Bye\r\n")
			return
		default:
			_, _ = fmt.Fprintf(conn, "500 Unrecognized command\r\n")
		}
	}
}

// runMockSMTPServerCloseOnData accepts DATA but closes the connection immediately,
// causing the client's writer.Write to fail.
func (suite *SMTPClientTestSuite) runMockSMTPServerCloseOnData(listener net.Listener, done chan bool) {
	defer func() { done <- true }()

	conn, error := listener.Accept()
	if error != nil {
		return
	}
	defer func() { _ = conn.Close() }()

	reader := bufio.NewReader(conn)
	_, _ = fmt.Fprintf(conn, "220 localhost SMTP\r\n")

	for {
		line, error := reader.ReadString('\n')
		if error != nil {
			return
		}
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "EHLO") || strings.HasPrefix(line, "HELO"):
			_, _ = fmt.Fprintf(conn, "250-localhost\r\n")
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")
		case strings.HasPrefix(line, "MAIL FROM:"):
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")
		case strings.HasPrefix(line, "RCPT TO:"):
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")
		case line == cmdData:
			_, _ = fmt.Fprintf(conn, "354 Start mail input\r\n")
			_ = conn.Close()
			return
		case line == cmdQuit:
			_, _ = fmt.Fprintf(conn, "221 Bye\r\n")
			return
		default:
			_, _ = fmt.Fprintf(conn, "500 Unrecognized command\r\n")
		}
	}
}

// runMockSMTPServerRejectDataTermination accepts data content but responds with
// an error when the data termination dot is sent, causing writer.Close to fail.
func (suite *SMTPClientTestSuite) runMockSMTPServerRejectDataTermination(listener net.Listener, done chan bool) {
	defer func() { done <- true }()

	conn, error := listener.Accept()
	if error != nil {
		return
	}
	defer func() { _ = conn.Close() }()

	reader := bufio.NewReader(conn)
	_, _ = fmt.Fprintf(conn, "220 localhost SMTP\r\n")

	for {
		line, error := reader.ReadString('\n')
		if error != nil {
			return
		}
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "EHLO") || strings.HasPrefix(line, "HELO"):
			_, _ = fmt.Fprintf(conn, "250-localhost\r\n")
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")
		case strings.HasPrefix(line, "MAIL FROM:"):
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")
		case strings.HasPrefix(line, "RCPT TO:"):
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")
		case line == cmdData:
			_, _ = fmt.Fprintf(conn, "354 Start mail input\r\n")
			for {
				dataLine, dataErr := reader.ReadString('\n')
				if dataErr != nil {
					return
				}
				if strings.TrimSpace(dataLine) == "." {
					break
				}
			}
			_, _ = fmt.Fprintf(conn, "554 Message rejected\r\n")
			return
		case line == cmdQuit:
			_, _ = fmt.Fprintf(conn, "221 Bye\r\n")
			return
		default:
			_, _ = fmt.Fprintf(conn, "500 Unrecognized command\r\n")
		}
	}
}

// runMockSMTPServerCustomQuit is like the standard mock but uses a custom QUIT response.
func (suite *SMTPClientTestSuite) runMockSMTPServerCustomQuit(
	listener net.Listener, done chan bool, quitResponse string,
) {
	defer func() { done <- true }()

	conn, error := listener.Accept()
	if error != nil {
		return
	}
	defer func() { _ = conn.Close() }()

	reader := bufio.NewReader(conn)
	_, _ = fmt.Fprintf(conn, "220 localhost SMTP\r\n")

	for {
		line, error := reader.ReadString('\n')
		if error != nil {
			return
		}
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "EHLO") || strings.HasPrefix(line, "HELO"):
			_, _ = fmt.Fprintf(conn, "250-localhost\r\n")
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")
		case strings.HasPrefix(line, "MAIL FROM:"):
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")
		case strings.HasPrefix(line, "RCPT TO:"):
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")
		case line == cmdData:
			_, _ = fmt.Fprintf(conn, "354 Start mail input\r\n")
			for {
				dataLine, dataErr := reader.ReadString('\n')
				if dataErr != nil {
					return
				}
				if strings.TrimSpace(dataLine) == "." {
					break
				}
			}
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")
		case line == cmdQuit:
			_, _ = fmt.Fprintf(conn, "%s\r\n", quitResponse)
			return
		default:
			_, _ = fmt.Fprintf(conn, "500 Unrecognized command\r\n")
		}
	}
}

// runMockSMTPServerNoAuth is like the standard mock but doesn't expect AUTH.
func (suite *SMTPClientTestSuite) runMockSMTPServerNoAuth(listener net.Listener, done chan bool) {
	suite.runMockSMTPServerCustomQuit(listener, done, "221 Bye")
}

// runMockSMTPServerRejectQuit is like the standard mock but returns an error for QUIT.
func (suite *SMTPClientTestSuite) runMockSMTPServerRejectQuit(listener net.Listener, done chan bool) {
	suite.runMockSMTPServerCustomQuit(listener, done, "500 QUIT error")
}

// --- Test runtime not initialized ---

func (suite *SMTPClientTestSuite) TestNewSMTPClientFromConfig_RuntimeNotInitialized() {
	// Reset the runtime to simulate an uninitialized state.
	config.ResetThunderRuntime()
	defer config.ResetThunderRuntime()

	// Should panic if runtime is not initialized.
	suite.PanicsWithValue("ThunderRuntime is not initialized", func() {
		_, _ = NewSMTPClientFromConfig()
	})

	// Re-initialize the runtime for subsequent tests.
	testConfig := &config.Config{}
	initErr := config.InitializeThunderRuntime("", testConfig)
	suite.Require().NoError(initErr)
}

// TestSendLiveEmail is a manual test utility to verify email delivery against real SMTP credentials.
// It loads the configuration from backend/cmd/server/repository/conf/deployment.yaml and attempts to send
// a test email to the specified address.
//
// By default, this test is skipped during normal test execution.
// To run this test manually, use the following command from the `backend` directory:
//
//	go test ./internal/system/email -v -run TestSMTPClientTestSuite/TestSendLiveEmail
func (suite *SMTPClientTestSuite) TestSendLiveEmail() {
	suite.T().Skip("Skipping live email test. To run, comment this line and use: " +
		"go test ./internal/system/email -v -run TestSMTPClientTestSuite/TestSendLiveEmail")

	config.ResetThunderRuntime()

	emailConfig, error := config.LoadConfig(
		"../../../cmd/server/repository/conf/deployment.yaml",
		"",
		"../../../cmd/server",
	)
	suite.Require().NoError(error, "Failed to load config")

	error = config.InitializeThunderRuntime("", emailConfig)
	suite.Require().NoError(error, "Failed to initialize thunder runtime")
	defer config.ResetThunderRuntime()

	client, error := NewSMTPClientFromConfig()
	suite.Require().NoError(error)

	emailData := EmailData{
		To:      []string{"test@example.com"},
		Subject: "Thunder Email System Test",
		Body: "<h1>Thunder Email System is Working!</h1>" +
			"<p>This is a live test email sent using the new email capability.</p>",
		IsHTML: true,
	}
	fmt.Printf("Sending test email to %s...\n", emailData.To[0])
	error = client.Send(emailData)
	suite.NoError(error, "Failed to send email")
	fmt.Println("Email sent successfully!")
}
