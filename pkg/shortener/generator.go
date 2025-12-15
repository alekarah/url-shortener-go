package shortener

import (
	"crypto/rand"
	"math/big"
)

// Base62 алфавит для генерации коротких кодов
const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Generator интерфейс для генерации коротких кодов
type Generator interface {
	Generate(length int) (string, error)
	EncodeID(id int64) string
}

// generator имплементация Generator
type generator struct{}

// NewGenerator создает новый генератор коротких кодов
func NewGenerator() Generator {
	return &generator{}
}

// Generate генерирует случайный короткий код заданной длины
func (g *generator) Generate(length int) (string, error) {
	if length <= 0 {
		length = 7 // значение по умолчанию
	}

	result := make([]byte, length)
	maxIndex := big.NewInt(int64(len(base62Chars)))

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, maxIndex)
		if err != nil {
			return "", err
		}
		result[i] = base62Chars[num.Int64()]
	}

	return string(result), nil
}

// EncodeID конвертирует числовой ID в Base62 строку
// Используется как альтернативный метод для детерминированной генерации
func (g *generator) EncodeID(id int64) string {
	if id == 0 {
		return string(base62Chars[0])
	}

	base := int64(len(base62Chars))
	result := make([]byte, 0, 11) // максимум 11 символов для int64

	for id > 0 {
		remainder := id % base
		result = append([]byte{base62Chars[remainder]}, result...)
		id = id / base
	}

	return string(result)
}

// DecodeToID декодирует Base62 строку обратно в ID
func DecodeToID(encoded string) int64 {
	base := int64(len(base62Chars))
	var result int64

	for _, char := range encoded {
		result = result * base
		// Находим индекс символа в base62Chars
		for i, c := range base62Chars {
			if c == char {
				result += int64(i)
				break
			}
		}
	}

	return result
}

// IsValidShortCode проверяет, что короткий код содержит только допустимые символы
func IsValidShortCode(code string) bool {
	if len(code) == 0 {
		return false
	}

	for _, char := range code {
		found := false
		for _, validChar := range base62Chars {
			if char == validChar {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
