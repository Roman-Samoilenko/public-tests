package scoring

import "encoding/json"

// calculateAxes вычисляет значения по всем осям.
// Используется режимами scales, string_map и combo.
// Возвращает map[axis_id]value с клампингом в [axis.Min, axis.Max].
func calculateAxes(cfg ResultConfig, questions []Question, answers map[string]interface{}) map[string]float64 {
	// Инициализируем нулями все оси
	axisValues := make(map[string]float64, len(cfg.Axes))
	for _, ax := range cfg.Axes {
		axisValues[ax.ID] = 0
	}

	for _, q := range questions {
		ans, ok := answers[q.ID]
		if !ok {
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
					for axID, weight := range opt.Scoring {
						if _, exists := axisValues[axID]; exists {
							axisValues[axID] += weight
						}
					}
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
					for axID, weight := range opt.Scoring {
						if _, exists := axisValues[axID]; exists {
							axisValues[axID] += weight
						}
					}
				}
			}

		case "scale":
			value := toFloat(ans)
			for axID, coeff := range q.Scoring {
				if _, exists := axisValues[axID]; exists {
					axisValues[axID] += value * coeff
				}
			}
		}
	}

	// Клампинг по диапазонам осей
	axisMap := make(map[string]Axis, len(cfg.Axes))
	for _, ax := range cfg.Axes {
		axisMap[ax.ID] = ax
	}
	for axID, val := range axisValues {
		if ax, ok := axisMap[axID]; ok {
			axisValues[axID] = clamp(val, ax.Min, ax.Max)
		}
	}

	return axisValues
}

// calculateScalesResult формирует итоговый JSON для режима "scales".
func calculateScalesResult(
	cfg ResultConfig,
	questions []Question,
	answers map[string]interface{},
) (json.RawMessage, error) {
	axes := calculateAxes(cfg, questions, answers)
	result := ScalesResult{
		Type: "scales",
		Axes: axes,
	}
	return json.Marshal(result)
}
