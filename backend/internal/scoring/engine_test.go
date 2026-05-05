package scoring_test

import (
	"encoding/json"
	"math"
	sc "quiz-platform/internal/scoring"
	"testing"
)

// -------------------------------------------------------
// Helpers
// -------------------------------------------------------

func mustMarshal(t *testing.T, v interface{}) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

func unmarshalScore(t *testing.T, raw json.RawMessage) sc.ScoreResult {
	t.Helper()
	var r sc.ScoreResult
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("unmarshal ScoreResult: %v", err)
	}
	return r
}

func unmarshalScales(t *testing.T, raw json.RawMessage) sc.ScalesResult {
	t.Helper()
	var r sc.ScalesResult
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("unmarshal ScalesResult: %v", err)
	}
	return r
}

func unmarshalStringMap(t *testing.T, raw json.RawMessage) sc.StringMapResult {
	t.Helper()
	var r sc.StringMapResult
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("unmarshal StringMapResult: %v", err)
	}
	return r
}

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.001
}

// -------------------------------------------------------
// Режим score
// -------------------------------------------------------

func TestScore_SingleChoice(t *testing.T) {
	cfg := mustMarshal(t, map[string]interface{}{
		"type": "score", "min": 0, "max": 100,
		"levels": []map[string]interface{}{
			{"min": 0, "max": 30, "label": "Низкий"},
			{"min": 31, "max": 70, "label": "Средний"},
			{"min": 71, "max": 100, "label": "Высокий"},
		},
	})
	questions := mustMarshal(t, []map[string]interface{}{
		{
			"id": "q1", "type": "single_choice",
			"options": []map[string]interface{}{
				{"id": "a", "scoring": map[string]interface{}{"score": 40}},
				{"id": "b", "scoring": map[string]interface{}{"score": 10}},
			},
		},
		{
			"id": "q2", "type": "single_choice",
			"options": []map[string]interface{}{
				{"id": "a", "scoring": map[string]interface{}{"score": 30}},
				{"id": "b", "scoring": map[string]interface{}{"score": 5}},
			},
		},
	})
	answers := map[string]interface{}{"q1": "a", "q2": "a"} // 40 + 30 = 70

	raw, err := sc.Calculate(cfg, questions, answers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := unmarshalScore(t, raw)

	if !approxEqual(r.Value, 70) {
		t.Errorf("expected value 70, got %f", r.Value)
	}
	if r.Level != "Средний" {
		t.Errorf("expected level Средний, got %q", r.Level)
	}
}

func TestScore_MultipleChoice(t *testing.T) {
	cfg := mustMarshal(t, map[string]interface{}{
		"type": "score", "min": 0, "max": 100,
	})
	questions := mustMarshal(t, []map[string]interface{}{
		{
			"id": "q1", "type": "multiple_choice",
			"options": []map[string]interface{}{
				{"id": "a", "scoring": map[string]interface{}{"score": 10}},
				{"id": "b", "scoring": map[string]interface{}{"score": 20}},
				{"id": "c", "scoring": map[string]interface{}{"score": 5}},
			},
		},
	})
	answers := map[string]interface{}{"q1": []interface{}{"a", "c"}} // 10 + 5 = 15

	raw, err := sc.Calculate(cfg, questions, answers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := unmarshalScore(t, raw)
	if !approxEqual(r.Value, 15) {
		t.Errorf("expected 15, got %f", r.Value)
	}
}

func TestScore_ScaleQuestion(t *testing.T) {
	cfg := mustMarshal(t, map[string]interface{}{
		"type": "score", "min": 0, "max": 100,
	})
	questions := mustMarshal(t, []map[string]interface{}{
		{
			"id": "q1", "type": "scale",
			"scoring": map[string]interface{}{"score": 5}, // коэффициент
		},
	})
	answers := map[string]interface{}{"q1": 8.0} // 8 * 5 = 40

	raw, err := sc.Calculate(cfg, questions, answers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := unmarshalScore(t, raw)
	if !approxEqual(r.Value, 40) {
		t.Errorf("expected 40, got %f", r.Value)
	}
}

func TestScore_Clamp(t *testing.T) {
	cfg := mustMarshal(t, map[string]interface{}{
		"type": "score", "min": 0, "max": 50,
	})
	questions := mustMarshal(t, []map[string]interface{}{
		{
			"id": "q1", "type": "single_choice",
			"options": []map[string]interface{}{
				{"id": "a", "scoring": map[string]interface{}{"score": 80}},
			},
		},
	})
	answers := map[string]interface{}{"q1": "a"}

	raw, err := sc.Calculate(cfg, questions, answers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := unmarshalScore(t, raw)
	if !approxEqual(r.Value, 50) { // зажато до max
		t.Errorf("expected 50 (clamped), got %f", r.Value)
	}
}

func TestScore_NoAnswers(t *testing.T) {
	cfg := mustMarshal(t, map[string]interface{}{
		"type": "score", "min": 0, "max": 100,
	})
	questions := mustMarshal(t, []map[string]interface{}{
		{
			"id": "q1", "type": "single_choice",
			"options": []map[string]interface{}{
				{"id": "a", "scoring": map[string]interface{}{"score": 50}},
			},
		},
	})
	answers := map[string]interface{}{} // нет ответов

	raw, err := sc.Calculate(cfg, questions, answers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := unmarshalScore(t, raw)
	if !approxEqual(r.Value, 0) {
		t.Errorf("expected 0, got %f", r.Value)
	}
}

func TestScore_TextQuestionIgnored(t *testing.T) {
	cfg := mustMarshal(t, map[string]interface{}{
		"type": "score", "min": 0, "max": 100,
	})
	questions := mustMarshal(t, []map[string]interface{}{
		{"id": "q1", "type": "text"},
	})
	answers := map[string]interface{}{"q1": "какой-то текст"}

	raw, err := sc.Calculate(cfg, questions, answers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := unmarshalScore(t, raw)
	if !approxEqual(r.Value, 0) {
		t.Errorf("text question should not contribute, got %f", r.Value)
	}
}

// -------------------------------------------------------
// Режим scales
// -------------------------------------------------------

func TestScales_Bipolar(t *testing.T) {
	cfg := mustMarshal(t, map[string]interface{}{
		"type": "scales",
		"axes": []map[string]interface{}{
			{"id": "democracy", "min": -50, "max": 50},
			{"id": "liberalism", "min": 0, "max": 100},
		},
	})
	questions := mustMarshal(t, []map[string]interface{}{
		{
			"id": "q1", "type": "single_choice",
			"options": []map[string]interface{}{
				{"id": "a", "scoring": map[string]interface{}{"democracy": 30, "liberalism": 20}},
				{"id": "b", "scoring": map[string]interface{}{"democracy": -20, "liberalism": 5}},
			},
		},
		{
			"id": "q2", "type": "single_choice",
			"options": []map[string]interface{}{
				{"id": "a", "scoring": map[string]interface{}{"democracy": 10}},
			},
		},
	})
	answers := map[string]interface{}{"q1": "a", "q2": "a"} // democracy=40, liberalism=20

	raw, err := sc.Calculate(cfg, questions, answers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := unmarshalScales(t, raw)
	if !approxEqual(r.Axes["democracy"], 40) {
		t.Errorf("democracy: expected 40, got %f", r.Axes["democracy"])
	}
	if !approxEqual(r.Axes["liberalism"], 20) {
		t.Errorf("liberalism: expected 20, got %f", r.Axes["liberalism"])
	}
}

func TestScales_Clamp(t *testing.T) {
	cfg := mustMarshal(t, map[string]interface{}{
		"type": "scales",
		"axes": []map[string]interface{}{
			{"id": "ax", "min": 0, "max": 10},
		},
	})
	questions := mustMarshal(t, []map[string]interface{}{
		{
			"id": "q1", "type": "single_choice",
			"options": []map[string]interface{}{
				{"id": "a", "scoring": map[string]interface{}{"ax": 999}},
			},
		},
	})
	answers := map[string]interface{}{"q1": "a"}

	raw, err := sc.Calculate(cfg, questions, answers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := unmarshalScales(t, raw)
	if !approxEqual(r.Axes["ax"], 10) {
		t.Errorf("expected clamped to 10, got %f", r.Axes["ax"])
	}
}

// -------------------------------------------------------
// Режим string_map
// -------------------------------------------------------

func TestStringMap_FindClosest(t *testing.T) {
	axes := []sc.Axis{
		{ID: "democracy", Min: -50, Max: 50},
		{ID: "liberalism", Min: 0, Max: 100},
	}
	results := []sc.ResultEntry{
		{Key: "liberal", Label: "Либерал", Target: map[string]float64{"democracy": 40, "liberalism": 80}},
		{Key: "authoritarian", Label: "Авторитарист", Target: map[string]float64{"democracy": -40, "liberalism": 20}},
		{Key: "centrist", Label: "Центрист", Target: map[string]float64{"democracy": 0, "liberalism": 50}},
	}

	axisValues := map[string]float64{"democracy": 35, "liberalism": 75}
	matched, ok := sc.FindClosest(axisValues, axes, results)
	if !ok {
		t.Fatal("expected match")
	}
	if matched.Key != "liberal" {
		t.Errorf("expected liberal, got %q", matched.Key)
	}
}

func TestStringMap_EmptyResults(t *testing.T) {
	cfg := mustMarshal(t, map[string]interface{}{
		"type":    "string_map",
		"axes":    []map[string]interface{}{{"id": "ax", "min": 0, "max": 100}},
		"results": []interface{}{},
	})
	questions := mustMarshal(t, []map[string]interface{}{})
	answers := map[string]interface{}{}

	// Не должно упасть — возвращает scales-результат
	_, err := sc.Calculate(cfg, questions, answers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// -------------------------------------------------------
// Режим combo
// -------------------------------------------------------

func TestCombo(t *testing.T) {
	cfg := mustMarshal(t, map[string]interface{}{
		"type": "combo",
		"axes": []map[string]interface{}{
			{"id": "democracy", "min": -50, "max": 50},
		},
		"results": []map[string]interface{}{
			{"key": "lib", "label": "Либерал", "target": map[string]interface{}{"democracy": 40}},
			{"key": "auth", "label": "Авторитарист", "target": map[string]interface{}{"democracy": -40}},
		},
	})
	questions := mustMarshal(t, []map[string]interface{}{
		{
			"id": "q1", "type": "single_choice",
			"options": []map[string]interface{}{
				{"id": "a", "scoring": map[string]interface{}{"democracy": 45}},
			},
		},
	})
	answers := map[string]interface{}{"q1": "a"}

	raw, err := sc.Calculate(cfg, questions, answers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := unmarshalStringMap(t, raw)
	if r.Type != "combo" {
		t.Errorf("expected type combo, got %q", r.Type)
	}
	if r.Matched != "lib" {
		t.Errorf("expected lib, got %q", r.Matched)
	}
	if r.Axes == nil {
		t.Error("combo result must include axes")
	}
}

// -------------------------------------------------------
// Граничные случаи Calculate
// -------------------------------------------------------

func TestCalculate_NilConfig(t *testing.T) {
	raw, err := sc.Calculate(nil, nil, nil)
	if err != nil || raw != nil {
		t.Errorf("nil config should return nil, nil; got %v, %v", raw, err)
	}
}

func TestCalculate_NullConfig(t *testing.T) {
	raw, err := sc.Calculate(json.RawMessage("null"), nil, nil)
	if err != nil || raw != nil {
		t.Errorf("null config should return nil, nil; got %v, %v", raw, err)
	}
}

func TestCalculate_NoneType(t *testing.T) {
	cfg := mustMarshal(t, map[string]interface{}{"type": "none"})
	raw, err := sc.Calculate(cfg, nil, nil)
	if err != nil || raw != nil {
		t.Errorf("type none should return nil, nil; got %v, %v", raw, err)
	}
}

func TestCalculate_UnknownType(t *testing.T) {
	cfg := mustMarshal(t, map[string]interface{}{"type": "future_mode"})
	qs := mustMarshal(t, []interface{}{})
	raw, err := sc.Calculate(cfg, qs, nil)
	if err != nil || raw != nil {
		t.Errorf("unknown type should return nil, nil; got %v, %v", raw, err)
	}
}

func TestCalculate_MissingAnswerKey(t *testing.T) {
	cfg := mustMarshal(t, map[string]interface{}{
		"type": "score", "min": 0, "max": 100,
	})
	questions := mustMarshal(t, []map[string]interface{}{
		{
			"id": "q1", "type": "single_choice",
			"options": []map[string]interface{}{
				{"id": "a", "scoring": map[string]interface{}{"score": 50}},
			},
		},
	})
	// Ответ на q1 отсутствует
	answers := map[string]interface{}{"q2": "a"}

	raw, err := sc.Calculate(cfg, questions, answers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := unmarshalScore(t, raw)
	if !approxEqual(r.Value, 0) {
		t.Errorf("missing answer should contribute 0, got %f", r.Value)
	}
}

func TestCalculate_NoScoringOnOption(t *testing.T) {
	cfg := mustMarshal(t, map[string]interface{}{
		"type": "score", "min": 0, "max": 100,
	})
	// Вариант без scoring — должен вернуть 0, не упасть
	questions := mustMarshal(t, []map[string]interface{}{
		{
			"id": "q1", "type": "single_choice",
			"options": []map[string]interface{}{
				{"id": "a"}, // нет поля scoring
			},
		},
	})
	answers := map[string]interface{}{"q1": "a"}

	raw, err := sc.Calculate(cfg, questions, answers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := unmarshalScore(t, raw)
	if !approxEqual(r.Value, 0) {
		t.Errorf("option without scoring should contribute 0, got %f", r.Value)
	}
}
