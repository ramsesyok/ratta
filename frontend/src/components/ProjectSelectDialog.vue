<script setup>
// ProjectSelectDialog はプロジェクトルート選択ダイアログの挙動を担う。
// UI 描画は Vuetify に委ねる。
import { computed, onMounted, ref } from 'vue'

import { Quit } from '../../wailsjs/runtime/runtime.js'
import { useAppStore } from '../stores/app'
import { useErrorsStore } from '../stores/errors'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: true,
  },
})

const emit = defineEmits(['update:modelValue', 'selected', 'cancel'])

const appStore = useAppStore()
const errorsStore = useErrorsStore()

const pathInput = ref('')
const errorMessage = ref('')

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

const isBusy = computed(() => appStore.isBusy)

onMounted(async () => {
  // 起動時情報を読み込み、直前のルートを入力欄へ反映する。
  if (!appStore.bootstrapLoaded) {
    await appStore.bootstrap()
  }
  pathInput.value = appStore.lastProjectRootPath ?? ''
})

async function handleValidate() {
  errorMessage.value = ''
  const result = await appStore.selectProjectRoot(pathInput.value)
  if (!result) {
    errorMessage.value = '処理に失敗しました。'
    return
  }
  if (!result.is_valid) {
    errorMessage.value = result.message ?? 'パスが無効です。'
    return
  }
  emit('selected', result.normalized_path ?? pathInput.value)
  isOpen.value = false
}

async function handleCreate() {
  errorMessage.value = ''
  const result = await appStore.createProjectRoot(pathInput.value)
  if (!result) {
    errorMessage.value = '処理に失敗しました。'
    return
  }
  if (!result.is_valid) {
    errorMessage.value = result.message ?? 'パスが無効です。'
    return
  }
  emit('selected', result.normalized_path ?? pathInput.value)
  isOpen.value = false
}

function handleCancel() {
  // キャンセル時はアプリ終了を通知する。
  errorsStore.markAllRead()
  Quit()
  emit('cancel')
}
</script>

<template>
  <v-dialog v-model="isOpen" persistent max-width="640">
    <v-card rounded="lg">
      <v-card-title class="text-h6">プロジェクトを選択</v-card-title>
      <v-card-text>
        <v-alert v-if="errorMessage" type="error" variant="tonal" class="mb-4">
          {{ errorMessage }}
        </v-alert>
        <v-text-field
          v-model="pathInput"
          label="プロジェクトルート"
          variant="outlined"
          density="comfortable"
          :disabled="isBusy"
        />
      </v-card-text>
      <v-card-actions class="justify-end">
        <v-btn
          data-testid="cancel"
          variant="text"
          color="secondary"
          :disabled="isBusy"
          @click="handleCancel"
        >
          キャンセル
        </v-btn>
        <v-btn
          data-testid="create"
          variant="tonal"
          color="primary"
          :loading="isBusy"
          @click="handleCreate"
        >
          新規作成
        </v-btn>
        <v-btn
          data-testid="validate"
          variant="flat"
          color="primary"
          :loading="isBusy"
          @click="handleValidate"
        >
          開く
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
