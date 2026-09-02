<template>
  <div class="question">
    <div class="q-header">
      <span class="q-num font-mono">{{ index + 1 }}</span>
      <div>
        <p class="q-text">{{ question.text }}</p>
        <span v-if="question.required" class="q-required font-mono">обязательный</span>
      </div>
    </div>

    <!-- Одиночный выбор -->
    <div v-if="question.type === 'single_choice'" class="options">
      <label
        v-for="opt in question.options" :key="opt.id"
        :class="['option', modelValue === opt.id && 'selected']"
      >
        <input type="radio" :name="`q${question.id}`" :value="opt.id"
          :checked="modelValue === opt.id"
          @change="$emit('update:modelValue', opt.id)"
        />
        <span class="option-mark"></span>
        <span class="option-text">{{ opt.text }}</span>
      </label>
    </div>

    <!-- Множественный выбор -->
    <div v-else-if="question.type === 'multiple_choice'" class="options">
      <label
        v-for="opt in question.options" :key="opt.id"
        :class="['option', 'checkbox', (modelValue || []).includes(opt.id) && 'selected']"
      >
        <input type="checkbox" :value="opt.id"
          :checked="(modelValue || []).includes(opt.id)"
          @change="toggleMulti(opt.id)"
        />
        <span class="option-mark checkbox-mark"></span>
        <span class="option-text">{{ opt.text }}</span>
      </label>
    </div>

    <!-- Шкала -->
    <div v-else-if="question.type === 'scale'" class="scale-wrap">
      <div class="scale-labels">
        <span class="scale-label">{{ question.min_label || question.min }}</span>
        <span class="scale-label">{{ question.max_label || question.max }}</span>
      </div>
      <div class="scale-options">
        <label
          v-for="n in scaleRange" :key="n"
          :class="['scale-opt', modelValue === n && 'selected']"
        >
          <input type="radio" :name="`q${question.id}`" :value="n"
            :checked="modelValue === n"
            @change="$emit('update:modelValue', n)"
          />
          <span class="scale-num">{{ n }}</span>
        </label>
      </div>
    </div>

    <!-- Текст -->
    <div v-else-if="question.type === 'text'">
      <textarea
        :value="modelValue || ''"
        @input="$emit('update:modelValue', $event.target.value)"
        rows="3"
        placeholder="Ваш ответ..."
      ></textarea>
    </div>

    <!-- Векторная шкала (grid) -->
    <div v-else-if="question.type === 'vector_scale'" class="grid-wrap">
      <div class="grid-table">
        <div class="grid-header">
          <div class="grid-row-label"></div>
          <div v-for="col in question.cols" :key="col" class="grid-col-header">{{ col }}</div>
        </div>
        <div v-for="row in question.rows" :key="row" class="grid-row">
          <div class="grid-row-label">{{ row }}</div>
          <div v-for="col in question.cols" :key="col" class="grid-cell">
  <label :class="['scale-opt small', isGridSelected(row, col) && 'selected']">
    <input
      :type="question.grid_multiple ? 'checkbox' : 'radio'"
      :name="`q${question.id}_${row}`"
      :value="col"
      :checked="isGridSelected(row, col)"
      @change="updateGrid(row, col)"
    />
    <span class="scale-num">·</span>
  </label>
</div>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  question:   { type: Object, required: true },
  index:      { type: Number, required: true },
  modelValue: { default: null },
})
const emit = defineEmits(['update:modelValue'])

const scaleRange = computed(() => {
  const min = props.question.min ?? 1
  const max = props.question.max ?? 5
  return Array.from({ length: max - min + 1 }, (_, i) => min + i)
})

function toggleMulti(id) {
  const cur = Array.isArray(props.modelValue) ? [...props.modelValue] : []
  const idx = cur.indexOf(id)
  if (idx === -1) cur.push(id)
  else cur.splice(idx, 1)
  emit('update:modelValue', cur)
}
function isGridSelected(row, col) {
  const val = (props.modelValue || {})[row]
  return props.question.grid_multiple
    ? Array.isArray(val) && val.includes(col)
    : val === col
}

function updateGrid(row, col) {
  const cur = props.modelValue && typeof props.modelValue === 'object' ? { ...props.modelValue } : {}
  if (props.question.grid_multiple) {
    const arr = Array.isArray(cur[row]) ? [...cur[row]] : []
    const idx = arr.indexOf(col)
    idx === -1 ? arr.push(col) : arr.splice(idx, 1)
    cur[row] = arr
  } else {
    cur[row] = col
  }
  emit('update:modelValue', cur)
}
</script>

<style scoped>
.question {
  padding: 1.5rem 0;
  border-bottom: 1px solid var(--border);
}
.question:last-child { border-bottom: none; }

.q-header {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.2rem;
}
.q-num {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  color: var(--accent);
  min-width: 1.5rem;
  padding-top: 0.1rem;
}
.q-text {
  font-size: 1rem;
  line-height: 1.5;
  font-weight: 400;
}
.q-required {
  font-size: 0.68rem;
  color: var(--danger);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

/* Radio / Checkbox опции */
.options { display: flex; flex-direction: column; gap: 0.5rem; padding-left: 2.5rem; }
.option {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  padding: 0.6rem 0.9rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  cursor: pointer;
  transition: border-color var(--transition), background var(--transition);
}
.option input { display: none; }
.option:hover { border-color: rgba(212,168,67,0.3); }
.option.selected { border-color: var(--accent); background: var(--accent-dim); }

.option-mark {
  width: 14px; height: 14px;
  border: 2px solid var(--text-muted);
  border-radius: 50%;
  flex-shrink: 0;
  transition: border-color var(--transition), background var(--transition);
}
.option.selected .option-mark {
  border-color: var(--accent);
  background: var(--accent);
}
.checkbox-mark { border-radius: 3px; }
.option-text { font-size: 0.92rem; }

/* Шкала */
.scale-wrap { padding-left: 2.5rem; }
.scale-labels {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}
.scale-label {
  font-family: var(--font-mono);
  font-size: 0.72rem;
  color: var(--text-muted);
}
.scale-options {
  display: flex;
  gap: 0.4rem;
  flex-wrap: wrap;
}
.scale-opt {
  display: flex;
  flex-direction: column;
  align-items: center;
  cursor: pointer;
}
.scale-opt input { display: none; }
.scale-num {
  width: 38px; height: 38px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: var(--font-mono);
  font-size: 0.85rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  transition: all var(--transition);
  color: var(--text-muted);
}
.scale-opt:hover .scale-num { border-color: rgba(212,168,67,0.4); color: var(--text); }
.scale-opt.selected .scale-num { border-color: var(--accent); background: var(--accent-dim); color: var(--accent); }
.scale-opt.small .scale-num { width: 28px; height: 28px; font-size: 1rem; }

/* Grid (vector_scale) */
.grid-wrap { padding-left: 2.5rem; overflow-x: auto; }
.grid-table { min-width: 300px; }
.grid-header, .grid-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.3rem 0;
}
.grid-header { border-bottom: 1px solid var(--border); margin-bottom: 0.3rem; }
.grid-row-label {
  width: 140px;
  flex-shrink: 0;
  font-size: 0.88rem;
  color: var(--text-muted);
}
.grid-col-header {
  flex: 1;
  font-family: var(--font-mono);
  font-size: 0.7rem;
  text-align: center;
  color: var(--text-muted);
  letter-spacing: 0.04em;
}
.grid-cell {
  flex: 1;
  display: flex;
  justify-content: center;
}
</style>
