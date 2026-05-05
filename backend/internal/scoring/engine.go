package scoring

import "encoding/json"

// Calculate вычисляет результат прохождения теста.
//
// Возвращает nil, nil если:
//   - configRaw пустой или nil
//   - type не задан или равен "none"
//   - questions пустые
//
// При ошибке парсинга конфига или вопросов возвращает nil, err.
// Ошибки самого вычисления (неизвестный тип ответа и т.п.) мягко игнорируются —
// возвращается частичный или нулевой результат.
func Calculate(
	configRaw json.RawMessage,
	questionsRaw json.RawMessage,
	answers map[string]interface{},
) (json.RawMessage, error) {
	// Нет конфига — scoring не настроен
	if len(configRaw) == 0 || string(configRaw) == "null" {
		return nil, nil
	}

	cfg, err := parseConfig(configRaw)
	if err != nil {
		return nil, err
	}

	if cfg.Type == "" || cfg.Type == "none" {
		return nil, nil
	}

	// Нет вопросов — нечего считать
	if len(questionsRaw) == 0 || string(questionsRaw) == "null" || string(questionsRaw) == "[]" {
		return nil, nil
	}

	questions, err := parseQuestions(questionsRaw)
	if err != nil {
		return nil, err
	}

	if answers == nil {
		answers = map[string]interface{}{}
	}

	switch cfg.Type {
	case "score":
		return calculateScore(cfg, questions, answers)
	case "scales":
		return calculateScalesResult(cfg, questions, answers)
	case "string_map":
		return calculateStringMapResult(cfg, questions, answers)
	case "combo":
		return calculateComboResult(cfg, questions, answers)
	default:
		// Неизвестный тип — не падаем
		return nil, nil
	}
}
