<script setup>
// ContractorPasswordDialog は Contractor パスワード入力のダイアログを担う。
// UI 描画は Vuetify に委ねる。
import { computed, ref } from 'vue'

import { Quit } from '../../wailsjs/runtime/runtime.js'
import { useAppStore } from '../stores/app'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: true,
  },
})

const emit = defineEmits(['update:modelValue', 'verified', 'closed'])

const appStore = useAppStore()

const password = ref('')
const errorMessage = ref('')
const failed = ref(false)

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

const isBusy = computed(() => appStore.isBusy)

async function handleVerify() {
  // 検証が失敗した場合はメッセージ表示後に閉じる動線を有効化する。
  errorMessage.value = ''
  const result = await appStore.verifyContractorPassword(password.value)
  if (!result) {
    errorMessage.value = '認証に失敗しました。'
    failed.value = true
    return
  }
  emit('verified', result)
  isOpen.value = false
}

function handleClose() {
  // 失敗後のクローズはアプリ終了に接続する。
  Quit()
  emit('closed')
}
</script>

<template>
  <v-dialog v-model="isOpen" persistent max-width="520">
    <v-card rounded="lg">
      <v-card-title class="text-h6">Contractor 認証</v-card-title>
      <v-card-text>
        <v-alert v-if="errorMessage" type="error" variant="tonal" class="mb-4">
          {{ errorMessage }}
        </v-alert>
        <v-text-field
          v-model="password"
          label="パスワード"
          type="password"
          variant="outlined"
          density="comfortable"
          :disabled="isBusy || failed"
        />
      </v-card-text>
      <v-card-actions class="justify-end">
        <v-btn
          data-testid="close"
          variant="text"
          color="secondary"
          :disabled="!failed"
          @click="handleClose"
        >
          閉じる
        </v-btn>
        <v-btn
          data-testid="verify"
          variant="flat"
          color="primary"
          :loading="isBusy"
          :disabled="failed"
          @click="handleVerify"
        >
          認証
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
