package email

import "testing"

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		in  string
		out bool
	}{
		{"test@example.com", true},
		{"user.name@domain.org", true},
		{" invalid@example.com ", true},
		{"test@example.com\r\n", false},
		{"not-an-email", false},
		{"@example.com", false},
		{"", false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			if IsValidEmail(tc.in) != tc.out {
				t.Fatalf("IsValidEmail(%q)= %v, want %v", tc.in, IsValidEmail(tc.in), tc.out)
			}
		})
	}
}
