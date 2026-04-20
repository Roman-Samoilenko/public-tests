-- -------------------------------------------------------
-- Теги для тестов
-- -------------------------------------------------------
ALTER TABLE tests ADD COLUMN tags TEXT[] NOT NULL DEFAULT '{}';

-- -------------------------------------------------------
-- Конфигурация результата (v1 — хранение, v2 — вычисление)
-- -------------------------------------------------------
ALTER TABLE tests ADD COLUMN result_config JSONB;

-- -------------------------------------------------------
-- Обновление ответов: updated_at и result
-- Уникальный constraint (test_id, user_id) ОСТАЁТСЯ — upsert через ON CONFLICT DO UPDATE
-- -------------------------------------------------------
ALTER TABLE test_answers ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE test_answers ADD COLUMN result JSONB;

-- -------------------------------------------------------
-- Никнейм в комментариях (фиксируется в момент создания из JWT)
-- -------------------------------------------------------
ALTER TABLE test_comments ADD COLUMN nickname VARCHAR(100) NOT NULL DEFAULT '';

-- -------------------------------------------------------
-- Полнотекстовый поиск по тестам (русский язык)
-- Генерируется автоматически из title + description + tags
-- -------------------------------------------------------
ALTER TABLE tests ADD COLUMN search_vector TSVECTOR;

-- 2. Создаём функцию-триггер
CREATE OR REPLACE FUNCTION tests_search_vector_update() 
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector := to_tsvector('russian', 
        COALESCE(NEW.title, '') || ' ' ||
        COALESCE(NEW.description, '') || ' ' ||
        COALESCE(array_to_string(NEW.tags, ' '), '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 3. Привязываем триггер к таблице
CREATE TRIGGER trigger_tests_search_vector_update
    BEFORE INSERT OR UPDATE ON tests
    FOR EACH ROW
    EXECUTE FUNCTION tests_search_vector_update();
-- -------------------------------------------------------
-- Индексы
-- -------------------------------------------------------
CREATE INDEX idx_tests_search ON tests USING GIN(search_vector);
CREATE INDEX idx_tests_tags   ON tests USING GIN(tags);
