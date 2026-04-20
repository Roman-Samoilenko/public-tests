CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- -------------------------------------------------------
-- Демографический профиль пользователя
-- -------------------------------------------------------
CREATE TABLE profiles (
    user_id     BIGINT PRIMARY KEY,  -- совпадает с users.id в auth-service
    age         SMALLINT CHECK (age > 0 AND age < 150),
    gender      CHAR(1)  CHECK (gender IN ('M', 'F', 'O')),
    income      INTEGER,             -- среднемесячный в у.е.
    children    SMALLINT CHECK (children >= 0),
    religion    VARCHAR(50),
    education   VARCHAR(50),        -- дополнительное поле
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- -------------------------------------------------------
-- Тесты (опросы)
-- -------------------------------------------------------
CREATE TABLE tests (
    id          BIGSERIAL PRIMARY KEY,
    author_id   BIGINT NOT NULL,
    title       VARCHAR(200) NOT NULL,
    description TEXT,
    -- questions: [{id, text, type, options:[{id,text,score}], min, max, ...}]
    questions   JSONB NOT NULL DEFAULT '[]',
    status      VARCHAR(20) NOT NULL DEFAULT 'published'
                    CHECK (status IN ('published', 'blocked')),
    is_official BOOLEAN NOT NULL DEFAULT FALSE,
    rating      INTEGER NOT NULL DEFAULT 0,
    pass_count  INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- -------------------------------------------------------
-- Ответы пользователей на тесты
-- -------------------------------------------------------
CREATE TABLE test_answers (
    id          BIGSERIAL PRIMARY KEY,
    test_id     BIGINT NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL,
    -- answers: {question_id: value, ...}
    -- value: строка / число / массив для multiple choice
    answers     JSONB NOT NULL,
    score       INTEGER,            -- NULL если тест без баллов
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (test_id, user_id)       -- один пользователь — один ответ на тест
);

-- -------------------------------------------------------
-- Голосования за тесты
-- -------------------------------------------------------
CREATE TABLE test_votes (
    user_id     BIGINT NOT NULL,
    test_id     BIGINT NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
    vote        SMALLINT NOT NULL CHECK (vote IN (1, -1)),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, test_id)
);

-- -------------------------------------------------------
-- Комментарии к тестам
-- -------------------------------------------------------
CREATE TABLE test_comments (
    id          BIGSERIAL PRIMARY KEY,
    test_id     BIGINT NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL,
    content     TEXT NOT NULL CHECK (char_length(content) BETWEEN 1 AND 1000),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- -------------------------------------------------------
-- Лог создания тестов (для модерации)
-- -------------------------------------------------------
CREATE TABLE moderation_log (
    id          BIGSERIAL PRIMARY KEY,
    test_id     BIGINT NOT NULL,
    author_id   BIGINT NOT NULL,
    title       VARCHAR(200) NOT NULL,
    action      VARCHAR(50) NOT NULL DEFAULT 'created',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- -------------------------------------------------------
-- Корреляции (заполняется worker'ом)
-- -------------------------------------------------------
CREATE TABLE correlations (
    id          BIGSERIAL PRIMARY KEY,
    -- source
    source_type VARCHAR(20) NOT NULL CHECK (source_type IN ('demographic', 'test_answer')),
    source_id   TEXT NOT NULL,   -- 'age' | 'gender' | 'test:1:q:2:val:3'
    -- target
    target_type VARCHAR(20) NOT NULL CHECK (target_type IN ('demographic', 'test_answer')),
    target_id   TEXT NOT NULL,
    -- статистика
    coeff       REAL NOT NULL,   -- коэффициент корреляции [-1, 1]
    p_value     REAL NOT NULL,
    sample_size INTEGER NOT NULL,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- -------------------------------------------------------
-- Индексы
-- -------------------------------------------------------
CREATE INDEX idx_tests_author     ON tests(author_id);
CREATE INDEX idx_tests_official   ON tests(is_official) WHERE is_official = TRUE;
CREATE INDEX idx_tests_status     ON tests(status);
CREATE INDEX idx_tests_rating     ON tests(rating DESC);
CREATE INDEX idx_answers_user     ON test_answers(user_id);
CREATE INDEX idx_answers_test     ON test_answers(test_id);
CREATE INDEX idx_comments_test    ON test_comments(test_id);
CREATE INDEX idx_modlog_test      ON moderation_log(test_id);
CREATE INDEX idx_corr_source      ON correlations(source_type, source_id);
CREATE INDEX idx_corr_target      ON correlations(target_type, target_id);

-- -------------------------------------------------------
-- Триггер: пересчёт рейтинга теста при изменении голосов
-- -------------------------------------------------------
CREATE OR REPLACE FUNCTION update_test_rating()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE tests
    SET rating = (
        SELECT COALESCE(SUM(vote), 0)
        FROM test_votes
        WHERE test_id = COALESCE(NEW.test_id, OLD.test_id)
    )
    WHERE id = COALESCE(NEW.test_id, OLD.test_id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_test_rating
AFTER INSERT OR UPDATE OR DELETE ON test_votes
FOR EACH ROW EXECUTE FUNCTION update_test_rating();
