package orm

import (
	"testing"
)

func TestHasConsecutiveUppercase(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// Valid strict CamelCase - should return false
		{"Id", false},
		{"SomeId", false},
		{"SomeJson", false},
		{"WorkingId", false},
		{"UserName", false},
		{"CreatedAt", false},
		{"HttpStatus", false},
		{"UrlPath", false},
		{"XmlParser", false},
		{"name", false},
		{"Name", false},
		{"a", false},
		{"A", false},
		{"", false},

		// Invalid - consecutive uppercase - should return true
		{"ID", true},
		{"SomeID", true},
		{"SomeJSON", true},
		{"HTTPStatus", true},
		{"URLPath", true},
		{"XMLParser", true},
		{"UserID", true},
		{"ALLCAPS", true},
		{"AB", true},
		{"ABC", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := hasConsecutiveUppercase(tt.input)
			if result != tt.expected {
				t.Errorf("hasConsecutiveUppercase(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToStrictCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Already valid - should remain unchanged
		{"Id", "Id"},
		{"SomeId", "SomeId"},
		{"SomeJson", "SomeJson"},
		{"UserName", "UserName"},
		{"name", "name"},
		{"Name", "Name"},
		{"", ""},

		// Invalid - should be corrected
		{"ID", "Id"},
		{"SomeID", "SomeId"},
		{"SomeJSON", "SomeJson"},
		{"HTTPStatus", "HttpStatus"},
		{"URLPath", "UrlPath"},
		{"XMLParser", "XmlParser"},
		{"UserID", "UserId"},
		{"ALLCAPS", "Allcaps"},
		{"AB", "Ab"},
		{"ABC", "Abc"},
		{"HTTPSProtocol", "HttpsProtocol"},
		{"GetUserID", "GetUserId"},
		{"ParseJSON", "ParseJson"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toStrictCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("toStrictCamelCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateFieldNaming(t *testing.T) {
	tests := []struct {
		fieldName   string
		shouldError bool
		wantContain string // expected substring in error message
	}{
		// Valid names - should not error
		{"Id", false, ""},
		{"SomeId", false, ""},
		{"SomeJson", false, ""},
		{"WorkingId", false, ""},
		{"UserName", false, ""},
		{"CreatedAt", false, ""},
		{"Name", false, ""},
		{"Count", false, ""},

		// Invalid names - should error with correction suggestion
		{"ID", true, "use 'Id' instead"},
		{"SomeID", true, "use 'SomeId' instead"},
		{"SomeJSON", true, "use 'SomeJson' instead"},
		{"HTTPStatus", true, "use 'HttpStatus' instead"},
		{"UserID", true, "use 'UserId' instead"},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			err := validateFieldNaming(tt.fieldName)

			if tt.shouldError {
				if err == nil {
					t.Errorf("validateFieldNaming(%q) = nil, want error", tt.fieldName)
					return
				}
				if tt.wantContain != "" && !contains(err.Error(), tt.wantContain) {
					t.Errorf("validateFieldNaming(%q) error = %q, want to contain %q", tt.fieldName, err.Error(), tt.wantContain)
				}
			} else {
				if err != nil {
					t.Errorf("validateFieldNaming(%q) = %v, want nil", tt.fieldName, err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
