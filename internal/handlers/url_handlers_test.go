package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"url-short/internal/models"
)

// mockURLService мок для тестирования handlers
type mockURLService struct {
	createFunc     func(context.Context, *models.CreateURLRequest) (*models.URLResponse, error)
	getOriginalURL func(context.Context, string) (string, error)
	getByID        func(context.Context, int64) (*models.URLResponse, error)
	getAllURLs     func(context.Context, int, int) ([]*models.URLResponse, error)
	deleteURL      func(context.Context, int64) error
}

func (m *mockURLService) CreateShortURL(ctx context.Context, req *models.CreateURLRequest) (*models.URLResponse, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockURLService) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	if m.getOriginalURL != nil {
		return m.getOriginalURL(ctx, shortCode)
	}
	return "", nil
}

func (m *mockURLService) GetURLByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	return nil, nil
}

func (m *mockURLService) GetURLByID(ctx context.Context, id int64) (*models.URLResponse, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return nil, nil
}

func (m *mockURLService) GetAllURLs(ctx context.Context, limit, offset int) ([]*models.URLResponse, error) {
	if m.getAllURLs != nil {
		return m.getAllURLs(ctx, limit, offset)
	}
	return []*models.URLResponse{}, nil
}

func (m *mockURLService) DeleteURL(ctx context.Context, id int64) error {
	if m.deleteURL != nil {
		return m.deleteURL(ctx, id)
	}
	return nil
}

func (m *mockURLService) IncrementClicks(ctx context.Context, id int64) error {
	return nil
}

// TestCreateShortURL_Success проверяет успешное создание ссылки
func TestCreateShortURL_Success(t *testing.T) {
	mockService := &mockURLService{
		createFunc: func(ctx context.Context, req *models.CreateURLRequest) (*models.URLResponse, error) {
			return &models.URLResponse{
				ID:          1,
				ShortCode:   "abc123",
				ShortURL:    "http://localhost:8080/abc123",
				OriginalURL: req.OriginalURL,
			}, nil
		},
	}

	handler := NewURLHandler(mockService)

	reqBody := `{"original_url":"https://example.com"}`
	req := httptest.NewRequest("POST", "/api/v1/urls", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.CreateShortURL(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusCreated)
	}

	var response models.URLResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ShortCode != "abc123" {
		t.Errorf("ShortCode = %s, want abc123", response.ShortCode)
	}

	if response.OriginalURL != "https://example.com" {
		t.Errorf("OriginalURL = %s, want https://example.com", response.OriginalURL)
	}
}

// TestCreateShortURL_InvalidJSON проверяет обработку невалидного JSON
func TestCreateShortURL_InvalidJSON(t *testing.T) {
	mockService := &mockURLService{}
	handler := NewURLHandler(mockService)

	reqBody := `{invalid json}`
	req := httptest.NewRequest("POST", "/api/v1/urls", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.CreateShortURL(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// TestCreateShortURL_EmptyURL проверяет обработку пустого URL
func TestCreateShortURL_EmptyURL(t *testing.T) {
	mockService := &mockURLService{}
	handler := NewURLHandler(mockService)

	reqBody := `{"original_url":""}`
	req := httptest.NewRequest("POST", "/api/v1/urls", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.CreateShortURL(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// TestGetURL_Success проверяет успешное получение URL по ID
func TestGetURL_Success(t *testing.T) {
	mockService := &mockURLService{
		getByID: func(ctx context.Context, id int64) (*models.URLResponse, error) {
			return &models.URLResponse{
				ID:          id,
				ShortCode:   "test123",
				ShortURL:    "http://localhost:8080/test123",
				OriginalURL: "https://example.com",
				ClicksCount: 10,
			}, nil
		},
	}

	handler := NewURLHandler(mockService)

	req := httptest.NewRequest("GET", "/api/v1/urls/1", nil)

	// Используем chi context для передачи параметра ID
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.GetURL(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	var response models.URLResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ID != 1 {
		t.Errorf("ID = %d, want 1", response.ID)
	}

	if response.ClicksCount != 10 {
		t.Errorf("ClicksCount = %d, want 10", response.ClicksCount)
	}
}

// TestGetURL_InvalidID проверяет обработку невалидного ID
func TestGetURL_InvalidID(t *testing.T) {
	mockService := &mockURLService{}
	handler := NewURLHandler(mockService)

	req := httptest.NewRequest("GET", "/api/v1/urls/invalid", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.GetURL(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// TestGetAllURLs_Success проверяет получение списка URL
func TestGetAllURLs_Success(t *testing.T) {
	mockService := &mockURLService{
		getAllURLs: func(ctx context.Context, limit, offset int) ([]*models.URLResponse, error) {
			return []*models.URLResponse{
				{
					ID:          1,
					ShortCode:   "test1",
					OriginalURL: "https://example1.com",
				},
				{
					ID:          2,
					ShortCode:   "test2",
					OriginalURL: "https://example2.com",
				},
			}, nil
		},
	}

	handler := NewURLHandler(mockService)

	req := httptest.NewRequest("GET", "/api/v1/urls?limit=10&offset=0", nil)
	w := httptest.NewRecorder()
	handler.GetAllURLs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	var response []*models.URLResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Response length = %d, want 2", len(response))
	}

	if response[0].ShortCode != "test1" {
		t.Errorf("First ShortCode = %s, want test1", response[0].ShortCode)
	}
}

// TestGetAllURLs_DefaultParams проверяет дефолтные параметры пагинации
func TestGetAllURLs_DefaultParams(t *testing.T) {
	mockService := &mockURLService{
		getAllURLs: func(ctx context.Context, limit, offset int) ([]*models.URLResponse, error) {
			// Проверяем что вызвалось с дефолтными значениями
			if limit != 50 {
				t.Errorf("limit = %d, want 50", limit)
			}
			if offset != 0 {
				t.Errorf("offset = %d, want 0", offset)
			}
			return []*models.URLResponse{}, nil
		},
	}

	handler := NewURLHandler(mockService)

	req := httptest.NewRequest("GET", "/api/v1/urls", nil)
	w := httptest.NewRecorder()
	handler.GetAllURLs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}
}

// TestDeleteURL_Success проверяет успешное удаление URL
func TestDeleteURL_Success(t *testing.T) {
	deleted := false
	mockService := &mockURLService{
		deleteURL: func(ctx context.Context, id int64) error {
			deleted = true
			return nil
		},
	}

	handler := NewURLHandler(mockService)

	req := httptest.NewRequest("DELETE", "/api/v1/urls/1", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.DeleteURL(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	if !deleted {
		t.Error("DeleteURL should have been called")
	}
}

// TestDeleteURL_InvalidID проверяет удаление с невалидным ID
func TestDeleteURL_InvalidID(t *testing.T) {
	mockService := &mockURLService{}
	handler := NewURLHandler(mockService)

	req := httptest.NewRequest("DELETE", "/api/v1/urls/abc", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.DeleteURL(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
