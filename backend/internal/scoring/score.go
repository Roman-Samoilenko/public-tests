package scoring

import "encoding/json"

// calculateScore вычисляет результат в режиме "score".
// Итог — сумма весов ответов, клампинг в [cfg.Min, cfg.Max],
// поиск уровня из cfg.Levels.
func calculateScore(
	cfg ResultConfig,
	questions []Question,
	answers map[string]interface{},
) (json.RawMessage, error) {
	var total float64

	for _, q := range questions {
		ans, ok := answers[q.ID]
		if !ok {
			continue
		}
		// Пропускаем матричные вопросы
		if q.Type == "vector_scale" {
			continue
		}

		switch q.Type {
		case "single_choice":
			selectedID, ok := ans.(string)
			if !ok {
				continue
			}
			for _, opt := range q.Options {
				if opt.ID == selectedID {
					total += opt.Scoring["score"]
					break
				}
			}

		case "multiple_choice":
			selectedIDs := toStringSlice(ans)
			selectedSet := make(map[string]bool, len(selectedIDs))
			for _, id := range selectedIDs {
				selectedSet[id] = true
			}
			for _, opt := range q.Options {
				if selectedSet[opt.ID] {
					total += opt.Scoring["score"]
				}
			}

		case "scale":
			value := toFloat(ans)
			coeff := q.Scoring["score"]
			total += value * coeff

			// text — не участвует в scoring
		}
	}

	// Клампинг
	total = clamp(total, cfg.Min, cfg.Max)

	// Поиск уровня
	level := ""
	for _, lvl := range cfg.Levels {
		if total >= lvl.Min && total <= lvl.Max {
			level = lvl.Label
			break
		}
	}

	result := ScoreResult{
		Type:  "score",
		Value: total,
		Level: level,
	}

	return json.Marshal(result)
}
