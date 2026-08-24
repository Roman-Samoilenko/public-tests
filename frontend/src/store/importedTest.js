import { ref } from 'vue'

// Реактивное состояние
const data = ref(null) // хранит импортированный тест

export function useImportedTestStore() {
  const setImportedTest = (test) => {
    data.value = test
  }

  const clear = () => {
    data.value = null
  }

  return {
    data,      // реактивная ссылка
    setImportedTest,
    clear
  }
}