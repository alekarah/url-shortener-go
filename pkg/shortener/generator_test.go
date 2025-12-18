package shortener

import (
	"testing"
)

// TestGenerate проверяет генерацию случайных кодов
func TestGenerate(t *testing.T) {
	generator := NewGenerator()

	tests := []struct {
		name   string
		length int
	}{
		{"length 6", 6},
		{"length 7", 7},
		{"length 8", 8},
		{"length 10", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := generator.Generate(tt.length)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			if len(code) != tt.length {
				t.Errorf("Generate() length = %d, want %d", len(code), tt.length)
			}

			// Проверяем что код содержит только допустимые символы
			for _, char := range code {
				if !isValidChar(char) {
					t.Errorf("Generate() contains invalid char: %c", char)
				}
			}
		})
	}
}

// TestGenerateUniqueness проверяет уникальность генерируемых кодов
func TestGenerateUniqueness(t *testing.T) {
	generator := NewGenerator()
	length := 7
	iterations := 1000

	codes := make(map[string]bool)

	for i := 0; i < iterations; i++ {
		code, err := generator.Generate(length)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		if codes[code] {
			t.Errorf("Generate() produced duplicate code: %s", code)
		}
		codes[code] = true
	}
}

// TestEncodeID проверяет кодирование числового ID
func TestEncodeID(t *testing.T) {
	generator := NewGenerator()

	tests := []struct {
		name string
		id   int64
	}{
		{"id 1", 1},
		{"id 100", 100},
		{"id 1000", 1000},
		{"id 999999", 999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := generator.EncodeID(tt.id)

			if len(code) == 0 {
				t.Error("EncodeID() returned empty string")
			}

			// Проверяем что код содержит только допустимые символы
			for _, char := range code {
				if !isValidChar(char) {
					t.Errorf("EncodeID() contains invalid char: %c", char)
				}
			}
		})
	}
}

// TestEncodeIDConsistency проверяет что один ID всегда дает один код
func TestEncodeIDConsistency(t *testing.T) {
	generator := NewGenerator()
	id := int64(12345)

	code1 := generator.EncodeID(id)
	code2 := generator.EncodeID(id)

	if code1 != code2 {
		t.Errorf("EncodeID() inconsistent: %s != %s", code1, code2)
	}
}

// TestIsValidShortCode проверяет валидацию коротких кодов
func TestIsValidShortCode(t *testing.T) {
	tests := []struct {
		name  string
		code  string
		valid bool
	}{
		{"valid alphanumeric", "abc123", true},
		{"valid with dash", "abc-123", true},
		{"valid with underscore", "abc_123", true},
		{"valid uppercase", "ABC123", true},
		{"valid mixed case", "AbC123", true},
		{"empty string", "", false},
		{"too long", "a123456789012345678901234567890123456789012345678901", false},
		{"with space", "abc 123", false},
		{"with special chars", "abc@123", false},
		{"with slash", "abc/123", false},
		{"cyrillic", "абв123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidShortCode(tt.code)
			if result != tt.valid {
				t.Errorf("IsValidShortCode(%q) = %v, want %v", tt.code, result, tt.valid)
			}
		})
	}
}

// TestGenerateZeroLength проверяет обработку нулевой длины
func TestGenerateZeroLength(t *testing.T) {
	generator := NewGenerator()

	_, err := generator.Generate(0)
	if err == nil {
		t.Error("Generate(0) should return error")
	}
}

// TestGenerateNegativeLength проверяет обработку отрицательной длины
func TestGenerateNegativeLength(t *testing.T) {
	generator := NewGenerator()

	_, err := generator.Generate(-5)
	if err == nil {
		t.Error("Generate(-5) should return error")
	}
}

// isValidChar проверяет что символ входит в Base62 алфавит
func isValidChar(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= 'a' && char <= 'z')
}
