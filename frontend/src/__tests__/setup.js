// setup.js は Vitest 実行時のブラウザAPI補完を行う。
// Vuetify が参照する visualViewport を定義する。
if (!globalThis.visualViewport) {
  globalThis.visualViewport = {
    width: 0,
    height: 0,
    scale: 1,
    addEventListener: () => {},
    removeEventListener: () => {}
  }
}
