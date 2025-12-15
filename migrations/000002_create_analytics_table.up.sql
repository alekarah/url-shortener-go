CREATE TABLE IF NOT EXISTS analytics (
    id BIGSERIAL PRIMARY KEY,
    url_id BIGINT NOT NULL REFERENCES urls(id) ON DELETE CASCADE,
    clicked_at TIMESTAMP DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    referer TEXT,
    country VARCHAR(2),
    city VARCHAR(100)
);

-- Индекс для быстрого получения статистики по URL
CREATE INDEX idx_analytics_url_id ON analytics(url_id);

-- Индекс для временного анализа кликов
CREATE INDEX idx_analytics_clicked_at ON analytics(clicked_at DESC);

-- Составной индекс для аналитики по URL и времени
CREATE INDEX idx_analytics_url_id_clicked_at ON analytics(url_id, clicked_at DESC);

-- Индекс для географической аналитики
CREATE INDEX idx_analytics_country ON analytics(country) WHERE country IS NOT NULL;
