package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"url-short/internal/models"
)

// mockURLRepository мок для тестирования URLService
type mockURLRepository struct {
	urls      map[string]*models.URL
	urlsByID  map[int64]*models.URL
	nextID    int64
	codeExist func(string) bool
}

func newMockURLRepository() *mockURLRepository {
	return &mockURLRepository{
		urls:     make(map[string]*models.URL),
		urlsByID: make(map[int64]*models.URL),
		nextID:   1,
	}
}

func (m *mockURLRepository) Create(ctx context.Context, url *models.URL) error {
	url.ID = m.nextID
	url.CreatedAt = time.Now()
	url.ClicksCount = 0
	m.nextID++

	m.urls[url.ShortCode] = url
	m.urlsByID[url.ID] = url
	return nil
}

func (m *mockURLRepository) GetByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	url, exists := m.urls[shortCode]
	if !exists {
		return nil, fmt.Errorf("URL с кодом %s не найден", shortCode)
	}
	return url, nil
}

func (m *mockURLRepository) GetByID(ctx context.Context, id int64) (*models.URL, error) {
	url, exists := m.urlsByID[id]
	if !exists {
		return nil, nil
	}
	return url, nil
}

func (m *mockURLRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.URL, error) {
	var urls []*models.URL
	for _, url := range m.urlsByID {
		urls = append(urls, url)
	}
	return urls, nil
}

func (m *mockURLRepository) Update(ctx context.Context, url *models.URL) error {
	m.urls[url.ShortCode] = url
	m.urlsByID[url.ID] = url
	return nil
}

func (m *mockURLRepository) Delete(ctx context.Context, id int64) error {
	url := m.urlsByID[id]
	if url != nil {
		delete(m.urls, url.ShortCode)
		delete(m.urlsByID, id)
	}
	return nil
}

func (m *mockURLRepository) IncrementClicks(ctx context.Context, id int64) error {
	if url, exists := m.urlsByID[id]; exists {
		url.ClicksCount++
	}
	return nil
}

func (m *mockURLRepository) ShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	if m.codeExist != nil {
		return m.codeExist(shortCode), nil
	}
	_, exists := m.urls[shortCode]
	return exists, nil
}

// mockGenerator мок генератор для предсказуемых тестов
type mockGenerator struct {
	code string
}

func (m *mockGenerator) Generate(length int) (string, error) {
	if m.code != "" {
		return m.code, nil
	}
	return "ABC123", nil
}

func (m *mockGenerator) EncodeID(id int64) string {
	return "ENC123"
}

// TestCreateShortURL_Success проверяет успешное создание короткой ссылки
func TestCreateShortURL_Success(t *testing.T) {
	repo := newMockURLRepository()
	gen := &mockGenerator{}
	service := NewURLService(repo, gen, nil, "http://localhost:8080", 3600)

	req := &models.CreateURLRequest{
		OriginalURL: "https://example.com",
	}

	result, err := service.CreateShortURL(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateShortURL() error = %v", err)
	}

	if result.ID == 0 {
		t.Error("CreateShortURL() ID should not be 0")
	}

	if result.ShortCode != "ABC123" {
		t.Errorf("CreateShortURL() ShortCode = %s, want ABC123", result.ShortCode)
	}

	if result.OriginalURL != "https://example.com" {
		t.Errorf("CreateShortURL() OriginalURL = %s, want https://example.com", result.OriginalURL)
	}

	if result.ShortURL != "http://localhost:8080/ABC123" {
		t.Errorf("CreateShortURL() ShortURL = %s, want http://localhost:8080/ABC123", result.ShortURL)
	}
}

// TestCreateShortURL_CustomCode проверяет создание с кастомным кодом
func TestCreateShortURL_CustomCode(t *testing.T) {
	repo := newMockURLRepository()
	gen := &mockGenerator{}
	service := NewURLService(repo, gen, nil, "http://localhost:8080", 3600)

	req := &models.CreateURLRequest{
		OriginalURL: "https://example.com",
		CustomCode:  "mycode",
	}

	result, err := service.CreateShortURL(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateShortURL() error = %v", err)
	}

	if result.ShortCode != "mycode" {
		t.Errorf("CreateShortURL() ShortCode = %s, want mycode", result.ShortCode)
	}
}

// TestCreateShortURL_DuplicateCustomCode проверяет обработку дубликата кастомного кода
func TestCreateShortURL_DuplicateCustomCode(t *testing.T) {
	repo := newMockURLRepository()
	gen := &mockGenerator{}
	service := NewURLService(repo, gen, nil, "http://localhost:8080", 3600)

	// Создаем первую ссылку
	req1 := &models.CreateURLRequest{
		OriginalURL: "https://example.com",
		CustomCode:  "mycode",
	}
	_, err := service.CreateShortURL(context.Background(), req1)
	if err != nil {
		t.Fatalf("First CreateShortURL() error = %v", err)
	}

	// Пытаемся создать вторую с тем же кодом
	req2 := &models.CreateURLRequest{
		OriginalURL: "https://another.com",
		CustomCode:  "mycode",
	}
	_, err = service.CreateShortURL(context.Background(), req2)
	if err == nil {
		t.Error("CreateShortURL() should return error for duplicate code")
	}
}

// TestGetOriginalURL_Success проверяет получение оригинального URL
func TestGetOriginalURL_Success(t *testing.T) {
	repo := newMockURLRepository()
	gen := &mockGenerator{}
	service := NewURLService(repo, gen, nil, "http://localhost:8080", 3600)

	// Создаем ссылку
	req := &models.CreateURLRequest{
		OriginalURL: "https://example.com",
		CustomCode:  "test123",
	}
	_, err := service.CreateShortURL(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateShortURL() error = %v", err)
	}

	// Получаем оригинальный URL
	originalURL, err := service.GetOriginalURL(context.Background(), "test123")
	if err != nil {
		t.Fatalf("GetOriginalURL() error = %v", err)
	}

	if originalURL != "https://example.com" {
		t.Errorf("GetOriginalURL() = %s, want https://example.com", originalURL)
	}
}

// TestGetOriginalURL_NotFound проверяет обработку несуществующего кода
func TestGetOriginalURL_NotFound(t *testing.T) {
	repo := newMockURLRepository()
	gen := &mockGenerator{}
	// Важно: передаем nil для redis, чтобы избежать panic
	service := &urlService{
		urlRepo:   repo,
		generator: gen,
		redis:     nil,
		baseURL:   "http://localhost:8080",
		cacheTTL:  3600,
	}

	_, err := service.GetOriginalURL(context.Background(), "nonexistent")
	if err == nil {
		t.Error("GetOriginalURL() should return error for non-existent code")
	}
}

// TestGetAllURLs_Pagination проверяет пагинацию
func TestGetAllURLs_Pagination(t *testing.T) {
	repo := newMockURLRepository()
	gen := &mockGenerator{code: "CODE"}
	service := NewURLService(repo, gen, nil, "http://localhost:8080", 3600)

	// Создаем несколько ссылок
	for i := 1; i <= 5; i++ {
		gen.code = string(rune('A' + i))
		req := &models.CreateURLRequest{
			OriginalURL: "https://example.com",
		}
		_, err := service.CreateShortURL(context.Background(), req)
		if err != nil {
			t.Fatalf("CreateShortURL() error = %v", err)
		}
	}

	// Получаем с лимитом
	urls, err := service.GetAllURLs(context.Background(), 3, 0)
	if err != nil {
		t.Fatalf("GetAllURLs() error = %v", err)
	}

	if len(urls) == 0 {
		t.Error("GetAllURLs() should return URLs")
	}
}

// TestGetAllURLs_DefaultLimit проверяет дефолтный лимит
func TestGetAllURLs_DefaultLimit(t *testing.T) {
	repo := newMockURLRepository()
	gen := &mockGenerator{}
	service := NewURLService(repo, gen, nil, "http://localhost:8080", 3600)

	// Тестируем с некорректными параметрами
	_, err := service.GetAllURLs(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("GetAllURLs() error = %v", err)
	}

	// Limit 0 должен стать 50 (по умолчанию)
	_, err = service.GetAllURLs(context.Background(), -1, -1)
	if err != nil {
		t.Fatalf("GetAllURLs() error = %v", err)
	}
}

// TestIncrementClicks проверяет инкремент счетчика
func TestIncrementClicks(t *testing.T) {
	repo := newMockURLRepository()
	gen := &mockGenerator{}
	service := NewURLService(repo, gen, nil, "http://localhost:8080", 3600)

	// Создаем ссылку
	req := &models.CreateURLRequest{
		OriginalURL: "https://example.com",
		CustomCode:  "click123",
	}
	result, err := service.CreateShortURL(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateShortURL() error = %v", err)
	}

	// Инкрементим клики
	err = service.IncrementClicks(context.Background(), result.ID)
	if err != nil {
		t.Fatalf("IncrementClicks() error = %v", err)
	}

	// Проверяем что счетчик увеличился
	url, _ := repo.GetByID(context.Background(), result.ID)
	if url.ClicksCount != 1 {
		t.Errorf("ClicksCount = %d, want 1", url.ClicksCount)
	}
}
