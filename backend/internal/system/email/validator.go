package email

import (
	"net/mail"
	"strings"
)

// IsValidEmail returns true for syntactically valid email addresses.
// This consolidates validation logic so other packages can reuse a single
// source of truth for what the system considers a valid address.
func IsValidEmail(emailAddr string) bool {
	// Reject CR/LF early (do not trim them away first) to prevent header injection.
	if strings.ContainsAny(emailAddr, "\r\n") {
		return false
	}

	emailAddr = strings.TrimSpace(emailAddr)
	if emailAddr == "" {
		return false
	}

	addr, err := mail.ParseAddress(emailAddr)
	if err != nil {
		return false
	}
	parsedEmail := addr.Address

	atIndex := strings.LastIndex(parsedEmail, "@")
	if atIndex == -1 || atIndex == 0 || atIndex == len(parsedEmail)-1 {
		return false
	}

	localPart := parsedEmail[:atIndex]
	domain := parsedEmail[atIndex+1:]

	if len(localPart) == 0 || strings.HasPrefix(localPart, ".") || strings.HasSuffix(localPart, ".") {
		return false
	}

	if !strings.Contains(domain, ".") ||
		strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") ||
		strings.HasPrefix(domain, "-") || strings.HasSuffix(domain, "-") {
		return false
	}

	return true
}
