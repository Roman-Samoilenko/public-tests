package formsImporter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"quiz-platform/internal/domain"
)

// Регэксп для извлечения FB_PUBLIC_LOAD_DATA_ из HTML страницы.
// Google вставляет данные в виде:
//
//	var FB_PUBLIC_LOAD_DATA_ = [...];
var reFormData = regexp.MustCompile(`FB_PUBLIC_LOAD_DATA_\s*=\s*(\[[\s\S]*?\]);\s*</script>`)

// Типы вопросов Google Forms (внутренняя нумерация).
const (
	gfTypeShortText = 0
	gfTypeParagraph = 1
	gfTypeRadio     = 2 // одиночный выбор
	gfTypeDropdown  = 3 // одиночный выбор (выпадающий список)
	gfTypeCheckbox  = 4 // множественный выбор (флажки)
	gfTypeScale     = 5 // linear scale
	gfTypeGrid      = 7 // multiple choice grid → vector_scale
	gfTypeDate      = 9
	gfTypeTime      = 10
	gfTypeRating    = 18
)

// GoogleFormsImporter получает и парсит публичные Google Forms.
type GoogleFormsImporter struct {
	client *http.Client
}

func NewGoogleFormsImporter(client *http.Client) *GoogleFormsImporter {
	return &GoogleFormsImporter{
		client: client,
	}
}

// Import принимает любую ссылку на Google Form и возвращает структуру теста.
func (g *GoogleFormsImporter) Import(ctx context.Context, rawURL string) (*domain.ImportedTest, error) {
	viewURL, err := normalizeGoogleFormURL(ctx, rawURL)
	if err != nil {
		return nil, err
	}

	html, err := g.fetchHTML(viewURL)
	if err != nil {
		return nil, fmt.Errorf("fetch form: %w", err)
	}

	raw, err := extractFormData(html)
	if err != nil {
		return nil, fmt.Errorf("extract form data: %w", err)
	}

	result, err := parseFormData(raw)
	if err != nil {
		return nil, fmt.Errorf("parse form data: %w", err)
	}

	result.SourceURL = viewURL
	result.SourceType = "google_forms"
	return result, nil
}

// --- шаг 1: нормализация URL ---

// normalizeGoogleFormURL принимает любой формат ссылки на Google Form
// и возвращает стандартный /viewform URL.
//
// Поддерживаемые форматы:
//
//	https://docs.google.com/forms/d/FORM_ID/viewform
//	https://docs.google.com/forms/d/FORM_ID/edit
//	https://forms.gle/SHORT_CODE  (редиректим через HEAD запрос)
func normalizeGoogleFormURL(ctx context.Context, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("empty URL")
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	host := strings.ToLower(parsed.Host)

	// Короткая ссылка forms.gle
	if host == "forms.gle" {
		resolved, err := resolveShortURLWithGet(ctx, raw)
		if err != nil {
			return "", fmt.Errorf("resolve short URL: %w", err)
		}
		return normalizeGoogleFormURL(ctx, resolved)
	}

	if host != "docs.google.com" {
		return "", fmt.Errorf("not a Google Forms URL: %s", host)
	}

	// Парсим путь
	path := strings.Trim(parsed.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[0] != "forms" || parts[1] != "d" {
		return "", fmt.Errorf("unexpected Google Forms path: %s", parsed.Path)
	}

	var formID string
	if len(parts) >= 4 && parts[2] == "e" {
		// Формат: /forms/d/e/FORM_ID/...
		formID = parts[3]
	} else {
		// Формат: /forms/d/FORM_ID/...
		formID = parts[2]
	}

	if formID == "" {
		return "", errors.New("could not extract form ID from URL")
	}

	// Всегда возвращаем канонический viewform URL (поддерживаем оба формата)
	// Google отдаёт данные для обоих, но безопаснее вернуть /forms/d/e/... если он был
	if len(parts) >= 4 && parts[2] == "e" {
		return fmt.Sprintf("https://docs.google.com/forms/d/e/%s/viewform", formID), nil
	}
	return fmt.Sprintf("https://docs.google.com/forms/d/%s/viewform", formID), nil
}

// --- шаг 2: загрузка HTML ---

func (g *GoogleFormsImporter) fetchHTML(formURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, formURL, nil)
	if err != nil {
		return "", err
	}

	// Максимально имитируем браузер
	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Cache-Control", "max-age=0")

	slog.Debug("fetching Google Form", "url", formURL)

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	// Логируем финальный URL после всех редиректов
	finalURL := resp.Request.URL.String()
	slog.Debug("google form response", "status", resp.Status, "final_url", finalURL)

	if resp.StatusCode != http.StatusOK {
		// Читаем тело ответа для отладки (первые 500 символов)
		bodyPreview := ""
		if body, err := io.ReadAll(io.LimitReader(resp.Body, 500)); err == nil {
			bodyPreview = string(body)
		}
		slog.Error("unexpected status from Google Forms",
			"status", resp.StatusCode,
			"url", formURL,
			"final_url", finalURL,
			"body_preview", bodyPreview,
		)
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	return string(body), nil
}

// --- шаг 3: извлечение JSON из HTML ---

func extractFormData(html string) ([]any, error) {
	matches := reFormData.FindStringSubmatch(html)
	if len(matches) < 2 {
		return nil, errors.New("FB_PUBLIC_LOAD_DATA_ not found; form may be private or URL is invalid")
	}

	var data []any
	if err := json.Unmarshal([]byte(matches[1]), &data); err != nil {
		return nil, fmt.Errorf("unmarshal form JSON: %w", err)
	}
	return data, nil
}

func parseFormData(data []any) (*domain.ImportedTest, error) {
	mainBlock, err := getArray(data, 1)
	if err != nil {
		return nil, fmt.Errorf("no main block: %w", err)
	}

	title, ok := getString(mainBlock, 8)
	if !ok {
		title = "Untitled Form"
	}
	desc, ok := getString(mainBlock, 0)
	if !ok {
		desc = ""
	}
	slog.Debug("mainBlock length", "len", len(mainBlock))

	questionsRaw, err := getArray(mainBlock, 1)
	if err != nil {
		slog.Warn("no questions array at index 1", "err", err, "mainBlock_len", len(mainBlock))
		for i, v := range mainBlock {
			if arr, ok := v.([]any); ok {
				slog.Debug("array found in mainBlock", "index", i, "len", len(arr))
			}
		}
		// Нет вопросов — не ошибка, просто пустой тест
		questionsRaw = []any{}
	}

	slog.Debug("questions raw count", "count", len(questionsRaw))

	questions := make([]domain.Question, 0, len(questionsRaw))
	for i, qRaw := range questionsRaw {
		qArr, ok := qRaw.([]any)
		if !ok {
			slog.Warn("question is not array", "index", i, "type", fmt.Sprintf("%T", qRaw))

			continue
		}
		q, err := parseQuestion(qArr)
		if err != nil {
			slog.Warn("failed to parse question", "index", i, "err", err, "raw", qArr)

			// Пропускаем нераспознанный вопрос, не падаем
			continue
		}
		if q == nil {
			slog.Debug("skipping non-question item", "index", i)
			continue
		}
		if q.ID == "" {
			q.ID = fmt.Sprintf("q%d", i+1)
		}
		questions = append(questions, *q)
	}

	return &domain.ImportedTest{
		Title:       title,
		Description: desc,
		Questions:   questions,
	}, nil
}

func parseQuestion(q []any) (*domain.Question, error) {
	if len(q) < 5 {
		return nil, errors.New("question array too short")
	}

	idRaw, _ := q[0].(float64)
	text, _ := q[1].(string)
	qType := int(getFloat(q, 3))

	required := false
	if subBlocks, err := getArray(q, 4); err == nil && len(subBlocks) > 0 {
		if firstSub, ok := subBlocks[0].([]any); ok && len(firstSub) > 2 {
			required = getFloat(firstSub, 2) == 1
		}
	}

	result := &domain.Question{
		ID:       strconv.FormatInt(int64(idRaw), 10),
		Text:     text,
		Required: required,
	}

	switch qType {

	case 6, 8, 11, 12:
		return nil, nil

	case gfTypeShortText, gfTypeParagraph:
		result.Type = domain.QuestionTypeText

	case gfTypeRadio, gfTypeDropdown:
		result.Type = domain.QuestionTypeSingleChoice
		result.Options = parseOptions(q)

	case gfTypeCheckbox:
		result.Type = domain.QuestionTypeMultipleChoice
		result.Options = parseOptions(q)

	case gfTypeScale:
		result.Type = domain.QuestionTypeScale
		minVal, maxVal, minLabel, maxLabel := parseScale(q)
		result.MinValue = &minVal
		result.MaxValue = &maxVal
		result.MinLabel = minLabel
		result.MaxLabel = maxLabel

	case gfTypeGrid:
		rows, cols, multi := parseGrid(q)
		result.Type = domain.QuestionTypeVectorScale
		result.Rows = rows
		result.Cols = cols
		result.GridMultiple = multi
		// Сохраняем флаг множественного выбора в отдельном поле, если нужно
		// Можно добавить в Question структуру поле GridMultiple bool
		// Но проще на фронтенде определить по типу options?
		// Но для сетки варианты не хранятся как options.
		// Предлагаю добавить в domain.Question поле GridMultiple, или оставить векторную шкалу,
		// а на фронтенде по наличию rows/cols и отсутствию options определять как сетку.

	case gfTypeRating:
		result.Type = domain.QuestionTypeScale
		minVal, maxVal := parseRatingRange(q)
		result.MinValue = &minVal
		result.MaxValue = &maxVal

	default:
		result.Type = domain.QuestionTypeText
	}

	return result, nil
}

func parseRatingRange(q []any) (minVal, maxVal int) {
	minVal = 1
	maxVal = 5
	subBlocks, err := getArray(q, 4)
	if err != nil || len(subBlocks) == 0 {
		return
	}
	firstSub, ok := subBlocks[0].([]any)
	if !ok || len(firstSub) < 2 {
		return
	}
	optList, err := getArray(firstSub, 1)
	if err != nil || len(optList) == 0 {
		return
	}
	maxVal = len(optList)
	return
}
func parseGrid(q []any) (rows, cols []string, multiple bool) {
	subBlocks, err := getArray(q, 4)
	if err != nil || len(subBlocks) == 0 {
		return
	}
	// Берём первый блок строки для извлечения столбцов
	firstRowBlock, ok := subBlocks[0].([]any)
	if !ok || len(firstRowBlock) < 4 {
		return
	}
	// Столбцы – индекс 1 (массив массивов)
	if colList, err := getArray(firstRowBlock, 1); err == nil {
		for _, c := range colList {
			if arr, ok := c.([]any); ok && len(arr) > 0 {
				if text, ok := arr[0].(string); ok && text != "" {
					cols = append(cols, text)
				}
			}
		}
	}
	// Проходим по всем строкам для извлечения названий строк и флага multiple
	for _, rowBlock := range subBlocks {
		rArr, ok := rowBlock.([]any)
		if !ok || len(rArr) < 4 {
			continue
		}
		// Название строки – индекс 3 (массив)
		if nameArr, err := getArray(rArr, 3); err == nil && len(nameArr) > 0 {
			if name, ok := nameArr[0].(string); ok && name != "" {
				rows = append(rows, name)
			}
		}
		// Флаг множественного выбора – последний элемент (обычно bool)
		// Флаг "множественный выбор" — массив [0] или [1] в последнем элементе строки
		if len(rArr) > 0 {
			if flagArr, ok := rArr[len(rArr)-1].([]any); ok && len(flagArr) > 0 {
				if f, ok := flagArr[0].(float64); ok && f == 1 {
					multiple = true
				}
			}
		}
	}
	return
}

// parseOptions извлекает варианты ответов из вопроса типа radio/checkbox/dropdown.
func parseOptions(q []any) []domain.QuestionOption {
	opts := make([]domain.QuestionOption, 0)

	// Новая структура: q[4][0][1] — массив опций
	subBlocks, err := getArray(q, 4)
	if err != nil || len(subBlocks) == 0 {
		return opts
	}

	firstSub, ok := subBlocks[0].([]any)
	if !ok || len(firstSub) < 2 {
		return opts
	}

	optList, err := getArray(firstSub, 1)
	if err != nil {
		return opts
	}

	for i, o := range optList {
		optArr, ok := o.([]any)
		if !ok || len(optArr) == 0 {
			continue
		}
		text, _ := optArr[0].(string)
		if text == "" {
			continue // пропускаем пустые опции (например "Other")
		}
		opts = append(opts, domain.QuestionOption{
			ID:   fmt.Sprintf("opt%d", i+1),
			Text: text,
		})
	}
	return opts
}

func parseScale(q []any) (minL, maxL int, minLabel, maxLabel string) {
	minL, maxL = 1, 5

	subBlocks, err := getArray(q, 4)
	if err != nil || len(subBlocks) == 0 {
		return minL, maxL, minLabel, maxLabel
	}

	firstSub, ok := subBlocks[0].([]any)
	if !ok || len(firstSub) < 2 {
		return minL, maxL, minLabel, maxLabel
	}

	optList, err := getArray(firstSub, 1)
	if err != nil || len(optList) < 2 {
		return minL, maxL, minLabel, maxLabel
	}

	if v, ok := optList[0].([]any); ok && len(v) > 0 {
		if s, ok := v[0].(string); ok {
			if n, convErr := strconv.Atoi(s); convErr == nil {
				minL = n
			}
		}
	}
	if v, ok := optList[len(optList)-1].([]any); ok && len(v) > 0 {
		if s, ok := v[0].(string); ok {
			if n, convErr := strconv.Atoi(s); convErr == nil {
				maxL = n
			}
		}
	}

	// Метки шкалы
	if len(firstSub) > 3 {
		minLabel, _ = firstSub[3].(string)
	}
	if len(firstSub) > 4 {
		maxLabel, _ = firstSub[4].(string)
	}
	return minL, maxL, minLabel, maxLabel
}

// --- утилиты для работы с []any ---

func getArray(arr []any, idx int) ([]any, error) {
	if idx >= len(arr) {
		return nil, fmt.Errorf("index %d out of range (len=%d)", idx, len(arr))
	}
	v, ok := arr[idx].([]any)
	if !ok {
		return nil, fmt.Errorf("element at %d is not array", idx)
	}
	return v, nil
}

func getString(arr []any, idx int) (string, bool) {
	if idx >= len(arr) {
		return "", false
	}
	s, ok := arr[idx].(string)
	return s, ok
}

func getFloat(arr []any, idx int) float64 {
	if idx >= len(arr) {
		return 0
	}
	v, _ := arr[idx].(float64)
	return v
}

// resolveShortURLWithGet выполняет GET-запрос с автоматическим редиректом
// и возвращает финальный URL (например, docs.google.com/forms/d/e/.../viewform).
func resolveShortURLWithGet(ctx context.Context, shortURL string) (string, error) {
	// Используем контекст и timeout вместе
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client := &http.Client{
		// Таймаут на уровне клиента как fallback
		Timeout: 10 * time.Second,
		// Ограничиваем количество редиректов
		CheckRedirect: func(_ *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("too many redirects (max 10)")
			}
			return nil // разрешить редирект
		},
	}

	// Используем контекст для запроса
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, shortURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// После всех редиректов в resp.Request.URL лежит конечный адрес
	finalURL := resp.Request.URL.String()
	if finalURL == "" {
		return "", errors.New("empty final URL")
	}

	return finalURL, nil
}
