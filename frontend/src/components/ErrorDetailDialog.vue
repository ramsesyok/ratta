<script setup>
// ErrorDetailDialog はエラー一覧と詳細情報の表示を担当する。
// エラー取得はストアに委ね、UIでフィルタとコピー操作のみ扱う。
import { computed, ref } from 'vue'

import { useCategoriesStore } from '../stores/categories'
import { useErrorsStore } from '../stores/errors'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['update:modelValue'])

const errorsStore = useErrorsStore()
const categoriesStore = useCategoriesStore()

const scope = ref('all')

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const selectedCategory = computed(() => categoriesStore.selectedCategory)

const filteredItems = computed(() => {
  if (scope.value !== 'category' || !selectedCategory.value) {
    return errorsStore.items
  }
  return errorsStore.items.filter((item) => item.category === selectedCategory.value)
})

// copyText は指定文字列をクリップボードへコピーする。
// 目的: エラーメッセージやパスを簡単に共有できるようにする。
// 入力: value はコピー対象の文字列。
// 出力: なし。
// エラー: クリップボード未対応時は何もしない。
// 副作用: クリップボードを書き換える。
// 並行性: 単一UIイベント前提。
// 不変条件: 空文字はコピーしない。
// 関連DD: DD-UI-007
async function copyText(value) {
  if (!value) {
    return
  }
  if (!navigator?.clipboard?.writeText) {
    return
  }
  await navigator.clipboard.writeText(value)
}

// resolveTargetPath は表示用の対象パス文字列を返す。
// 目的: ApiError の有無に関わらず表示値を統一する。
// 入力: item は UiErrorItem。
// 出力: 対象パスの文字列。
// エラー: なし。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: 値が無い場合は "未設定" を返す。
// 関連DD: DD-UI-007
function resolveTargetPath(item) {
  return item.api?.target_path || '未設定'
}
</script>

<template>
  <v-dialog v-model="isOpen" max-width="900">
    <v-card rounded="lg">
      <v-card-title class="text-h6">エラー詳細</v-card-title>
      <v-card-text>
        <div class="d-flex align-center justify-space-between mb-4">
          <v-btn-toggle v-model="scope" density="comfortable" class="mr-4" mandatory>
            <v-btn data-testid="scope-all" value="all">全件</v-btn>
            <v-btn
              data-testid="scope-category"
              value="category"
              :disabled="!selectedCategory"
            >
              カテゴリ
            </v-btn>
          </v-btn-toggle>
          <span class="text-caption">
            選択カテゴリ: {{ selectedCategory || '未選択' }}
          </span>
        </div>

        <v-alert
          v-if="filteredItems.length === 0"
          type="info"
          variant="tonal"
        >
          表示対象のエラーはありません。
        </v-alert>

        <v-list v-else density="compact">
          <v-list-item
            v-for="item in filteredItems"
            :key="item.id"
            data-testid="error-item"
          >
            <v-list-item-title class="text-body-1">
              メッセージ: {{ item.user_message }}
              <v-btn
                data-testid="copy-message"
                size="small"
                variant="text"
                @click="copyText(item.user_message)"
              >
                コピー
              </v-btn>
            </v-list-item-title>
            <v-list-item-subtitle>
              コード: {{ item.api?.error_code || 'unknown' }}
            </v-list-item-subtitle>
            <v-list-item-subtitle>
              対象パス: {{ resolveTargetPath(item) }}
              <v-btn
                data-testid="copy-path"
                size="small"
                variant="text"
                @click="copyText(item.api?.target_path || '')"
              >
                コピー
              </v-btn>
            </v-list-item-subtitle>
            <v-list-item-subtitle>
              発生元: {{ item.source }} / {{ item.action }} / {{ item.occurred_at }}
            </v-list-item-subtitle>
            <v-expansion-panels v-if="item.api?.detail" variant="accordion">
              <v-expansion-panel>
                <v-expansion-panel-title>詳細</v-expansion-panel-title>
                <v-expansion-panel-text>
                  {{ item.api.detail }}
                </v-expansion-panel-text>
              </v-expansion-panel>
            </v-expansion-panels>
          </v-list-item>
        </v-list>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>
