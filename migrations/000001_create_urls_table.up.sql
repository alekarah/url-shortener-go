CREATE TABLE IF NOT EXISTS urls (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(10) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    user_id BIGINT,
    clicks_count BIGINT DEFAULT 0,
    last_clicked_at TIMESTAMP
);

-- Индекс для быстрого поиска по короткому коду (главный use case)
CREATE INDEX idx_urls_short_code ON urls(short_code);

-- Индекс для поиска по дате создания (аналитика)
CREATE INDEX idx_urls_created_at ON urls(created_at DESC);

-- Индекс для очистки истекших ссылок
CREATE INDEX idx_urls_expires_at ON urls(expires_at) WHERE expires_at IS NOT NULL;

-- Индекс для фильтрации по пользователю (если добавим аутентификацию)
CREATE INDEX idx_urls_user_id ON urls(user_id) WHERE user_id IS NOT NULL;
