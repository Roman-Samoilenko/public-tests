<template>
  <div class="result-display">

    <!-- ── score ───────────────────────────────────────── -->
    <template v-if="result.type === 'score'">
      <p class="result-eyebrow font-mono">Результат</p>
      <div class="score-value">{{ formatted(result.value) }}</div>
      <div v-if="result.level" class="score-level">{{ result.level }}</div>

      <!-- Прогресс-бар -->
      <div class="progress-track" :title="`${formatted(result.value)} из ${formatted(cfgMax)}`">
        <div class="progress-fill" :style="{ width: scorePercent + '%' }"></div>
      </div>
      <div class="progress-labels font-mono">
        <span>{{ formatted(cfgMin) }}</span>
        <span>{{ formatted(cfgMax) }}</span>
      </div>
    </template>

    <!-- ── scales ──────────────────────────────────────── -->
    <template v-if="result.type === 'scales' || (result.type === 'combo' && result.axes)">
      <p class="result-eyebrow font-mono">{{ result.type === 'combo' ? 'Шкалы' : 'Результат' }}</p>
      <div class="axes-list">
        <div
          v-for="axis in resolvedAxes"
          :key="axis.id"
          class="axis-row"
        >
          <div class="axis-header">
            <!-- Биполярная: min < 0 — два лейбла -->
            <template v-if="axis.min < 0">
              <span class="axis-side-label font-mono left">{{ axis.left_label || axis.label }}</span>
              <span class="axis-center-label font-mono">{{ axis.label }}</span>
              <span class="axis-side-label font-mono right">{{ axis.right_label || axis.label }}</span>
            </template>
            <!-- Унополярная: один лейбл справа -->
            <template v-else>
              <span class="axis-label font-mono">{{ axis.right_label || axis.label }}</span>
            </template>
          </div>

          <div class="axis-track-wrap">
            <!-- Биполярная шкала: центр = 0 -->
            <template v-if="axis.min < 0">
              <div class="axis-track bipolar">
                <div
                  class="axis-fill bipolar-fill"
                  :style="bipolarStyle(axis)"
                ></div>
                <div class="axis-center-line"></div>
              </div>
            </template>

            <!-- Унополярная шкала -->
            <template v-else>
              <div class="axis-track">
                <div
                  class="axis-fill"
                  :style="{ width: unipolarPercent(axis) + '%' }"
                ></div>
              </div>
            </template>
          </div>

          <div class="axis-value font-mono">
            {{ formatAxisValue(result.axes[axis.id], axis) }}
          </div>
        </div>
      </div>
    </template>

    <!-- ── string_map / combo matched ─────────────────── -->
    <template v-if="result.type === 'string_map' || result.type === 'combo'">
      <div class="matched-card">
        <p class="matched-eyebrow font-mono">
          {{ result.type === 'combo' ? 'Ваш тип' : 'Результат' }}
        </p>
        <h2 class="matched-label">{{ result.label }}</h2>
        <p v-if="result.description" class="matched-desc">{{ result.description }}</p>
      </div>
    </template>

    <!-- ── Пояснение автора (любой режим) ─────────────── -->
    <div v-if="authorNote" class="author-note">
      <p class="author-note-label font-mono">Примечание автора</p>
      <p class="author-note-text">{{ authorNote }}</p>
    </div>

  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  // Распарсенный объект из AnswerResult.result
  result: { type: Object, required: true },
  // test.result_config — для доступа к метаданным осей и description
  resultConfig: { type: Object, default: null },
})

// ── Описание автора из result_config ──
const authorNote = computed(() => props.resultConfig?.description || '')

// ── Для режима score ──
const cfgMin = computed(() => props.resultConfig?.min ?? 0)
const cfgMax = computed(() => props.resultConfig?.max ?? 100)

const scorePercent = computed(() => {
  const span = cfgMax.value - cfgMin.value
  if (span === 0) return 0
  return Math.min(100, Math.max(0,
    ((props.result.value - cfgMin.value) / span) * 100
  ))
})

// ── Оси из result_config (нужны метаданные min/max/labels) ──
const resolvedAxes = computed(() => {
  const cfgAxes = props.resultConfig?.axes || []
  if (!props.result.axes) return []
  // Берём только те оси, для которых есть значение в результате
  return cfgAxes.filter(ax => props.result.axes[ax.id] !== undefined)
})

// ── Вычисление позиции для биполярной шкалы ──
function bipolarStyle(axis) {
  const val = props.result.axes[axis.id] ?? 0
  const span = axis.max - axis.min
  if (span === 0) return { left: '50%', width: '0%' }

  const zeroPos = ((0 - axis.min) / span) * 100    // % где находится 0
  const valPos  = ((val  - axis.min) / span) * 100  // % где находится значение

  if (valPos >= zeroPos) {
    return { left: zeroPos + '%', width: (valPos - zeroPos) + '%' }
  } else {
    return { left: valPos + '%', width: (zeroPos - valPos) + '%' }
  }
}

function unipolarPercent(axis) {
  const val = props.result.axes[axis.id] ?? 0
  const span = axis.max - axis.min
  if (span === 0) return 0
  return Math.min(100, Math.max(0, ((val - axis.min) / span) * 100))
}

function formatAxisValue(val, axis) {
  if (val === undefined || val === null) return '—'
  const v = Math.round(val * 10) / 10
  if (axis.min < 0 && v > 0) return '+' + v
  return String(v)
}

function formatted(n) {
  return Math.round((n ?? 0) * 10) / 10
}
</script>

<style scoped>
.result-display {
  display: flex;
  flex-direction: column;
  gap: 1.2rem;
  padding: 1.5rem;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 6px;
  margin-top: 1.5rem;
}

.result-eyebrow {
  font-family: var(--font-mono);
  font-size: 0.68rem;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: -0.4rem;
}

/* ── Score ─────────────────────────────────────────── */
.score-value {
  font-family: var(--font-serif);
  font-size: 3rem;
  font-weight: 700;
  line-height: 1;
  color: var(--accent);
}
.score-level {
  font-family: var(--font-mono);
  font-size: 0.85rem;
  color: var(--text-muted);
  letter-spacing: 0.06em;
}

.progress-track {
  height: 6px;
  background: var(--bg-input);
  border-radius: 3px;
  overflow: hidden;
  margin-top: 0.3rem;
}
.progress-fill {
  height: 100%;
  background: var(--accent);
  border-radius: 3px;
  transition: width 0.6s cubic-bezier(0.4, 0, 0.2, 1);
}
.progress-labels {
  display: flex;
  justify-content: space-between;
  font-family: var(--font-mono);
  font-size: 0.68rem;
  color: var(--text-muted);
  margin-top: 0.3rem;
}

/* ── Axes ───────────────────────────────────────────── */
.axes-list {
  display: flex;
  flex-direction: column;
  gap: 1.2rem;
}

.axis-row {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.axis-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}
.axis-label {
  font-family: var(--font-mono);
  font-size: 0.78rem;
  color: var(--text-muted);
}
.axis-side-label {
  font-family: var(--font-mono);
  font-size: 0.7rem;
  color: var(--text-muted);
  flex: 1;
}
.axis-side-label.right { text-align: right; }
.axis-center-label {
  font-family: var(--font-mono);
  font-size: 0.72rem;
  color: var(--accent);
  text-align: center;
  flex: 0 0 auto;
  padding: 0 0.5rem;
}

.axis-track-wrap { position: relative; }

/* Общий трек */
.axis-track {
  height: 6px;
  background: var(--bg-input);
  border-radius: 3px;
  position: relative;
  overflow: visible;
}

/* Биполярный заполнитель — может начинаться не с 0 */
.bipolar-fill {
  position: absolute;
  top: 0;
  height: 100%;
  background: var(--accent);
  border-radius: 3px;
  transition: left 0.6s cubic-bezier(0.4, 0, 0.2, 1),
              width 0.6s cubic-bezier(0.4, 0, 0.2, 1);
}

/* Унополярный заполнитель — от левого края */
.axis-fill:not(.bipolar-fill) {
  height: 100%;
  background: var(--accent);
  border-radius: 3px;
  transition: width 0.6s cubic-bezier(0.4, 0, 0.2, 1);
}

/* Вертикальная черта в центре для биполярных шкал */
.axis-center-line {
  position: absolute;
  top: -3px;
  bottom: -3px;
  left: 50%;
  width: 1px;
  background: var(--border);
  transform: translateX(-50%);
}

.axis-value {
  font-family: var(--font-mono);
  font-size: 0.78rem;
  color: var(--accent);
  align-self: flex-end;
}

/* ── Matched (string_map / combo) ───────────────────── */
.matched-card {
  padding: 1rem 0 0.5rem;
  border-top: 1px solid var(--border);
}
.matched-eyebrow {
  font-family: var(--font-mono);
  font-size: 0.68rem;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 0.4rem;
}
.matched-label {
  font-family: var(--font-serif);
  font-size: 1.8rem;
  font-weight: 700;
  color: var(--accent);
  margin-bottom: 0.5rem;
}
.matched-desc {
  font-size: 0.92rem;
  color: var(--text-muted);
  line-height: 1.6;
}

/* ── Пояснение автора ───────────────────────────────── */
.author-note {
  padding: 1rem;
  background: rgba(255,255,255,0.02);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  border-top: 1px solid var(--border);
}
.author-note-label {
  font-family: var(--font-mono);
  font-size: 0.65rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 0.4rem;
}
.author-note-text {
  font-size: 0.88rem;
  color: var(--text-muted);
  line-height: 1.6;
  white-space: pre-wrap;
}
</style>
