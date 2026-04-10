const API = 'http://172.28.74.191:30083/api';
const token = localStorage.getItem('token');
let editingId = null;
if (!token) location.href = '/admin/login.html';

function showPage(page) {
  document.querySelectorAll('.sidebar a').forEach(a => a.classList.remove('active'));
  event.target.classList.add('active');
  if (page === 'articles') loadArticlesPage();
  else if (page === 'categories') loadCategoriesPage();
  else if (page === 'images') loadImagesPage();
}

function loadArticlesPage() {
  document.getElementById('content').innerHTML = `
    <div class="card">
      <div style="display: flex; justify-content: space-between; margin-bottom: 20px;">
        <h3>文章列表</h3>
        <button class="btn btn-primary" onclick="openEditor()">+ 新建文章</button>
      </div>
      <div id="article-list" style="display: grid; gap: 15px;"></div>
    </div>
  `;
  loadArticles();
}

function loadArticles() {
  fetch(`${API}/articles`).then(res => res.json()).then(json => {
    const list = document.getElementById('article-list');
    list.innerHTML = '';
    json.data.forEach(article => {
      const div = document.createElement('div');
      div.style.cssText = 'border: 1px solid #e9ecef; padding: 20px; border-radius: 8px;';
      div.innerHTML = `<h4>${article.title}</h4><p style="color: #666;">${article.summary || '暂无摘要'}</p><div style="display: flex; justify-content: space-between; margin-top: 10px;"><span style="color: #999;">${new Date(article.created_at).toLocaleDateString()}</span><div><button class="btn btn-primary" onclick="openEditor(${article.id})">编辑</button><button class="btn btn-danger" onclick="deleteArticle(${article.id})">删除</button></div></div>`;
      list.appendChild(div);
    });
  });
}

function deleteArticle(id) {
  if (!confirm('确定删除？')) return;
  fetch(`${API}/articles/${id}`, {method: 'DELETE', headers: {'Authorization': `Bearer ${token}`}}).then(() => loadArticles());
}

function openEditor(id = null) {
  editingId = id;
  document.getElementById('editorTitle').textContent = id ? '编辑文章' : '新建文章';
  document.getElementById('editorModal').style.display = 'block';
  fetch(`${API}/categories`).then(res => res.json()).then(json => {
    const sel = document.getElementById('editCategory');
    sel.innerHTML = '<option value="">选择分类</option>';
    json.data.forEach(c => sel.innerHTML += `<option value="${c.id}">${c.name}</option>`);
  });
  if (id) {
    fetch(`${API}/articles/${id}`).then(res => res.json()).then(json => {
      const a = json.data;
      document.getElementById('editTitle').value = a.title;
      document.getElementById('editSummary').value = a.summary || '';
      document.getElementById('editCover').value = a.cover_image || '';
      document.getElementById('editCategory').value = a.category_id || '';
      document.getElementById('editContent').value = a.content;
    });
  } else {
    document.getElementById('editTitle').value = '';
    document.getElementById('editSummary').value = '';
    document.getElementById('editCover').value = '';
    document.getElementById('editContent').value = '';
  }
}

function closeEditor() { document.getElementById('editorModal').style.display = 'none'; }

function saveArticle() {
  const data = {
    title: document.getElementById('editTitle').value,
    summary: document.getElementById('editSummary').value,
    cover_image: document.getElementById('editCover').value,
    category_id: parseInt(document.getElementById('editCategory').value) || 0,
    content: document.getElementById('editContent').value
  };
  const url = editingId ? `${API}/articles/${editingId}` : `${API}/articles`;
  const method = editingId ? 'PUT' : 'POST';
  fetch(url, {method, headers: {'Content-Type': 'application/json', 'Authorization': `Bearer ${token}`}, body: JSON.stringify(data)}).then(() => { closeEditor(); loadArticles(); });
}

function loadCategoriesPage() {
  document.getElementById('content').innerHTML = `<div class="card"><h3>分类管理</h3><div style="margin: 20px 0;"><input type="text" id="newCat" placeholder="新分类名称" style="width: 300px; display: inline-block;"><button class="btn btn-primary" onclick="addCategory()">添加</button></div><div id="cat-list" style="display: grid; gap: 10px;"></div></div>`;
  loadCategories();
}

function loadCategories() {
  fetch(`${API}/categories`).then(res => res.json()).then(json => {
    const list = document.getElementById('cat-list');
    list.innerHTML = '';
    json.data.forEach(cat => {
      const div = document.createElement('div');
      div.style.cssText = 'border: 1px solid #e9ecef; padding: 15px; border-radius: 8px; display: flex; justify-content: space-between;';
      div.innerHTML = `<span>${cat.name}</span><button class="btn btn-danger" onclick="deleteCat(${cat.id})">删除</button>`;
      list.appendChild(div);
    });
  });
}

function addCategory() {
  const name = document.getElementById('newCat').value;
  if (!name) return alert('请输入分类名称');
  fetch(`${API}/categories`, {method: 'POST', headers: {'Content-Type': 'application/json', 'Authorization': `Bearer ${token}`}, body: JSON.stringify({name})}).then(() => { document.getElementById('newCat').value = ''; loadCategories(); });
}

function deleteCat(id) {
  if (!confirm('确定删除？')) return;
  fetch(`${API}/categories/${id}`, {method: 'DELETE', headers: {'Authorization': `Bearer ${token}`}}).then(() => loadCategories());
}




function logout() {
  localStorage.removeItem('token');
  location.href = '/admin/login.html';
}

loadArticlesPage();


let imageCache = null;
let currentPage = 1;
const pageSize = 20;
function loadImageHistory() {
  if (imageCache) {
    renderImages();
    return;
  }
  fetch(`${API}/images`, {headers: {'Authorization': `Bearer ${token}`}})
  .then(res => res.json())
  .then(json => {
    imageCache = json.data || [];
    renderImages();
  });
}


function changePage(page) {
  const total = Math.ceil(imageCache.length / pageSize);
  if (page < 1 || page > total) return;
  currentPage = page;
  renderImages();
}
function loadImagesPage() {
  document.getElementById('content').innerHTML = `<div class="card"><h3>图床管理</h3><div id="uploadArea" style="border: 2px dashed #ddd; padding: 60px; text-align: center; border-radius: 10px; margin: 20px 0; cursor: pointer;"><p style="font-size: 16px; color: #666;">📷 点击或拖拽图片到此处上传（支持多图）</p><input type="file" id="imgFile" accept="image/*" multiple style="display:none"></div><div id="uploadResult"></div><h4 style="margin-top: 30px;">图片列表</h4><div id="imageHistory"></div></div>`;
  const area = document.getElementById('uploadArea');
  const input = document.getElementById('imgFile');
  area.onclick = () => input.click();
  area.ondragover = (e) => { e.preventDefault(); area.style.borderColor = '#667eea'; };
  area.ondragleave = () => { area.style.borderColor = '#ddd'; };
  area.ondrop = (e) => { e.preventDefault(); area.style.borderColor = '#ddd'; if(e.dataTransfer.files.length) uploadMultiple(e.dataTransfer.files); };
  input.onchange = (e) => { if(e.target.files.length) uploadMultiple(e.target.files); };
  loadImageHistory();
}
function uploadMultiple(files) {
  const result = document.getElementById('uploadResult');
  result.innerHTML = `<p style="text-align:center; color:#666;">上传中... 0/${files.length}</p>`;
  
  let completed = 0;
  const promises = Array.from(files).map(file => {
    const formData = new FormData();
    formData.append('image', file);
    return fetch(`${API}/upload`, {
      method: 'POST',
      headers: {'Authorization': `Bearer ${token}`},
      body: formData
    }).then(() => {
      completed++;
      result.innerHTML = `<p style="text-align:center; color:#666;">上传中... ${completed}/${files.length}</p>`;
    });
  });
  
  Promise.all(promises).then(() => {
    result.innerHTML = `<div style="border: 1px solid #e9ecef; padding: 20px; border-radius: 10px; margin-top: 20px;"><h4 style="color: #51cf66;">✅ 上传成功 ${files.length} 张图片！</h4></div>`;
    imageCache = null;
    loadImageHistory();
  });
}
let selectedImages = [];


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
  
  Promise.all(selectedImages.map(key => 
    fetch(`${API}/images/${key}`, {
      method: 'DELETE',
      headers: {'Authorization': `Bearer ${token}`}
    })
  )).then(() => {
    imageCache = null;
    selectedImages = [];
    loadImageHistory();
  });
}
function renderImages() {
  const list = document.getElementById('imageHistory');
  const start = (currentPage - 1) * pageSize;
  const end = start + pageSize;
  const pageData = imageCache.slice(start, end);
  
  list.innerHTML = '<div style="margin-bottom: 15px;"><button class="btn btn-danger" onclick="deleteSelected()">删除选中</button></div>';
  
  const grid = document.createElement('div');
  grid.style.cssText = 'display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 15px;';
  
  pageData.forEach((item) => {
    const div = document.createElement('div');
    div.style.cssText = 'border: 1px solid #e9ecef; border-radius: 8px; overflow: hidden; position: relative;';
    
    const checkbox = document.createElement('input');
    checkbox.type = 'checkbox';
    checkbox.style.cssText = 'position: absolute; top: 10px; left: 10px; width: 20px; height: 20px;';
    checkbox.onchange = () => toggleSelect(item.key);
    
    const img = document.createElement('img');
    img.src = item.url;
    img.style.cssText = 'width: 100%; height: 150px; object-fit: cover;';
    
    const btnDiv = document.createElement('div');
    btnDiv.style.padding = '10px';
    
    const btn = document.createElement('button');
    btn.className = 'btn btn-primary';
    btn.style.cssText = 'width: 100%;';
    btn.textContent = '复制链接';
    btn.onclick = () => copyUrl(item.url);
    
    btnDiv.appendChild(btn);
    div.appendChild(checkbox);
    div.appendChild(img);
    div.appendChild(btnDiv);
    grid.appendChild(div);
  });
  
  list.appendChild(grid);
  
  const total = Math.ceil(imageCache.length / pageSize);
  if (total > 1) {
    const pagination = document.createElement('div');
    pagination.style.cssText = 'margin-top: 20px; text-align: center;';
    pagination.innerHTML = `<button class="btn" onclick="changePage(${currentPage - 1})" ${currentPage === 1 ? 'disabled' : ''}>上一页</button> <span style="margin: 0 15px;">第 ${currentPage} / ${total} 页</span> <button class="btn" onclick="changePage(${currentPage + 1})" ${currentPage === total ? 'disabled' : ''}>下一页</button>`;
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
