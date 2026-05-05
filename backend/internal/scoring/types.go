package scoring

import "encoding/json"

// -------------------------------------------------------
// Конфигурация результата (десериализованная из result_config JSONB)
// -------------------------------------------------------

// ResultConfig — общий конверт для всех режимов.
type ResultConfig struct {
	Type string `json:"type"` // "score" | "scales" | "string_map" | "combo"
	// Режим score
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Levels []Level `json:"levels"`
	// Режимы с осями
	Axes []Axis `json:"axes"`
	// Режимы с типами
	Results []ResultEntry `json:"results"`
	// Текстовое пояснение автора (любой режим)
	Description string `json:"description"`
}

// Level — диапазон баллов с меткой (режим score).
type Level struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Label string  `json:"label"`
}

// Axis — одна ось шкалы.
type Axis struct {
	ID         string  `json:"id"`
	Label      string  `json:"label"`
	Min        float64 `json:"min"`
	Max        float64 `json:"max"`
	LeftLabel  string  `json:"left_label"` // показывается если min < 0
	RightLabel string  `json:"right_label"`
}

// ResultEntry — один именованный результат (режимы string_map / combo).
type ResultEntry struct {
	Key         string             `json:"key"`
	Label       string             `json:"label"`
	Description string             `json:"description"`
	Target      map[string]float64 `json:"target"` // axis_id → желаемое значение
}

// -------------------------------------------------------
// Вопросы (десериализованные из questions JSONB)
// -------------------------------------------------------

// Question — вопрос теста с полем scoring.
type Question struct {
	ID      string   `json:"id"`
	Type    string   `json:"type"` // "single_choice" | "multiple_choice" | "scale" | "text"
	Options []Option `json:"options"`
	// Для типа scale: коэффициент на уровне вопроса
	Scoring map[string]float64 `json:"scoring"`
}

// Option — вариант ответа с весами.
type Option struct {
	ID      string             `json:"id"`
	Scoring map[string]float64 `json:"scoring"` // axis_id/score → вес
}

// -------------------------------------------------------
// Выходные форматы результата (сохраняются в test_answers.result)
// -------------------------------------------------------

// ScoreResult — результат режима score.
type ScoreResult struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
	Level string  `json:"level,omitempty"`
}

// ScalesResult — результат режима scales.
type ScalesResult struct {
	Type string             `json:"type"`
	Axes map[string]float64 `json:"axes"`
}

// StringMapResult — результат режимов string_map и combo.
type StringMapResult struct {
	Type        string             `json:"type"`
	Axes        map[string]float64 `json:"axes,omitempty"`
	Matched     string             `json:"matched"`
	Label       string             `json:"label"`
	Description string             `json:"description"`
}

// -------------------------------------------------------
// Вспомогательные функции
// -------------------------------------------------------

// parseConfig десериализует result_config из JSONB.
func parseConfig(raw json.RawMessage) (ResultConfig, error) {
	var cfg ResultConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// parseQuestions десериализует questions из JSONB.
func parseQuestions(raw json.RawMessage) ([]Question, error) {
	var qs []Question
	if err := json.Unmarshal(raw, &qs); err != nil {
		return nil, err
	}
	return qs, nil
}

// toFloat пробует привести произвольное значение к float64.
// Строки, числа, булевы — всё обрабатывается.
// При неудаче возвращает 0.
func toFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case json.Number:
		f, _ := val.Float64()
		return f
	case string:
		// JSON числа иногда приходят как строки из JS
		var n = json.Number(val)
		f, _ := n.Float64()
		return f
	case bool:
		if val {
			return 1
		}
		return 0
	}
	return 0
}

// clamp ограничивает значение в диапазоне [minVal, maxVal].
func clamp(v, minVal, maxVal float64) float64 {
	if v < minVal {
		return minVal
	}
	if v > maxVal {
		return maxVal
	}
	return v
}

// toStringSlice безопасно превращает interface{} в []string.
// Используется для multiple_choice ответов.
func toStringSlice(v interface{}) []string {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case []interface{}:
		out := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return val
	case string:
		if val == "" {
			return nil
		}
		return []string{val}
	}
	return nil
}
