(function () {
  const DEFAULT_API_BASE = 'http://172.28.74.191:31083/api';

  function normalizeApiBase(value) {
    if (!value || typeof value !== 'string') return DEFAULT_API_BASE;
    return value.replace(/\/$/, '');
  }

  function getApiBase() {
    const candidates = [
      window.APP_CONFIG && window.APP_CONFIG.apiBase,
      window.API_BASE,
      document.documentElement && document.documentElement.dataset && document.documentElement.dataset.apiBase,
      window.localStorage && window.localStorage.getItem('api_base')
    ];

    const resolved = candidates.find((item) => typeof item === 'string' && item.trim());
    const apiBase = normalizeApiBase(resolved || DEFAULT_API_BASE);

    if (!resolved) {
      console.warn('[config] 未检测到自定义 API_BASE，已使用默认值:', apiBase);
    }

    window.API_BASE = apiBase;
    window.APP_CONFIG = Object.assign({}, window.APP_CONFIG, {
      apiBase,
      getApiBase: () => apiBase
    });
    return apiBase;
  }

  window.getApiBase = getApiBase;
  getApiBase();
})();
