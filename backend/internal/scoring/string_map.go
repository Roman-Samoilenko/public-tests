package scoring

import (
	"encoding/json"
	"math"
)

// FindClosest находит ResultEntry с минимальным евклидовым расстоянием
// от вычисленных значений осей до target каждого результата.
// Расстояние считается по нормализованным значениям [0,1] чтобы
// оси с разными диапазонами влияли пропорционально.
func FindClosest(axisValues map[string]float64, axes []Axis, results []ResultEntry) (ResultEntry, bool) {
	if len(results) == 0 {
		return ResultEntry{}, false
	}

	// Строим map для быстрого доступа к метаданным оси
	axisMap := make(map[string]Axis, len(axes))
	for _, ax := range axes {
		axisMap[ax.ID] = ax
	}

	// normalize приводит значение к диапазону [0,1] для данной оси
	normalize := func(axID string, val float64) float64 {
		ax, ok := axisMap[axID]
		if !ok {
			return 0
		}
		span := ax.Max - ax.Min
		if span == 0 {
			return 0
		}
		return (val - ax.Min) / span
	}

	bestIdx := 0
	bestDist := math.MaxFloat64

	for i, res := range results {
		var sumSq float64
		for axID, targetVal := range res.Target {
			actualNorm := normalize(axID, axisValues[axID])
			targetNorm := normalize(axID, targetVal)
			diff := actualNorm - targetNorm
			sumSq += diff * diff
		}
		dist := math.Sqrt(sumSq)
		if dist < bestDist {
			bestDist = dist
			bestIdx = i
		}
	}

	return results[bestIdx], true
}

// calculateStringMapResult формирует итоговый JSON для режима "string_map".
func calculateStringMapResult(
	cfg ResultConfig,
	questions []Question,
	answers map[string]interface{},
) (json.RawMessage, error) {
	axes := calculateAxes(cfg, questions, answers)

	matched, ok := FindClosest(axes, cfg.Axes, cfg.Results)
	if !ok {
		// Нет результатов в конфиге — возвращаем только оси
		return calculateScalesResult(cfg, questions, answers)
	}

	result := StringMapResult{
		Type:        "string_map",
		Matched:     matched.Key,
		Label:       matched.Label,
		Description: matched.Description,
	}
	return json.Marshal(result)
}

// calculateComboResult формирует итоговый JSON для режима "combo":
// и оси, и ближайший тип.
func calculateComboResult(
	cfg ResultConfig,
	questions []Question,
	answers map[string]interface{},
) (json.RawMessage, error) {
	axes := calculateAxes(cfg, questions, answers)

	matched, ok := FindClosest(axes, cfg.Axes, cfg.Results)
	if !ok {
		// Нет results — отдаём как scales
		result := ScalesResult{Type: "combo", Axes: axes}
		return json.Marshal(result)
	}

	result := StringMapResult{
		Type:        "combo",
		Axes:        axes,
		Matched:     matched.Key,
		Label:       matched.Label,
		Description: matched.Description,
	}
	return json.Marshal(result)
}
