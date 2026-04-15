const API = window.getApiBase();
const token = localStorage.getItem('token');
let editingId = null;
let articleFilters = {
  status: '',
  search: '',
  categoryId: ''
};
let articleCategories = [];
let imageCache = null;
let articlePagination = {
  page: 1,
  pageSize: 10,
  total: 0
};
let currentImagePage = 1;
const imagePageSize = 20;
let selectedImages = [];
let markdownEditor = null;
let autosaveTimer = null;
let editorDirty = false;
let isSavingArticle = false;
let imagePickerMode = null;
let suppressAutosave = false;
let currentEditorDraftKey = null;

const AUTOSAVE_DELAY = 2000;
const DRAFT_KEY_NEW = 'admin:draft:new';
const SAVE_STATES = {
  idle: { text: '未保存', className: 'info' },
  savingLocal: { text: '正在自动保存到本地...', className: 'info' },
  saveFailed: { text: '自动保存失败', className: 'error' },
  savedRemote: { text: '已保存到服务器', className: 'success' }
};

function getCoverInput() {
  return document.getElementById('editCover');
}

function getCoverPreview() {
  return document.getElementById('coverPreview');
}

function getEditorModal() {
  return document.getElementById('editorModal');
}

function getImagePickerModal() {
  return document.getElementById('imagePickerModal');
}

function getSaveStateElement() {
  return document.getElementById('editorSaveState');
}

function getEditorStatsElement() {
  return document.getElementById('editorStats');
}

function getEditorModeHintElement() {
  return document.getElementById('editorModeHint');
}

function setEditorModeHint(text) {
  const element = getEditorModeHintElement();
  if (element) element.textContent = text;
}

function setEditorSaveState(key, customText = '') {
  const element = getSaveStateElement();
  if (!element) return;

  const preset = SAVE_STATES[key] || SAVE_STATES.idle;
  element.className = `editor-status-chip ${preset.className}`;
  element.textContent = customText || preset.text;
}

function formatTimeLabel(date = new Date()) {
  return date.toLocaleTimeString('zh-CN', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  });
}

function updateEditorStats(content = '') {
  const element = getEditorStatsElement();
  if (!element) return;

  const normalized = String(content || '').replace(/\s+/g, ' ').trim();
  const charCount = normalized ? normalized.length : 0;
  const readingMinutes = Math.max(1, Math.ceil(charCount / 300));
  element.textContent = `${charCount} 字 · 预计 ${readingMinutes} 分钟阅读`;
}

function handleCoverInputChange() {
  renderCoverPreview(getCoverInput()?.value.trim() || '');
  markEditorDirty();
}

async function refreshRandomCover() {
  try {
    await ensureImageCache();
    const randomUrl = pickRandomCoverUrl();
    if (!randomUrl) {
      showFeedback('当前图床没有可用图片，暂时换不了新封面。', 'error');
      return;
    }
    applyCoverValue(randomUrl);
    markEditorDirty();
    showFeedback('已随机换了一张封面，看看这张顺不顺眼 😎', 'success');
  } catch (error) {
    showFeedback(`随机封面失败：${error.message}`, 'error');
  }
}

function pickRandomCoverUrl() {
  const images = Array.isArray(imageCache) ? imageCache.filter((item) => item && item.url) : [];
  if (!images.length) return '';
  const index = Math.floor(Math.random() * images.length);
  return images[index]?.url || '';
}

function applyCoverValue(url = '', options = {}) {
  const { markDirty = false } = options;
  const input = getCoverInput();
  if (input) input.value = url || '';
  renderCoverPreview(url || '');
  if (markDirty) markEditorDirty();
}

function renderCoverPreview(url = '') {
  const preview = getCoverPreview();
  if (!preview) return;

  if (!url) {
    preview.className = 'cover-preview empty';
    preview.innerHTML = `
      <div class="cover-preview-row">
        <div class="cover-preview-thumb cover-preview-thumb-empty">🖼️</div>
        <div class="cover-preview-side">
          <div class="cover-preview-actions">
            <strong>未选择封面</strong>
            <button type="button" class="btn btn-ghost btn-cover-refresh" onclick="refreshRandomCover()">换一张</button>
          </div>
          <p>默认会随机挑一张图，你要是不喜欢，直接改 URL 就行。</p>
        </div>
      </div>
    `;
    return;
  }

  preview.className = 'cover-preview';
  preview.innerHTML = `
    <div class="cover-preview-row">
      <img class="cover-preview-thumb" src="${escapeHtml(url)}" alt="cover preview">
      <div class="cover-preview-side">
        <div class="cover-preview-actions">
          <strong>当前封面</strong>
          <button type="button" class="btn btn-ghost btn-cover-refresh" onclick="refreshRandomCover()">换一张</button>
        </div>
        <p title="${escapeHtml(url)}">${escapeHtml(url)}</p>
      </div>
    </div>
  `;
}

async function ensureImageCache() {
  if (Array.isArray(imageCache)) return imageCache;
  const json = await requestJson(`${API}/images`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  imageCache = Array.isArray(json.data) ? json.data : [];
  return imageCache;
}

if (!token) location.href = '/admin/login.html';

function escapeHtml(value) {
  return String(value ?? '').replace(/[&<>"']/g, (char) => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;'
  }[char]));
}

function setActiveSidebar(link) {
  document.querySelectorAll('.sidebar a').forEach((a) => a.classList.remove('active'));
  if (link) link.classList.add('active');
}

function showPage(page, link) {
  setActiveSidebar(link);
  if (page === 'articles') loadArticlesPage();
  else if (page === 'categories') loadCategoriesPage();
  else if (page === 'images') loadImagesPage();
}

function showFeedback(message, type = 'info') {
  const feedback = document.getElementById('page-feedback');
  if (!feedback) {
    if (type === 'error') alert(message);
    return;
  }

  feedback.className = `feedback-banner ${type}`;
  feedback.style.display = 'flex';
  feedback.innerHTML = `
    <span class="feedback-icon">${type === 'success' ? '✅' : type === 'error' ? '⚠️' : 'ℹ️'}</span>
    <span>${escapeHtml(message)}</span>
  `;
}

function clearFeedback() {
  const feedback = document.getElementById('page-feedback');
  if (feedback) {
    feedback.style.display = 'none';
    feedback.className = 'feedback-banner';
    feedback.textContent = '';
  }
}

async function requestJson(url, options = {}) {
  const response = await fetch(url, options);
  let payload = null;
  try {
    payload = await response.json();
  } catch (error) {
    payload = null;
  }

  if (response.status === 401) {
    localStorage.removeItem('token');
    location.href = '/admin/login.html';
    throw new Error('登录已过期，请重新登录');
  }

  if (!response.ok) {
    throw new Error((payload && (payload.error || payload.message)) || `请求失败（${response.status}）`);
  }

  return payload;
}

function buildArticleQuery() {
  const params = new URLSearchParams();
  if (articleFilters.status) params.set('status', articleFilters.status);
  if (articleFilters.search) params.set('search', articleFilters.search);
  if (articleFilters.categoryId) params.set('category_id', articleFilters.categoryId);
  params.set('page', String(articlePagination.page));
  params.set('page_size', String(articlePagination.pageSize));
  const query = params.toString();
  return query ? `?${query}` : '';
}

async function loadDashboardStats() {
  const articleTotal = document.getElementById('dashboard-article-total');
  const publishedTotal = document.getElementById('dashboard-published-total');
  const categoryImageTotal = document.getElementById('dashboard-category-image-total');
  if (!articleTotal || !publishedTotal || !categoryImageTotal || !token) return;

  try {
    const json = await requestJson(`${API}/dashboard/stats`, {
      headers: { 'Authorization': `Bearer ${token}` }
    });
    const stats = json.data || {};
    articleTotal.textContent = String(stats.article_total ?? 0);
    publishedTotal.textContent = `${stats.published_total ?? 0} / ${stats.draft_total ?? 0}`;
    categoryImageTotal.textContent = `${stats.category_total ?? 0} / ${stats.image_total ?? 0}`;
  } catch (error) {
    articleTotal.textContent = '--';
    publishedTotal.textContent = '-- / --';
    categoryImageTotal.textContent = '-- / --';
    console.warn('loadDashboardStats failed', error);
  }
}

function renderStateCard(message, type = 'info') {
  const icons = { info: '⏳', empty: '🧅', error: '⚠️' };
  return `
    <div class="state-card ${type}">
      <div class="state-icon">${icons[type] || 'ℹ️'}</div>
      <div class="state-message">${escapeHtml(message)}</div>
    </div>
  `;
}

function renderArticlesLoading(message = '正在加载文章列表...') {
  const list = document.getElementById('article-list');
  if (!list) return;
  list.innerHTML = renderStateCard(message, 'info');
}

function syncCategoryFilterOptions() {
  const select = document.getElementById('articleCategoryFilter');
  if (!select) return;

  const currentValue = articleFilters.categoryId || '';
  select.innerHTML = '<option value="">全部分类</option>';
  articleCategories.forEach((category) => {
    select.innerHTML += `<option value="${category.id}">${escapeHtml(category.name)}</option>`;
  });
  select.value = currentValue;
}

function loadArticlesPage() {
  document.getElementById('content').innerHTML = `
    <section class="card section-card">
      <div class="section-head">
        <div>
          <div class="section-kicker">✦ Content desk</div>
          <h3>文章管理</h3>
          <p>支持按状态、分类和关键词筛选，告别在列表里徒手捞针。</p>
        </div>
        <button class="btn btn-primary" onclick="openEditor()">+ 新建文章</button>
      </div>

      <div id="page-feedback" class="feedback-banner" style="display:none;"></div>

      <div class="filter-grid">
        <label class="field compact-field">
          <span>状态筛选</span>
          <select id="articleStatusFilter">
            <option value="">全部状态</option>
            <option value="published">已发布</option>
            <option value="draft">草稿</option>
          </select>
        </label>

        <label class="field compact-field">
          <span>分类筛选</span>
          <select id="articleCategoryFilter">
            <option value="">全部分类</option>
          </select>
        </label>

        <label class="field compact-field grow">
          <span>关键词</span>
          <input id="articleSearchInput" type="text" placeholder="搜索标题、摘要或正文关键词">
        </label>

        <div class="filter-actions">
          <button class="btn btn-primary" onclick="applyArticleFilters()">筛选</button>
          <button class="btn" onclick="resetArticleFilters()">重置</button>
        </div>
      </div>

      <div id="article-list" class="article-list"></div>
      <div id="article-pagination"></div>
    </section>
  `;

  syncCategoryFilterOptions();
  document.getElementById('articleStatusFilter').value = articleFilters.status;
  document.getElementById('articleCategoryFilter').value = articleFilters.categoryId;
  document.getElementById('articleSearchInput').value = articleFilters.search;
  renderArticlesLoading();
  loadArticles();
}

function applyArticleFilters() {
  articleFilters.status = document.getElementById('articleStatusFilter').value;
  articleFilters.categoryId = document.getElementById('articleCategoryFilter').value;
  articleFilters.search = document.getElementById('articleSearchInput').value.trim();
  articlePagination.page = 1;
  clearFeedback();
  renderArticlesLoading('正在应用筛选条件...');
  loadArticles();
}

function resetArticleFilters() {
  articleFilters = { status: '', search: '', categoryId: '' };
  articlePagination.page = 1;
  document.getElementById('articleStatusFilter').value = '';
  document.getElementById('articleCategoryFilter').value = '';
  document.getElementById('articleSearchInput').value = '';
  clearFeedback();
  renderArticlesLoading('正在重置筛选条件...');
  loadArticles();
}

function renderArticlePagination() {
  const container = document.getElementById('article-pagination');
  if (!container) return;

  const totalPages = Math.max(1, Math.ceil((articlePagination.total || 0) / articlePagination.pageSize));
  if ((articlePagination.total || 0) <= articlePagination.pageSize) {
    container.innerHTML = '';
    return;
  }

  container.innerHTML = `
    <div class="pagination article-pagination">
      <button class="btn" onclick="changeArticlePage(${articlePagination.page - 1})" ${articlePagination.page <= 1 ? 'disabled' : ''}>上一页</button>
      <span>第 ${articlePagination.page} / ${totalPages} 页 · 共 ${articlePagination.total} 篇</span>
      <button class="btn" onclick="changeArticlePage(${articlePagination.page + 1})" ${articlePagination.page >= totalPages ? 'disabled' : ''}>下一页</button>
    </div>
  `;
}

function changeArticlePage(page) {
  const totalPages = Math.max(1, Math.ceil((articlePagination.total || 0) / articlePagination.pageSize));
  if (page < 1 || page > totalPages) return;
  articlePagination.page = page;
  renderArticlesLoading('正在切换页码...');
  loadArticles();
}

async function loadArticles() {
  const list = document.getElementById('article-list');
  if (!list) return;

  try {
    if (!articleCategories.length) {
      await loadCategoryOptions(articleFilters.categoryId || '', { quiet: true, syncFilter: true });
    }
    const json = await requestJson(`${API}/articles${buildArticleQuery()}`);
    const articles = Array.isArray(json.data) ? json.data : [];
    articlePagination.total = Number(json.total || 0);
    articlePagination.page = Number(json.page || articlePagination.page || 1);
    articlePagination.pageSize = Number(json.page_size || articlePagination.pageSize || 10);
    list.innerHTML = '';

    if (!articles.length) {
      renderArticlePagination();
      list.innerHTML = renderStateCard('这里暂时没有符合条件的文章，像冰箱里只剩下一根葱。', 'empty');
      return;
    }

    articles.forEach((article) => {
      const div = document.createElement('article');
      div.className = 'article-card';
      const statusText = article.status === 'draft' ? '草稿' : '已发布';
      const statusClass = article.status === 'draft' ? 'draft' : 'published';
      const topBadge = article.is_top ? '<span class="badge top">置顶</span>' : '';
      const categoryText = article.category && article.category.name ? article.category.name : '未分类';
      const tagsArray = article.tags ? String(article.tags).split(',').map((tag) => tag.trim()).filter(Boolean) : [];
      const tagsHtml = tagsArray.length
        ? tagsArray.map((tag) => `<span class="tag-pill"># ${escapeHtml(tag)}</span>`).join('')
        : '<span class="meta-empty">暂无标签</span>';

      div.innerHTML = `
        <div class="article-card-head">
          <div class="article-title-wrap">
            <div class="article-title-row">
              <h4>${escapeHtml(article.title)}</h4>
              <span class="badge ${statusClass}">${statusText}</span>
              ${topBadge}
            </div>
            <p>${escapeHtml(article.summary || '暂无摘要，作者可能正在和灵感打拉扯战。')}</p>
          </div>
          <div class="article-actions">
            <button class="btn btn-primary" onclick="openEditor(${article.id})">编辑</button>
            <button class="btn btn-danger" onclick="deleteArticle(${article.id})">删除</button>
          </div>
        </div>

        <div class="article-meta-grid">
          <div class="meta-card">
            <small>分类</small>
            <strong>${escapeHtml(categoryText)}</strong>
          </div>
          <div class="meta-card">
            <small>创建时间</small>
            <strong>${new Date(article.created_at).toLocaleDateString()}</strong>
          </div>
          <div class="meta-card tags-card">
            <small>标签</small>
            <div class="tag-row">${tagsHtml}</div>
          </div>
        </div>
      `;
      list.appendChild(div);
    });

    renderArticlePagination();
  } catch (error) {
    const pagination = document.getElementById('article-pagination');
    if (pagination) pagination.innerHTML = '';
    list.innerHTML = renderStateCard(`加载文章失败：${error.message}`, 'error');
    showFeedback(`加载文章失败：${error.message}`, 'error');
  }
}

async function deleteArticle(id) {
  if (!confirm('确定删除这篇文章吗？')) return;

  try {
    clearFeedback();
    showFeedback('正在删除文章...', 'info');
    await requestJson(`${API}/articles/${id}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token}` }
    });
    showFeedback('文章删除成功', 'success');
    renderArticlesLoading('正在刷新文章列表...');
    await loadArticles();
  } catch (error) {
    showFeedback(`删除失败：${error.message}`, 'error');
  }
}

async function loadCategoryOptions(selectedId = '', options = {}) {
  const { quiet = false, syncFilter = false } = options;
  const sel = document.getElementById('editCategory');
  if (sel) {
    sel.innerHTML = '<option value="">选择分类</option>';
  }

  try {
    const json = await requestJson(`${API}/categories`);
    articleCategories = Array.isArray(json.data) ? json.data : [];
    if (sel) {
      articleCategories.forEach((c) => {
        sel.innerHTML += `<option value="${c.id}">${escapeHtml(c.name)}</option>`;
      });
      sel.value = selectedId ? String(selectedId) : '';
    }
    if (syncFilter) {
      syncCategoryFilterOptions();
    }
  } catch (error) {
    if (!quiet) {
      showFeedback(`加载分类失败：${error.message}`, 'error');
    }
  }
}

function initMarkdownEditor() {
  const textarea = document.getElementById('editContent');
  if (!textarea || markdownEditor || typeof EasyMDE === 'undefined') return;

  markdownEditor = new EasyMDE({
    element: textarea,
    spellChecker: false,
    autoDownloadFontAwesome: false,
    forceSync: true,
    status: false,
    renderingConfig: {
      singleLineBreaks: false,
      codeSyntaxHighlighting: false
    },
    toolbar: [
      'bold',
      'italic',
      'heading',
      '|',
      'quote',
      'unordered-list',
      'ordered-list',
      '|',
      'code',
      'link',
      {
        name: 'image-library',
        action: () => openImagePicker('markdown'),
        className: 'fa fa-picture-o',
        title: '从图床插入图片'
      },
      '|',
      'preview',
      'side-by-side',
      'fullscreen'
    ]
  });

  markdownEditor.codemirror.on('change', () => {
    updateEditorStats(getEditorContent());
    markEditorDirty();
  });
}

function destroyMarkdownEditor() {
  if (!markdownEditor) return;
  markdownEditor.toTextArea();
  markdownEditor = null;
}

function getEditorContent() {
  if (markdownEditor) return markdownEditor.value();
  return document.getElementById('editContent')?.value || '';
}

function setEditorContent(value) {
  const nextValue = value || '';
  if (markdownEditor) {
    markdownEditor.value(nextValue);
  } else {
    const textarea = document.getElementById('editContent');
    if (textarea) textarea.value = nextValue;
  }
  updateEditorStats(nextValue);
}

function insertMarkdownImage(url, altText = '图片描述') {
  if (!markdownEditor) return;
  const cm = markdownEditor.codemirror;
  const doc = cm.getDoc();
  const text = `![${altText}](${url})`;
  doc.replaceSelection(text);
  cm.focus();
  markEditorDirty();
}

function openImagePicker(mode) {
  imagePickerMode = mode;
  const modal = getImagePickerModal();
  const tip = document.getElementById('imagePickerModeText');
  if (tip) {
    tip.textContent = mode === 'cover'
      ? '当前操作：选择封面图，点一下就会回填到封面输入框。'
      : '当前操作：插入正文配图，点一下就会插入到当前光标位置。';
  }

  if (modal) modal.style.display = 'block';

  ensureImageCache()
    .then(() => renderImagePicker())
    .catch((error) => {
      const grid = document.getElementById('imagePickerGrid');
      if (grid) grid.innerHTML = renderStateCard(`加载图片失败：${error.message}`, 'error');
    });
}

function closeImagePicker() {
  imagePickerMode = null;
  const modal = getImagePickerModal();
  if (modal) modal.style.display = 'none';
}

function renderImagePicker() {
  const grid = document.getElementById('imagePickerGrid');
  if (!grid) return;

  const images = Array.isArray(imageCache) ? imageCache.filter((item) => item && item.url) : [];
  if (!images.length) {
    grid.innerHTML = renderStateCard('图床里还没有图片，先去上传两张，再来让编辑器吃饱。', 'empty');
    return;
  }

  grid.innerHTML = images.map((item) => `
    <button type="button" class="image-picker-item" onclick="handleImagePick('${escapeHtml(item.url).replace(/'/g, '&#39;')}')">
      <img src="${escapeHtml(item.url)}" alt="image option">
      <span>${escapeHtml(item.url)}</span>
    </button>
  `).join('');
}

function decodeHtmlEntities(value) {
  const textarea = document.createElement('textarea');
  textarea.innerHTML = value;
  return textarea.value;
}

function handleImagePick(rawUrl) {
  const url = decodeHtmlEntities(rawUrl);
  if (!url) return;

  if (imagePickerMode === 'cover') {
    applyCoverValue(url, { markDirty: true });
  } else if (imagePickerMode === 'markdown') {
    insertMarkdownImage(url);
  }

  closeImagePicker();
}

function getDraftStorageKey() {
  return editingId ? `admin:draft:article:${editingId}` : DRAFT_KEY_NEW;
}

function buildEditorSnapshot() {
  return {
    title: document.getElementById('editTitle')?.value.trim() || '',
    summary: document.getElementById('editSummary')?.value.trim() || '',
    cover_image: document.getElementById('editCover')?.value.trim() || '',
    category_id: document.getElementById('editCategory')?.value || '',
    tags: document.getElementById('editTags')?.value.trim() || '',
    is_top: !!document.getElementById('editIsTop')?.checked,
    status: document.getElementById('editStatus')?.value || 'draft',
    content: getEditorContent(),
    updated_at: new Date().toISOString(),
    editing_id: editingId || null
  };
}

function saveDraftToLocal() {
  try {
    const key = getDraftStorageKey();
    const snapshot = buildEditorSnapshot();
    localStorage.setItem(key, JSON.stringify(snapshot));
    currentEditorDraftKey = key;
    editorDirty = false;
    setEditorSaveState('idle', `已自动保存到本地 ${formatTimeLabel(new Date(snapshot.updated_at))}`);
  } catch (error) {
    console.warn('saveDraftToLocal failed', error);
    setEditorSaveState('saveFailed');
  }
}

function loadDraftFromLocal() {
  const key = getDraftStorageKey();
  const raw = localStorage.getItem(key);
  if (!raw) return null;

  try {
    const parsed = JSON.parse(raw);
    return parsed && typeof parsed === 'object' ? parsed : null;
  } catch (error) {
    console.warn('loadDraftFromLocal failed', error);
    return null;
  }
}

function clearDraftFromLocal(key = getDraftStorageKey()) {
  if (!key) return;
  localStorage.removeItem(key);
  if (currentEditorDraftKey === key) {
    currentEditorDraftKey = null;
  }
}

function applySnapshotToEditor(snapshot) {
  if (!snapshot) return;

  suppressAutosave = true;
  document.getElementById('editTitle').value = snapshot.title || '';
  document.getElementById('editSummary').value = snapshot.summary || '';
  applyCoverValue(snapshot.cover_image || '');
  document.getElementById('editCategory').value = snapshot.category_id ? String(snapshot.category_id) : '';
  document.getElementById('editTags').value = snapshot.tags || '';
  document.getElementById('editIsTop').checked = !!snapshot.is_top;
  document.getElementById('editStatus').value = snapshot.status || (editingId ? 'published' : 'draft');
  setEditorContent(snapshot.content || '');
  suppressAutosave = false;
}

function queueDraftAutosave() {
  if (suppressAutosave || !getEditorModal() || getEditorModal().style.display !== 'block') return;

  setEditorSaveState('savingLocal');
  if (autosaveTimer) window.clearTimeout(autosaveTimer);
  autosaveTimer = window.setTimeout(() => {
    saveDraftToLocal();
    autosaveTimer = null;
  }, AUTOSAVE_DELAY);
}

function markEditorDirty() {
  if (suppressAutosave) return;
  editorDirty = true;
  setEditorSaveState('idle');
  queueDraftAutosave();
}

function resetEditorFields() {
  suppressAutosave = true;
  document.getElementById('editTitle').value = '';
  document.getElementById('editSummary').value = '';
  document.getElementById('editCover').value = '';
  document.getElementById('editCategory').value = '';
  document.getElementById('editTags').value = '';
  document.getElementById('editIsTop').checked = false;
  document.getElementById('editStatus').value = editingId ? 'published' : 'draft';
  setEditorContent('');
  renderCoverPreview('');
  suppressAutosave = false;
}

function bindEditorFieldListeners() {
  const fieldIds = ['editTitle', 'editSummary', 'editCategory', 'editTags', 'editStatus'];
  fieldIds.forEach((id) => {
    const element = document.getElementById(id);
    if (!element || element.dataset.autosaveBound === 'true') return;
    element.addEventListener('input', markEditorDirty);
    element.addEventListener('change', markEditorDirty);
    element.dataset.autosaveBound = 'true';
  });

  const checkbox = document.getElementById('editIsTop');
  if (checkbox && checkbox.dataset.autosaveBound !== 'true') {
    checkbox.addEventListener('change', markEditorDirty);
    checkbox.dataset.autosaveBound = 'true';
  }
}

async function maybeRestoreDraft() {
  const draft = loadDraftFromLocal();
  if (!draft) return;

  const message = editingId
    ? `检测到本地草稿（${formatTimeLabel(new Date(draft.updated_at || Date.now()))}）。\n恢复本地草稿？选择“取消”将继续使用服务器内容。`
    : `检测到未完成的新文章草稿（${formatTimeLabel(new Date(draft.updated_at || Date.now()))}）。\n要恢复它吗？`;

  const shouldRestore = confirm(message);
  if (shouldRestore) {
    applySnapshotToEditor(draft);
    setEditorSaveState('idle', `已恢复本地草稿 ${formatTimeLabel(new Date(draft.updated_at || Date.now()))}`);
    editorDirty = false;
  }
}

// Active mainline editor lives in index.html modal. Legacy editor.html is reference-only.
async function openEditor(id = null) {
  editingId = id;
  currentEditorDraftKey = getDraftStorageKey();
  clearFeedback();
  setEditorModeHint(id ? `编辑文章 #${id}` : '新建文章');
  document.getElementById('editorTitle').textContent = id ? '编辑文章' : '新建文章';
  getEditorModal().style.display = 'block';

  bindEditorFieldListeners();
  initMarkdownEditor();
  resetEditorFields();
  setEditorSaveState('idle');
  editorDirty = false;

  const preloadResults = await Promise.allSettled([
    loadCategoryOptions(),
    ensureImageCache()
  ]);
  const preloadErrors = preloadResults
    .filter((result) => result.status === 'rejected')
    .map((result) => result.reason?.message || '未知错误');

  if (preloadErrors.length) {
    showFeedback(`部分编辑器资源加载失败：${preloadErrors.join('；')}。你仍然可以继续编辑，本地草稿也照常可用。`, 'error');
  }

  if (!id) {
    const draft = loadDraftFromLocal();
    if (draft) {
      await maybeRestoreDraft();
    } else {
      const randomCover = pickRandomCoverUrl();
      if (randomCover) applyCoverValue(randomCover);
    }
    updateEditorStats(getEditorContent());
    return;
  }

  try {
    showFeedback('正在加载文章详情...', 'info');
    const json = await requestJson(`${API}/articles/${id}`);
    const a = json.data;
    applySnapshotToEditor({
      title: a.title || '',
      summary: a.summary || '',
      cover_image: a.cover_image || '',
      category_id: a.category_id ? String(a.category_id) : '',
      tags: a.tags || '',
      is_top: !!a.is_top,
      status: a.status || 'published',
      content: a.content || ''
    });
    if (!preloadErrors.length) {
      clearFeedback();
    }
    await maybeRestoreDraft();
    updateEditorStats(getEditorContent());
    editorDirty = false;
  } catch (error) {
    showFeedback(`加载文章详情失败：${error.message}`, 'error');
  }
}

function closeEditor() {
  const modal = getEditorModal();
  if (!modal) return;

  if (editorDirty || autosaveTimer) {
    const shouldClose = confirm('当前还有未提交到服务器的改动。\n确定关闭吗？本地草稿会保留，稍后可恢复。');
    if (!shouldClose) return;
  }

  if (autosaveTimer) {
    window.clearTimeout(autosaveTimer);
    autosaveTimer = null;
    if (editorDirty) saveDraftToLocal();
  }

  modal.style.display = 'none';
  closeImagePicker();
  editorDirty = false;
  setEditorSaveState('idle');
}

async function saveArticle() {
  if (isSavingArticle) return;

  const saveButton = document.querySelector('#editorModal .btn.btn-primary');
  const originalText = saveButton ? saveButton.textContent : '保存';
  const draftKeyBeforeSave = getDraftStorageKey();
  const data = {
    title: document.getElementById('editTitle').value.trim(),
    summary: document.getElementById('editSummary').value.trim(),
    cover_image: document.getElementById('editCover').value.trim(),
    category_id: parseInt(document.getElementById('editCategory').value, 10) || 0,
    tags: document.getElementById('editTags').value.trim(),
    is_top: document.getElementById('editIsTop').checked,
    status: document.getElementById('editStatus').value,
    content: getEditorContent()
  };

  if (!data.title) {
    showFeedback('标题不能为空', 'error');
    return;
  }

  const url = editingId ? `${API}/articles/${editingId}` : `${API}/articles`;
  const method = editingId ? 'PUT' : 'POST';

  try {
    clearFeedback();
    isSavingArticle = true;
    if (saveButton) {
      saveButton.disabled = true;
      saveButton.textContent = '保存中...';
    }
    setEditorSaveState('savingLocal', '正在保存到服务器...');
    showFeedback(editingId ? '正在保存文章...' : '正在创建文章...', 'info');
    const response = await requestJson(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(data)
    });

    if (!editingId && response && response.data && response.data.id) {
      editingId = response.data.id;
    }

    if (autosaveTimer) {
      window.clearTimeout(autosaveTimer);
      autosaveTimer = null;
    }
    clearDraftFromLocal(draftKeyBeforeSave);
    clearDraftFromLocal(getDraftStorageKey());
    editorDirty = false;
    setEditorSaveState('savedRemote');
    closeEditor();
    showFeedback(editingId ? '文章更新成功' : '文章创建成功', 'success');
    renderArticlesLoading('正在刷新文章列表...');
    await loadArticles();
  } catch (error) {
    showFeedback(`保存失败：${error.message}`, 'error');
    setEditorSaveState('saveFailed', '保存到服务器失败，本地草稿仍保留');
  } finally {
    isSavingArticle = false;
    if (saveButton) {
      saveButton.disabled = false;
      saveButton.textContent = originalText;
    }
  }
}

function loadCategoriesPage() {
  document.getElementById('content').innerHTML = `
    <section class="card section-card">
      <div class="section-head">
        <div>
          <div class="section-kicker">✦ Taxonomy desk</div>
          <h3>分类管理</h3>
          <p>给内容宇宙分组排班，别让文章在后台里流浪。</p>
        </div>
      </div>

      <div id="page-feedback" class="feedback-banner" style="display:none;"></div>

      <div class="category-toolbar">
        <label class="field compact-field grow">
          <span>新分类名称</span>
          <input type="text" id="newCat" placeholder="例如：K3s、运维、前端设计">
        </label>
        <button class="btn btn-primary" onclick="addCategory()">添加分类</button>
      </div>

      <div id="cat-list" class="category-list"></div>
    </section>
  `;
  loadCategories();
}

async function loadCategories() {
  const list = document.getElementById('cat-list');
  if (!list) return;
  list.innerHTML = renderStateCard('正在加载分类列表...', 'info');

  try {
    const json = await requestJson(`${API}/categories`);
    const categories = Array.isArray(json.data) ? json.data : [];
    list.innerHTML = '';

    if (!categories.length) {
      list.innerHTML = renderStateCard('还没有分类，快先建一个，不然后台像没贴标签的快递站。', 'empty');
      return;
    }

    categories.forEach((cat) => {
      const div = document.createElement('div');
      div.className = 'category-card';
      div.innerHTML = `
        <div class="category-card-main">
          <small>Category</small>
          <strong>${escapeHtml(cat.name)}</strong>
        </div>
        <div class="category-card-actions">
          <button class="btn" onclick="renameCategory(${cat.id}, '${escapeHtml(cat.name).replace(/'/g, '&#39;')}')">重命名</button>
          <button class="btn btn-danger" onclick="deleteCat(${cat.id})">删除</button>
        </div>
      `;
      list.appendChild(div);
    });
  } catch (error) {
    list.innerHTML = renderStateCard(`加载分类失败：${error.message}`, 'error');
    showFeedback(`加载分类失败：${error.message}`, 'error');
  }
}

async function addCategory() {
  const input = document.getElementById('newCat');
  const name = input.value.trim();
  if (!name) {
    showFeedback('请输入分类名称', 'error');
    return;
  }

  try {
    showFeedback('正在添加分类...', 'info');
    await requestJson(`${API}/categories`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ name })
    });
    input.value = '';
    articleCategories = [];
    await loadCategories();
    await loadDashboardStats();
    showFeedback('分类添加成功', 'success');
  } catch (error) {
    showFeedback(`添加分类失败：${error.message}`, 'error');
  }
}

async function renameCategory(id, currentName) {
  const nextName = prompt('请输入新的分类名称：', currentName || '');
  if (nextName === null) return;

  const name = nextName.trim();
  if (!name) {
    showFeedback('分类名称不能为空', 'error');
    return;
  }

  try {
    showFeedback('正在重命名分类...', 'info');
    await requestJson(`${API}/categories/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ name })
    });
    articleCategories = [];
    await loadCategories();
    await loadCategoryOptions(articleFilters.categoryId || '', { quiet: true, syncFilter: true });
    await loadDashboardStats();
    showFeedback('分类重命名成功', 'success');
  } catch (error) {
    showFeedback(`分类重命名失败：${error.message}`, 'error');
  }
}

async function deleteCat(id) {
  if (!confirm('确定删除这个分类吗？')) return;
  try {
    showFeedback('正在删除分类...', 'info');
    await requestJson(`${API}/categories/${id}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token}` }
    });
    articleCategories = [];
    await loadCategories();
    await loadDashboardStats();
    showFeedback('分类删除成功', 'success');
  } catch (error) {
    showFeedback(`删除分类失败：${error.message}`, 'error');
  }
}

function logout() {
  localStorage.removeItem('token');
  location.href = '/admin/login.html';
}

function loadImageHistory() {
  if (imageCache) {
    renderImages();
    return;
  }
  const list = document.getElementById('imageHistory');
  if (list) list.innerHTML = renderStateCard('正在加载图片列表...', 'info');
  fetch(`${API}/images`, { headers: { 'Authorization': `Bearer ${token}` } })
    .then((res) => res.json())
    .then((json) => {
      imageCache = json.data || [];
      renderImages();
    })
    .catch((error) => {
      if (list) list.innerHTML = renderStateCard(`加载图片失败：${error.message}`, 'error');
    });
}

function changePage(page) {
  const total = Math.ceil(imageCache.length / imagePageSize);
  if (page < 1 || page > total) return;
  currentImagePage = page;
  renderImages();
}

function loadImagesPage() {
  document.getElementById('content').innerHTML = `
    <section class="card section-card">
      <div class="section-head">
        <div>
          <div class="section-kicker">✦ Media desk</div>
          <h3>图床管理</h3>
          <p>支持拖拽上传、多图管理和批量删除，别让图片在仓库里野蛮生长。</p>
        </div>
      </div>

      <div id="uploadArea" class="upload-dropzone">
        <div class="upload-emoji">🖼️</div>
        <h4>点击或拖拽图片到这里上传</h4>
        <p>支持多图上传，传完就能复制链接去写文章。</p>
        <input type="file" id="imgFile" accept="image/*" multiple style="display:none">
      </div>

      <div id="uploadResult"></div>

      <div class="images-headline">
        <h4>图片列表</h4>
        <p>支持选中后批量删除，也可以逐个复制链接。</p>
      </div>

      <div id="imageHistory"></div>
    </section>
  `;
  const area = document.getElementById('uploadArea');
  const input = document.getElementById('imgFile');
  area.onclick = () => input.click();
  area.ondragover = (e) => { e.preventDefault(); area.classList.add('dragover'); };
  area.ondragleave = () => { area.classList.remove('dragover'); };
  area.ondrop = (e) => {
    e.preventDefault();
    area.classList.remove('dragover');
    if (e.dataTransfer.files.length) uploadMultiple(e.dataTransfer.files);
  };
  input.onchange = (e) => { if (e.target.files.length) uploadMultiple(e.target.files); };
  loadImageHistory();
}

function uploadMultiple(files) {
  const result = document.getElementById('uploadResult');
  result.innerHTML = `
    <div class="upload-status-card">
      <strong>上传进行中</strong>
      <p>正在处理 0/${files.length} 张图片...</p>
    </div>
  `;

  let completed = 0;
  const promises = Array.from(files).map((file) => {
    const formData = new FormData();
    formData.append('image', file);
    return fetch(`${API}/upload`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${token}` },
      body: formData
    }).then(() => {
      completed++;
      result.innerHTML = `
        <div class="upload-status-card">
          <strong>上传进行中</strong>
          <p>正在处理 ${completed}/${files.length} 张图片...</p>
        </div>
      `;
    });
  });

  Promise.all(promises).then(() => {
    result.innerHTML = `
      <div class="upload-status-card success">
        <strong>上传完成</strong>
        <p>✅ 已成功上传 ${files.length} 张图片，现在可以去文章里贴图了。</p>
      </div>
    `;
    imageCache = null;
    loadImageHistory();
  });
}

function toggleSelect(key) {
  const idx = selectedImages.indexOf(key);
  if (idx > -1) {
    selectedImages.splice(idx, 1);
  } else {
    selectedImages.push(key);
  }
}

function deleteSelected() {
  if (selectedImages.length === 0) return alert('请选择要删除的图片');
  if (!confirm(`确定删除 ${selectedImages.length} 张图片？`)) return;

  Promise.all(selectedImages.map((key) =>
    fetch(`${API}/images/${key}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token}` }
    })
  )).then(() => {
    imageCache = null;
    selectedImages = [];
    loadImageHistory();
  });
}

function renderImages() {
  const list = document.getElementById('imageHistory');
  if (!list) return;

  const allImages = Array.isArray(imageCache) ? imageCache : [];
  const total = Math.ceil(allImages.length / imagePageSize) || 1;
  if (currentImagePage > total) currentImagePage = total;

  if (!allImages.length) {
    list.innerHTML = renderStateCard('图片库还是空的，连一张表情包都没有。', 'empty');
    return;
  }

  const start = (currentImagePage - 1) * imagePageSize;
  const end = start + imagePageSize;
  const pageData = allImages.slice(start, end);

  list.innerHTML = `
    <div class="images-toolbar">
      <button class="btn btn-danger" onclick="deleteSelected()">删除选中</button>
      <span>已选择 ${selectedImages.length} 张 / 共 ${allImages.length} 张</span>
    </div>
  `;

  const grid = document.createElement('div');
  grid.className = 'image-grid';

  pageData.forEach((item) => {
    const div = document.createElement('div');
    div.className = 'image-card';

    const checked = selectedImages.includes(item.key);
    div.innerHTML = `
      <label class="image-select">
        <input type="checkbox" ${checked ? 'checked' : ''}>
      </label>
      <img src="${escapeHtml(item.url)}" alt="uploaded image">
      <div class="image-card-body">
        <div class="image-url">${escapeHtml(item.url)}</div>
        <button class="btn btn-primary">复制链接</button>
      </div>
    `;

    const checkbox = div.querySelector('input[type="checkbox"]');
    checkbox.onchange = () => toggleSelect(item.key);
    div.querySelector('button').onclick = () => copyUrl(item.url);
    grid.appendChild(div);
  });

  list.appendChild(grid);

  if (total > 1) {
    const pagination = document.createElement('div');
    pagination.className = 'pagination';
    pagination.innerHTML = `
      <button class="btn" onclick="changePage(${currentImagePage - 1})" ${currentImagePage === 1 ? 'disabled' : ''}>上一页</button>
      <span>第 ${currentImagePage} / ${total} 页</span>
      <button class="btn" onclick="changePage(${currentImagePage + 1})" ${currentImagePage === total ? 'disabled' : ''}>下一页</button>
    `;
    list.appendChild(pagination);
  }
}

function copyUrl(url) {
  const textarea = document.createElement('textarea');
  textarea.value = url;
  textarea.style.position = 'fixed';
  textarea.style.opacity = '0';
  document.body.appendChild(textarea);
  textarea.select();
  try {
    document.execCommand('copy');
    showFeedback('链接已复制，可以快乐贴图了 ✨', 'success');
  } catch (err) {
    showFeedback('复制失败，请手动复制链接', 'error');
  }
  document.body.removeChild(textarea);
}

document.addEventListener('keydown', (event) => {
  const modal = getEditorModal();
  if (!modal || modal.style.display !== 'block') return;
  if (!(event.ctrlKey || event.metaKey) || event.key.toLowerCase() !== 's') return;

  event.preventDefault();
  saveArticle();
});

window.addEventListener('beforeunload', (event) => {
  if (!(editorDirty || autosaveTimer)) return;
  event.preventDefault();
  event.returnValue = '';
});

loadDashboardStats();
loadArticlesPage();
