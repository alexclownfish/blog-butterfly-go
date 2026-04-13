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
let currentPage = 1;
const pageSize = 20;
let selectedImages = [];

function getCoverInput() {
  return document.getElementById('editCover');
}

function getCoverPreview() {
  return document.getElementById('coverPreview');
}

function handleCoverInputChange() {
  renderCoverPreview(getCoverInput()?.value.trim() || '');
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

function applyCoverValue(url = '') {
  const input = getCoverInput();
  if (input) input.value = url || '';
  renderCoverPreview(url || '');
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
  const query = params.toString();
  return query ? `?${query}` : '';
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
  clearFeedback();
  renderArticlesLoading('正在应用筛选条件...');
  loadArticles();
}

function resetArticleFilters() {
  articleFilters = { status: '', search: '', categoryId: '' };
  document.getElementById('articleStatusFilter').value = '';
  document.getElementById('articleCategoryFilter').value = '';
  document.getElementById('articleSearchInput').value = '';
  clearFeedback();
  renderArticlesLoading('正在重置筛选条件...');
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
    list.innerHTML = '';

    if (!articles.length) {
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
  } catch (error) {
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

async function openEditor(id = null) {
  editingId = id;
  clearFeedback();
  document.getElementById('editorTitle').textContent = id ? '编辑文章' : '新建文章';
  document.getElementById('editorModal').style.display = 'block';

  document.getElementById('editTitle').value = '';
  document.getElementById('editSummary').value = '';
  document.getElementById('editCover').value = '';
  document.getElementById('editCategory').value = '';
  document.getElementById('editTags').value = '';
  document.getElementById('editIsTop').checked = false;
  document.getElementById('editStatus').value = id ? 'published' : 'draft';
  document.getElementById('editContent').value = '';
  renderCoverPreview('');

  await Promise.all([
    loadCategoryOptions(),
    ensureImageCache()
  ]);

  if (!id) {
    applyCoverValue(pickRandomCoverUrl());
    return;
  }

  try {
    showFeedback('正在加载文章详情...', 'info');
    const json = await requestJson(`${API}/articles/${id}`);
    const a = json.data;
    document.getElementById('editTitle').value = a.title || '';
    document.getElementById('editSummary').value = a.summary || '';
    applyCoverValue(a.cover_image || '');
    document.getElementById('editCategory').value = a.category_id ? String(a.category_id) : '';
    document.getElementById('editTags').value = a.tags || '';
    document.getElementById('editIsTop').checked = !!a.is_top;
    document.getElementById('editStatus').value = a.status || 'published';
    document.getElementById('editContent').value = a.content || '';
    clearFeedback();
  } catch (error) {
    showFeedback(`加载文章详情失败：${error.message}`, 'error');
  }
}

function closeEditor() {
  document.getElementById('editorModal').style.display = 'none';
}

async function saveArticle() {
  const saveButton = document.querySelector('#editorModal .btn.btn-primary');
  const originalText = saveButton ? saveButton.textContent : '保存';
  const data = {
    title: document.getElementById('editTitle').value.trim(),
    summary: document.getElementById('editSummary').value.trim(),
    cover_image: document.getElementById('editCover').value.trim(),
    category_id: parseInt(document.getElementById('editCategory').value, 10) || 0,
    tags: document.getElementById('editTags').value.trim(),
    is_top: document.getElementById('editIsTop').checked,
    status: document.getElementById('editStatus').value,
    content: document.getElementById('editContent').value
  };

  if (!data.title) {
    showFeedback('标题不能为空', 'error');
    return;
  }

  const url = editingId ? `${API}/articles/${editingId}` : `${API}/articles`;
  const method = editingId ? 'PUT' : 'POST';

  try {
    clearFeedback();
    if (saveButton) {
      saveButton.disabled = true;
      saveButton.textContent = '保存中...';
    }
    showFeedback(editingId ? '正在保存文章...' : '正在创建文章...', 'info');
    await requestJson(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(data)
    });
    closeEditor();
    showFeedback(editingId ? '文章更新成功' : '文章创建成功', 'success');
    renderArticlesLoading('正在刷新文章列表...');
    await loadArticles();
  } catch (error) {
    showFeedback(`保存失败：${error.message}`, 'error');
  } finally {
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
        <div>
          <small>Category</small>
          <strong>${escapeHtml(cat.name)}</strong>
        </div>
        <button class="btn btn-danger" onclick="deleteCat(${cat.id})">删除</button>
      `;
      list.appendChild(div);
    });
  } catch (error) {
    list.innerHTML = renderStateCard(`加载分类失败：${error.message}`, 'error');
  }
}

async function addCategory() {
  const input = document.getElementById('newCat');
  const name = input.value.trim();
  if (!name) {
    alert('请输入分类名称');
    return;
  }

  try {
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
  } catch (error) {
    alert(`添加分类失败：${error.message}`);
  }
}

async function deleteCat(id) {
  if (!confirm('确定删除这个分类吗？')) return;
  try {
    await requestJson(`${API}/categories/${id}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token}` }
    });
    articleCategories = [];
    await loadCategories();
  } catch (error) {
    alert(`删除分类失败：${error.message}`);
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
  const total = Math.ceil(imageCache.length / pageSize);
  if (page < 1 || page > total) return;
  currentPage = page;
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
  const total = Math.ceil(allImages.length / pageSize) || 1;
  if (currentPage > total) currentPage = total;

  if (!allImages.length) {
    list.innerHTML = renderStateCard('图片库还是空的，连一张表情包都没有。', 'empty');
    return;
  }

  const start = (currentPage - 1) * pageSize;
  const end = start + pageSize;
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
      <button class="btn" onclick="changePage(${currentPage - 1})" ${currentPage === 1 ? 'disabled' : ''}>上一页</button>
      <span>第 ${currentPage} / ${total} 页</span>
      <button class="btn" onclick="changePage(${currentPage + 1})" ${currentPage === total ? 'disabled' : ''}>下一页</button>
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
    alert('已复制');
  } catch (err) {
    alert('复制失败');
  }
  document.body.removeChild(textarea);
}

loadArticlesPage();
