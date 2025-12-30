// app.js はアプリ共通状態と起動処理のアクションを担い、UIの描画は扱わない。
// Wails API 呼び出しは utils のラッパーを経由する。
import { defineStore } from 'pinia'

import {
  createProjectRoot,
  detectMode,
  getAppBootstrap,
  saveLastProjectRoot,
  validateProjectRoot,
  verifyContractorPassword
} from '../utils/apiClient'
import { useErrorsStore } from './errors'

// useAppStore は DD-STORE-005/012 のアプリ共通ストアを提供する。
// 目的: モード・プロジェクトルート・起動状態を管理する。
// 入力: Pinia の内部状態。
// 出力: app ストア。
// エラー: なし。
// 副作用: なし。
// 並行性: Pinia の更新に従う。
// 不変条件: mode は "Vendor"/"Contractor" のいずれか。
// 関連DD: DD-STORE-005, DD-STORE-012
export const useAppStore = defineStore('app', {
  state: () => ({
    mode: 'Vendor',
    projectRoot: null,
    lastProjectRootPath: null,
    pageSize: 20,
    bootstrapLoaded: false,
    contractorAuthRequired: false,
    isBusy: false
  }),
  actions: {
    // bootstrap は起動時情報を取得して状態へ反映する。
    // 目的: 初期表示に必要な設定値を読み込む。
    // 入力: なし。
    // 出力: なし。
    // エラー: 取得失敗時は errors ストアに登録する。
    // 副作用: バックエンド呼び出しを行う。
    // 並行性: 同時実行は想定しない。
    // 不変条件: 成功時に bootstrapLoaded を true にする。
    // 関連DD: DD-STORE-012
    async bootstrap() {
      const errors = useErrorsStore()
      this.isBusy = true
      try {
        const data = await getAppBootstrap()
        this.pageSize = data.ui_page_size ?? this.pageSize
        this.lastProjectRootPath = data.last_project_root_path ?? null
        this.contractorAuthRequired = data.has_contractor_auth_file ?? false
        this.bootstrapLoaded = true
      } catch (e) {
        errors.capture(e, { source: 'app', action: 'bootstrap' })
      } finally {
        this.isBusy = false
      }
    },
    // selectProjectRoot は既存パスを検証し、設定を保存する。
    // 目的: 選択したプロジェクトルートを確定する。
    // 入力: path は選択パス。
    // 出力: 検証結果 DTO。
    // エラー: 失敗時は errors ストアに登録する。
    // 副作用: バックエンド呼び出しを行う。
    // 並行性: 同時実行は想定しない。
    // 不変条件: 成功時に projectRoot と lastProjectRootPath を更新する。
    // 関連DD: DD-STORE-012
    async selectProjectRoot(path) {
      const errors = useErrorsStore()
      this.isBusy = true
      try {
        const result = await validateProjectRoot(path)
        if (result.is_valid) {
          await saveLastProjectRoot(result.normalized_path ?? path)
          this.projectRoot = result.normalized_path ?? path
          this.lastProjectRootPath = this.projectRoot
        }
        return result
      } catch (e) {
        errors.capture(e, { source: 'app', action: 'selectProjectRoot' })
        return null
      } finally {
        this.isBusy = false
      }
    },
    // createProjectRoot は新規作成後に設定を保存する。
    // 目的: 新規プロジェクトルートを作成して選択状態にする。
    // 入力: path は作成パス。
    // 出力: 検証結果 DTO。
    // エラー: 失敗時は errors ストアに登録する。
    // 副作用: バックエンド呼び出しを行う。
    // 並行性: 同時実行は想定しない。
    // 不変条件: 成功時に projectRoot と lastProjectRootPath を更新する。
    // 関連DD: DD-STORE-012
    async createProjectRoot(path) {
      const errors = useErrorsStore()
      this.isBusy = true
      try {
        const result = await createProjectRoot(path)
        if (result.is_valid) {
          await saveLastProjectRoot(result.normalized_path ?? path)
          this.projectRoot = result.normalized_path ?? path
          this.lastProjectRootPath = this.projectRoot
        }
        return result
      } catch (e) {
        errors.capture(e, { source: 'app', action: 'createProjectRoot' })
        return null
      } finally {
        this.isBusy = false
      }
    },
    // detectMode は起動時モード判定を行う。
    // 目的: Contractor パスワード要求の有無とモードを取得する。
    // 入力: なし。
    // 出力: ModeDTO。
    // エラー: 失敗時は errors ストアに登録する。
    // 副作用: バックエンド呼び出しを行う。
    // 並行性: 同時実行は想定しない。
    // 不変条件: contractorAuthRequired と mode を更新する。
    // 関連DD: DD-STORE-012
    async detectMode() {
      const errors = useErrorsStore()
      this.isBusy = true
      try {
        const result = await detectMode()
        this.mode = result.mode
        this.contractorAuthRequired = result.requires_password ?? false
        return result
      } catch (e) {
        errors.capture(e, { source: 'app', action: 'detectMode' })
        return null
      } finally {
        this.isBusy = false
      }
    },
    // verifyContractorPassword は Contractor パスワードを検証する。
    // 目的: Contractor モードへの移行を確定する。
    // 入力: password は入力パスワード。
    // 出力: ModeDTO。
    // エラー: 失敗時は errors ストアに登録する。
    // 副作用: バックエンド呼び出しを行う。
    // 並行性: 同時実行は想定しない。
    // 不変条件: 成功時に mode を更新する。
    // 関連DD: DD-STORE-012
    async verifyContractorPassword(password) {
      const errors = useErrorsStore()
      this.isBusy = true
      try {
        const result = await verifyContractorPassword(password)
        this.mode = result.mode
        this.contractorAuthRequired = result.requires_password ?? false
        return result
      } catch (e) {
        errors.capture(e, { source: 'app', action: 'verifyContractorPassword' })
        return null
      } finally {
        this.isBusy = false
      }
    }
  }
})
