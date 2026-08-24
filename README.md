# PublicTests — Платформа Публичных Тестов

Высокопроизводительная веб-платформа для создания, прохождения и аналитики публичных тестов и опросов. Система реализована на микросервисной архитектуре с разделением на независимые сервисы аутентификации и бизнес-логики, использует событийную модель авторизации на основе JWT с механизмом ротации токенов, полнотекстовый поиск на уровне базы данных и автоматический импорт опросов из Google Forms.

Платформа ориентирована на сбор статистических данных и исследование корреляций между демографическими характеристиками пользователей и их ответами на тесты, что делает её инструментом как для развлечения, так и для академических и маркетинговых исследований.

## Обзор архитектуры

Платформа реализует паттерн **Multi-Tier Microservices** с централизованной маршрутизацией через Nginx Reverse Proxy, независимым Auth Service с собственной базой данных и основным Backend Service с расширенной бизнес-логикой. Взаимодействие между слоями строится на JWT-токенах с коротким временем жизни (access token 15 минут) и долгоживущих refresh-токенах (7 дней), хранящихся в Redis.

### Архитектура высокого уровня

```mermaid
graph TB
    Client[Browser / API Client]

    subgraph "Edge Layer"
        Nginx[Nginx Reverse Proxy<br/>:8000 dev / :80 prod]
    end

    subgraph "Application Layer"
        Auth[Auth Service<br/>:8081<br/>Go + net/http]
        Backend[Backend Service<br/>:8080<br/>Go + chi/v5]
        Frontend[Frontend Dev Server<br/>:5173<br/>Vue 3 + Vite]
    end

    subgraph "Data Layer"
        PG_Auth[(PostgreSQL Auth<br/>auth_db<br/>users, refresh_tokens)]
        PG_Main[(PostgreSQL Main<br/>main_db<br/>tests, answers, votes)]
        Redis[(Redis 7<br/>refresh token store<br/>session cache)]
    end

    Client -->|HTTP| Nginx
    Nginx -->|/api/auth/*| Auth
    Nginx -->|/api/*| Backend
    Nginx -->|/* dev| Frontend

    Auth -->|bcrypt + JWT| PG_Auth
    Auth -->|refresh tokens| Redis
    Backend -->|ORM-free SQL| PG_Main
    Backend -->|JWT validation| Auth

    style Nginx fill:#4a90e2,stroke:#2171c9,color:#fff
    style Auth fill:#e27a4a,stroke:#c96021,color:#fff
    style Backend fill:#4ae290,stroke:#21c962,color:#fff
    style Frontend fill:#9b59b6,stroke:#7d3c98,color:#fff
    style PG_Auth fill:#336791,stroke:#1a3f5c,color:#fff
    style PG_Main fill:#336791,stroke:#1a3f5c,color:#fff
    style Redis fill:#dc382d,stroke:#a82820,color:#fff
```

### Принципы проектирования

**Разделение ответственности (Separation of Concerns)**

- **Auth Service:** Единственная точка истины об идентификации пользователей. Управляет регистрацией через OTP-коды (email), выдачей пар access/refresh токенов, ротацией сессий. Использует bcrypt с индивидуальным pepper для хранения паролей.
- **Backend Service:** Вся бизнес-логика платформы — управление тестами, сбор ответов, система голосований, комментарии, профили пользователей, импорт из Google Forms, административные функции. Валидирует JWT самостоятельно, не обращаясь к Auth Service на каждый запрос.
- **Frontend SPA:** Vue 3 приложение с клиентским роутингом, автоматическим обновлением токенов и оптимистичными обновлениями UI.

**Безопасность (Defense in Depth)**

- Разделённые базы данных для Auth и Main сервисов — компрометация одной схемы не даёт доступа к другой.
- Access token TTL 15 минут минимизирует окно атаки при перехвате.
- Refresh tokens хранятся в Redis с возможностью точечной инвалидации без перезапуска сервисов.
- CONTACT_PEPPER для хэшей паролей — дополнительный уровень защиты против rainbow table атак при утечке БД.
- Middleware recovery на уровне HTTP-сервера предотвращает утечку стек-трейсов в production.

**Масштабируемость**

- Stateless Backend Service — горизонтальное масштабирование без необходимости sticky sessions.
- GIN-индексы на PostgreSQL для полнотекстового поиска и массивов тегов обеспечивают стабильную латентность при росте данных.
- Redis как централизованное хранилище сессий позволяет запускать несколько инстанций Auth Service.

## Технологический стек

### Backend Services

| Компонент | Технология | Версия | Роль |
|-----------|-----------|--------|------|
| **Auth Service** | Go + net/http | 1.22+ | Идентификация, OTP, JWT, refresh tokens |
| **Backend Service** | Go + chi/v5 | 1.22+ | Бизнес-логика, REST API, Google Forms import |
| **Frontend** | Vue 3 + Vite | Vue 3.4 / Vite 5 | SPA, клиентский роутинг, UI |
| **Reverse Proxy** | Nginx | Alpine | Маршрутизация, dev/prod конфигурации |

### Data Layer

| СУБД | Роль | Особенности |
|------|------|-------------|
| **PostgreSQL 15** | Auth storage | users, refresh_tokens, UNIQUE(login), bcrypt hashes |
| **PostgreSQL 15** | Main storage | tests, answers, votes, comments, profiles, correlations |
| **Redis 7** | Session store | Refresh token blacklist и хранилище, TTL-based expiry |

### Ключевые библиотеки

| Библиотека | Применение |
|-----------|-----------|
| `github.com/go-chi/chi/v5` | HTTP routing, middleware chain |
| `github.com/golang-jwt/jwt` | JWT generation и validation |
| `golang.org/x/crypto/bcrypt` | Password hashing |
| `github.com/lib/pq` | PostgreSQL driver |
| `github.com/redis/go-redis` | Redis client |
| `vue-router` | Client-side SPA routing |

## Функциональность платформы

### Система тестов

Ядро платформы — гибкая система создания и прохождения тестов с поддержкой шести типов вопросов и расширяемой конфигурацией результатов.

**Поддерживаемые типы вопросов:**

- `text` — свободный ввод текста (короткий или длинный)
- `single_choice` — выбор одного варианта (radio/dropdown)
- `multiple_choice` — множественный выбор (checkbox)
- `scale` — линейная шкала с подписями минимума и максимума
- `vector_scale` — матрица оценок (сетка строк × столбцов)
- `date`, `time` — специальные типы для временных данных

Структура вопросов хранится как `JSONB` в PostgreSQL, что позволяет добавлять новые типы без миграций схемы. Каждый вопрос содержит поле `required`, массив `options` для выборных типов, поля `min_value`/`max_value` для шкал и `rows`/`cols` для матриц.

**Конфигурация результатов (result_config):**

Тесты поддерживают опциональную конфигурацию интерпретации результатов. В v1 конфигурация хранится без применения — результат фиксируется как-есть. В v2 (roadmap) конфигурация позволит автоматически вычислять баллы и подбирать текстовые интерпретации на основе диапазонов.

### Импорт из Google Forms

Уникальная возможность платформы — автоматический парсинг публичных Google Forms без использования официального API.

```mermaid
sequenceDiagram
    participant User
    participant Frontend
    participant Backend
    participant GoogleForms

    User->>Frontend: Вставляет ссылку на Google Form
    Frontend->>Backend: POST /api/import/google-forms {url}
    Backend->>Backend: Нормализация URL
    Note over Backend: forms.gle → resolve redirect<br/>/forms/d/FORM_ID → /viewform<br/>/forms/d/e/FORM_ID → /viewform
    Backend->>GoogleForms: GET /forms/d/.../viewform<br/>User-Agent: Chrome/120
    GoogleForms-->>Backend: HTML с FB_PUBLIC_LOAD_DATA_
    Backend->>Backend: Regexp extract JSON
    Backend->>Backend: Parse questions by type
    Note over Backend: 0=text, 2=radio, 3=checkbox<br/>4=dropdown, 5=scale, 7=grid
    Backend-->>Frontend: ImportedTest{title, questions[]}
    Frontend->>User: Превью теста для редактирования
    User->>Frontend: Подтверждает/редактирует
    Frontend->>Backend: POST /api/tests
```

Импортер поддерживает все форматы ссылок Google Forms: короткие `forms.gle`, стандартные `/forms/d/FORM_ID` и `/forms/d/e/FORM_ID`. Автоматически преобразует типы вопросов Google Forms во внутренние типы платформы, извлекает варианты ответов, метки шкал и конфигурации матриц.

### Система аутентификации

```mermaid
sequenceDiagram
    participant C as Client
    participant N as Nginx
    participant A as Auth Service
    participant PG as PostgreSQL (auth_db)
    participant R as Redis

    C->>N: POST /api/auth/send-code {email}
    N->>A: Proxy
    A->>PG: Upsert user record
    A->>A: Generate OTP code
    A-->>C: 200 OK (code sent to email)

    C->>N: POST /api/auth/verify {email, code}
    N->>A: Proxy
    A->>PG: Verify OTP, get user_id + nickname
    A->>A: Sign access JWT (15m, user_id, nickname, is_admin)
    A->>A: Sign refresh token (7d)
    A->>R: SETEX refresh_token TTL=7d
    A-->>C: {access_token, user} + HttpOnly cookie (refresh)

    Note over C,R: Обычные запросы
    C->>N: GET /api/tests Authorization: Bearer <access_token>
    N->>N: Pass through
    Note over N: Backend валидирует JWT самостоятельно

    Note over C,R: Истёк access token
    C->>N: POST /api/auth/refresh (cookie refresh_token)
    N->>A: Proxy
    A->>R: GET refresh_token (проверка не revoked)
    A->>A: Sign new access JWT
    A->>R: Rotate refresh token
    A-->>C: {access_token}
```

Frontend реализует прозрачный механизм retry при получении 401: при первой ошибке запускается `tryRefresh()` с защитой от конкурентных вызовов (singleton Promise), после успешного обновления токена исходный запрос повторяется автоматически.

### Аналитика и корреляции

Платформа включает инфраструктуру для статистического анализа данных пользователей.

**Таблица correlations** хранит вычисленные коэффициенты корреляции между:

- Демографическими характеристиками (`demographic`): возраст, пол, доход, количество детей, религия, образование
- Ответами на тесты (`test_answer`): конкретные значения ответов на конкретные вопросы

Каждая запись содержит коэффициент корреляции `[-1, 1]`, p-value для оценки статистической значимости и размер выборки. Корреляции вычисляются фоновым worker-процессом по расписанию.

**Демографический профиль** пользователя включает возраст, пол, доход, количество детей, религию и уровень образования — эти данные используются для сегментации аудитории при построении аналитических отчётов.

### Рейтинговая система и социальные функции

```mermaid
graph LR
    subgraph "Рейтинг теста"
        Vote[test_votes<br/>user_id + test_id<br/>vote: +1 / -1]
        Trigger[PostgreSQL Trigger<br/>trg_test_rating]
        Rating[tests.rating<br/>SUM of votes]
    end

    subgraph "Социальные функции"
        Comments[test_comments<br/>content 1-1000 chars<br/>nickname из JWT]
        PassCount[tests.pass_count<br/>инкремент при ответе]
    end

    Vote -->|AFTER INSERT/UPDATE/DELETE| Trigger
    Trigger -->|UPDATE tests SET rating| Rating
```

Рейтинг тестов обновляется через PostgreSQL-триггер `trg_test_rating`, что гарантирует консистентность данных без дополнительной логики на уровне приложения. Один пользователь может поставить только один голос на тест (Primary Key по паре user_id + test_id), но может изменить свой голос.

Никнейм в комментариях фиксируется в момент создания из JWT-claims и не изменяется при смене ника пользователем — это обеспечивает иммутабельность истории обсуждений.

## Компоненты системы

### Auth Service

Компактный микросервис на стандартной библиотеке Go с четкой слоистой архитектурой и минимальными внешними зависимостями.

**Архитектура слоёв:**

```mermaid
graph TD
    subgraph "Auth Service"
        HTTP[HTTP Handlers<br/>handlers/]
        MW[Middleware<br/>Logger · Recovery · Timeout]
        SVC[Service Layer<br/>token generation · OTP]
        STORE[Storage Layer<br/>storage/postgres.go]
    end

    Client --> HTTP
    HTTP --> MW
    MW --> SVC
    SVC --> STORE

    STORE --> PG[(PostgreSQL<br/>auth_db)]
    SVC --> RD[(Redis<br/>refresh tokens)]

    MW -.->|Structured JSON logs| Logs[log/slog]
```

**Endpoints:**

- `POST /api/auth/send-code` — отправка OTP-кода на email, создание пользователя при первом обращении
- `POST /api/auth/verify` — верификация OTP, выдача access + refresh tokens
- `POST /api/auth/refresh` — обновление пары токенов по refresh cookie
- `POST /api/auth/logout` — инвалидация refresh token в Redis
- `GET /health` — health check для orchestration

**Технические детали:**

- Access Token: HMAC-SHA256, claims: `user_id`, `nickname`, `is_admin`, `exp` (15 минут)
- Refresh Token: криптографически случайный UUID v4, TTL 168 часов в Redis
- Pepper для паролей: конкатенация с `CONTACT_PEPPER` перед bcrypt предотвращает атаки при утечке хэшей
- Context timeouts на всех DB операциях (5 секунд)
- `sync.Once` для инициализации глобального логгера

### Backend Service

Основной сервис бизнес-логики на Go с chi-роутером, чистой архитектурой и расширенными возможностями работы с данными.

**Структура обработчиков:**

```mermaid
graph TD
    subgraph "Backend Service Handlers"
        Tests[TestHandler<br/>tests.go]
        Profiles[ProfileHandler<br/>profile.go]
        Import[ImportHandler<br/>import.go]
        Admin[AdminHandler<br/>admin.go]
    end

    subgraph "Middleware"
        AuthMW[JWT Auth Middleware<br/>ClaimsFromContext]
    end

    subgraph "Repository Layer"
        TestRepo[TestRepository<br/>postgres/test.go]
        AnswerRepo[AnswerRepository<br/>postgres/answer.go]
        ProfileRepo[ProfileRepository<br/>postgres/profile.go]
    end

    subgraph "Service Layer"
        GFImporter[GoogleFormsImporter<br/>service/importer/]
    end

    Router[chi.Router<br/>router.go] --> AuthMW
    AuthMW --> Tests & Profiles & Import & Admin
    Tests --> TestRepo & AnswerRepo
    Profiles --> ProfileRepo
    Import --> GFImporter
```

**REST API Endpoints:**

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| `GET` | `/api/tests` | Список тестов с фильтрацией | Optional |
| `POST` | `/api/tests` | Создание теста | Required |
| `GET` | `/api/tests/:id` | Получение теста по ID | Optional |
| `POST` | `/api/tests/:id/answers` | Отправка ответов | Required |
| `GET` | `/api/tests/:id/my-answer` | Мой ответ на тест | Required |
| `POST` | `/api/tests/:id/vote` | Голосование за тест (+1/-1) | Required |
| `GET` | `/api/tests/:id/comments` | Комментарии к тесту | Optional |
| `POST` | `/api/tests/:id/comments` | Добавить комментарий | Required |
| `DELETE` | `/api/tests/:id/comments/:commentId` | Удалить комментарий | Required (owner) |
| `GET` | `/api/profile` | Профиль текущего пользователя | Required |
| `PUT` | `/api/profile` | Обновление профиля | Required |
| `GET` | `/api/profile/answers` | История ответов пользователя | Required |
| `POST` | `/api/import/google-forms` | Импорт из Google Forms | Required |
| `PUT` | `/api/admin/tests/:id/official` | Пометить тест официальным | Admin |
| `PUT` | `/api/admin/tests/:id/status` | Изменить статус теста | Admin |

**Фильтрация и сортировка тестов:**

Endpoint `GET /api/tests` поддерживает богатый набор параметров:

- `sort=rating|created_at|pass_count` — сортировка
- `filter=all|official|my` — предустановленные фильтры
- `search=...` — полнотекстовый поиск через `TSVECTOR`
- `tags=tag1,tag2` — фильтрация по тегам (GIN-индекс)
- `author_id=123` — тесты конкретного автора
- `limit=12&offset=0` — пагинация (max 100 на страницу)

### Frontend SPA

Vue 3 Single Page Application с Composition API, клиентским роутингом и централизованным управлением состоянием аутентификации.

**Архитектура клиента:**

```mermaid
graph TD
    subgraph "Vue SPA"
        Router[Vue Router<br/>createWebHistory]
        AuthStore[Auth Store<br/>store/auth.js<br/>reactive state]
        APIClient[API Client<br/>api/index.js]
    end

    subgraph "Pages"
        Feed[FeedPage<br/>лента тестов]
        TestPage[TestPage<br/>прохождение]
        Create[CreateTestPage<br/>конструктор]
        Profile[ProfilePage<br/>профиль]
        Auth[AuthPage<br/>вход/регистрация]
        Admin[AdminPage<br/>модерация]
    end

    subgraph "Components"
        TestCard[TestCard<br/>карточка теста]
        QuestionRenderer[QuestionRenderer<br/>рендер вопросов]
    end

    Router --> Feed & TestPage & Create & Profile & Auth & Admin
    Feed --> TestCard
    TestPage --> QuestionRenderer
    APIClient --> AuthStore
```

**Управление токенами на клиенте:**

```javascript
// Прозрачный retry при 401 — пользователь не замечает обновления токена
async function request(base, path, options, retry = true) {
    // ... send request ...
    if (res.status === 401 && retry) {
        const refreshed = await tryRefresh() // singleton Promise — не дублируем refresh
        if (refreshed) return request(base, path, options, false)
        await fullLogout()
    }
}
```

Singleton-паттерн для `tryRefresh()` предотвращает race condition при параллельных запросах с истёкшим токеном — только один refresh-запрос уходит на сервер, остальные ждут его результата.

**JWT-парсинг на клиенте:**

Состояние пользователя восстанавливается из JWT payload при загрузке страницы без дополнительного запроса к серверу. Парсинг выполняется без верификации подписи (для UI), поэтому выполняется проверка `exp * 1000 > Date.now()` для корректной обработки устаревших токенов.

### База данных — Схема

**auth_db — схема аутентификации:**

```sql
CREATE TABLE users (
    id          BIGSERIAL PRIMARY KEY,
    login       VARCHAR(255) UNIQUE NOT NULL,  -- email
    pass_hash   VARCHAR(255) NOT NULL,         -- bcrypt + pepper
    nickname    VARCHAR(100) NOT NULL,
    is_admin    BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE refresh_tokens (
    token       UUID PRIMARY KEY,
    user_id     BIGINT REFERENCES users(id),
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);
```

**main_db — схема платформы:**

```mermaid
erDiagram
    PROFILES {
        bigint user_id PK "Ссылка на users.id в auth_db"
        smallint age "Числовой признак"
        char(1) gender "Категориальный признак"
        integer income "Числовой признак"
        smallint children "Числовой признак"
        varchar(50) religion "Категориальный признак"
        varchar(50) education "Категориальный признак"
        timestamptz updated_at
    }

    TESTS {
        bigserial id PK
        bigint author_id FK "Ссылка на users.id в auth_db"
        varchar(200) title
        text description
        jsonb questions "Структура вопросов"
        varchar(20) status "published/blocked"
        boolean is_official
        integer rating "Вычисляется триггером"
        integer pass_count
        text[] tags "GIN-индекс"
        jsonb result_config
        tsvector search_vector "GIN-индекс, автообновление"
        timestamptz created_at
        timestamptz updated_at
    }

    TEST_ANSWERS {
        bigserial id PK
        bigint test_id FK "Ссылка на tests(id)"
        bigint user_id "Ссылка на users.id в auth_db"
        jsonb answers "Ответы на вопросы"
        integer score "Вычисляемый балл"
        jsonb result "Результат (интерпретация)"
        timestamptz created_at
        timestamptz updated_at
    }

    TEST_VOTES {
        bigint user_id PK "Ссылка на users.id в auth_db"
        bigint test_id PK "Ссылка на tests(id)"
        smallint vote "1 или -1"
        timestamptz created_at
    }

    TEST_COMMENTS {
        bigserial id PK
        bigint test_id FK "Ссылка на tests(id)"
        bigint user_id "Ссылка на users.id в auth_db"
        varchar(100) nickname "Фиксируется из JWT"
        text content "1-1000 символов"
        timestamptz created_at
    }

    MODERATION_LOG {
        bigserial id PK
        bigint test_id "Ссылка на tests(id)"
        bigint author_id "Ссылка на users.id в auth_db"
        varchar(200) title
        varchar(50) action "created/blocked/unblocked"
        timestamptz created_at
    }

    CORRELATIONS {
        bigserial id PK
        varchar(20) source_type "demographic/test_answer"
        text source_id "Идентификатор признака"
        varchar(20) target_type "demographic/test_answer"
        text target_id "Идентификатор признака"
        real coeff "Коэффициент корреляции"
        real p_value "Статистическая значимость"
        integer sample_size "Размер выборки"
        timestamptz computed_at
    }

    %% Связи
    TESTS ||--o{ TEST_ANSWERS : "имеет ответы"
    TESTS ||--o{ TEST_VOTES : "имеет голоса"
    TESTS ||--o{ TEST_COMMENTS : "имеет комментарии"
    PROFILES ||--o{ CORRELATIONS : "используется для корреляций"
    TEST_ANSWERS ||--o{ CORRELATIONS : "используется для корреляций"
```

### Полнотекстовый поиск

Платформа использует нативный PostgreSQL full-text search с автоматическим обновлением через триггер:

```sql
-- Автоматическая генерация search_vector при каждом INSERT/UPDATE
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

-- GIN-индекс для поиска за O(log N)
CREATE INDEX idx_tests_search ON tests USING GIN(search_vector);
CREATE INDEX idx_tests_tags   ON tests USING GIN(tags);
```

Русскоязычная конфигурация `to_tsvector('russian', ...)` обеспечивает корректную морфологическую нормализацию: запрос "тест" найдёт "тесты", "тестирования", "тестовый".

### Nginx конфигурация

Два отдельных конфига обеспечивают одинаковое поведение в dev и prod:

```mermaid
graph LR
    Client[Browser :8000]

    subgraph "nginx-dev (Docker)"
        Nginx[Nginx<br/>:8000]
    end

    subgraph "Services"
        FE[Frontend Dev Server<br/>Vite HMR :5173]
        BE[Backend :8080]
        AU[Auth :8081]
    end

    Client --> Nginx
    Nginx -->|/api/auth/*| AU
    Nginx -->|/api/*| BE
    Nginx -->|/* + WS| FE

    style Nginx fill:#4a90e2,color:#fff
```

В dev-конфиге включена поддержка WebSocket-соединений для Vite HMR (Hot Module Replacement): `proxy_set_header Upgrade $http_upgrade; proxy_set_header Connection "upgrade"`.

## Поток данных

### Создание теста и публикация

```mermaid
sequenceDiagram
    participant U as User
    participant FE as Frontend
    participant BE as Backend
    participant PG as PostgreSQL

    U->>FE: Заполняет форму создания теста
    Note over FE: Поддержка 6 типов вопросов<br/>drag-and-drop порядок<br/>опциональный result_config

    FE->>BE: POST /api/tests<br/>{title, description, questions[], tags[], result_config}
    BE->>BE: JWT middleware — extract userID
    BE->>BE: Validate: title != "", questions != []

    BE->>PG: INSERT INTO tests<br/>(author_id, title, questions, tags, status='published')
    Note over PG: Trigger: update search_vector<br/>to_tsvector('russian', title+desc+tags)
    PG-->>BE: test_id

    BE->>PG: INSERT INTO moderation_log<br/>(test_id, author_id, action='created')
    PG-->>BE: OK

    BE-->>FE: 201 Created {id, title, status, ...}
    FE->>U: Redirect to /tests/:id
```

### Прохождение теста и аналитика

```mermaid
sequenceDiagram
    participant U as User
    participant FE as Frontend
    participant BE as Backend
    participant PG as PostgreSQL

    U->>FE: Открывает тест
    FE->>BE: GET /api/tests/:id
    BE->>PG: SELECT test + questions
    PG-->>BE: test data
    BE->>PG: SELECT my-answer (если авторизован)
    PG-->>BE: existing answer or null
    BE-->>FE: {test, myAnswer}
    FE->>U: Рендер вопросов по типу

    U->>FE: Заполняет и отправляет ответы
    FE->>BE: POST /api/tests/:id/answers<br/>{answers: {q_id: value}}
    BE->>BE: Validate test exists

    BE->>PG: INSERT INTO test_answers<br/>ON CONFLICT (test_id, user_id) DO UPDATE
    Note over PG: Upsert — один пользователь,<br/>один ответ на тест

    BE->>PG: UPDATE tests SET pass_count = pass_count + 1
    PG-->>BE: OK
    BE-->>FE: 201 Created {id, answers, score, result}

    Note over PG: Фоновый worker (асинхронно)
    PG->>PG: Пересчёт корреляций<br/>demographic ↔ test_answer
```

### Жизненный цикл авторизации на клиенте

```mermaid
stateDiagram-v2
    [*] --> CheckToken: Загрузка страницы

    CheckToken --> Authenticated: JWT valid + exp > now
    CheckToken --> Anonymous: no token / expired

    Authenticated --> MakeRequest: Пользователь выполняет действие
    MakeRequest --> Success: 200 OK
    MakeRequest --> TokenExpired: 401 Unauthorized

    TokenExpired --> RefreshPending: tryRefresh() singleton
    RefreshPending --> RefreshSuccess: новый access token
    RefreshPending --> LoggedOut: refresh истёк / revoked

    RefreshSuccess --> RetryRequest: повтор исходного запроса
    RetryRequest --> Success
    LoggedOut --> Anonymous: clearToken() + redirect /auth

    Anonymous --> Authenticated: POST /auth/verify → setToken()

    Success --> MakeRequest: следующее действие
```

## Развертывание

### Топология Docker Compose

```mermaid
graph TB
    subgraph "Docker Network: stat-platform-net"
        subgraph "Reverse Proxy"
            NGX[nginx-dev<br/>Port: 8000:8000]
        end

        subgraph "Application Services"
            FE_CTR[frontend<br/>Vite dev server<br/>:5173]
            BE_CTR[backend<br/>Go binary<br/>:8080]
            AU_CTR[auth-service<br/>Go binary<br/>:8081]
        end

        subgraph "Data Services"
            PG_AUTH[postgres-auth<br/>:5432<br/>auth_db]
            PG_MAIN[postgres-main<br/>:5433→:5432<br/>main_db]
            RD[redis<br/>:6379]
        end
    end

    Browser[Browser] --> NGX
    NGX --> FE_CTR & BE_CTR & AU_CTR
    BE_CTR --> PG_MAIN
    AU_CTR --> PG_AUTH & RD

    PG_AUTH -.->|Healthcheck<br/>pg_isready| AU_CTR
    PG_MAIN -.->|Healthcheck<br/>pg_isready| BE_CTR
    RD -.->|Healthcheck<br/>redis-cli ping| AU_CTR
```

**Healthcheck зависимости:**

Auth Service и Backend Service стартуют только после прохождения healthcheck соответствующих баз данных, что предотвращает панику при cold start кластера. Redis также проверяется перед стартом Auth Service.

### Переменные окружения

**Auth Service:**

| Переменная | Описание | Пример |
|-----------|----------|--------|
| `HTTP_PORT` | Порт HTTP сервера | `8081` |
| `DATABASE_URL` | PostgreSQL DSN | `postgres://auth_user:auth_pass@postgres-auth:5432/auth_db?sslmode=disable` |
| `JWT_SECRET` | HMAC-SHA256 ключ (min 32 chars) | `dev-secret-change-in-production-min-32-chars` |
| `CONTACT_PEPPER` | Pepper для bcrypt | `dev-pepper-change-in-production` |
| `ACCESS_TOKEN_TTL` | TTL access token | `15m` |
| `REFRESH_TOKEN_TTL` | TTL refresh token | `168h` |
| `REDIS_ADDR` | Адрес Redis | `redis:6379` |

**Backend Service:**

| Переменная | Описание | Пример |
|-----------|----------|--------|
| `HTTP_PORT` | Порт HTTP сервера | `8080` |
| `DATABASE_URL` | PostgreSQL DSN | `postgres://main_user:main_pass@postgres-main:5432/main_db?sslmode=disable` |
| `JWT_SECRET` | Тот же ключ что у Auth Service | `dev-secret-change-in-production-min-32-chars` |

**Frontend (Vite):**

| Переменная | Описание | Пример |
|-----------|----------|--------|
| `VITE_API_BASE` | Базовый URL API (если не через Nginx) | `/api` |
| `NODE_ENV` | Окружение | `development` |

### Миграции базы данных

Миграции применяются автоматически через `docker-entrypoint-initdb.d` при первом старте контейнеров PostgreSQL:
Порядок применения гарантирован числовым префиксом файлов.


## Лицензия

MIT License
