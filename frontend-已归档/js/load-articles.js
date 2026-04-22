// 动态加载文章列表
(function() {
  const API = 'http://172.28.74.191:30083/api';
  let currentPage = 1;
  const pageSize = 10;
  
  async function loadArticles(page = 1) {
    try {
      const res = await fetch(`${API}/articles?page=${page}&page_size=${pageSize}`);
      const json = await res.json();
      const articles = json.data || [];
      const total = json.total || 0;
      
      const container = document.getElementById('recent-posts');
      if (!container) return;
      
      // 清空现有内容（保留日历）
      const calendar = container.querySelector('.calendar');
      container.innerHTML = '';
      if (calendar) container.appendChild(calendar);
      
      // 渲染文章列表
      articles.forEach(article => {
        const div = document.createElement('div');
        div.className = 'recent-post-item';
        div.innerHTML = `
          <div class="post_cover left_radius">
            <a href="/post.html?id=${article.id}" title="${article.title}">
              <img class="post_bg" src="${article.cover_image || 'https://img.alexcld.com/img/girl3.jpg'}" alt="${article.title}">
            </a>
          </div>
          <div class="recent-post-info">
            <a class="article-title" href="/post.html?id=${article.id}" title="${article.title}">${article.title}</a>
            <div class="article-meta-wrap">
              <span class="post-meta-date">
                <i class="far fa-calendar-alt"></i>
                <time>${new Date(article.created_at).toLocaleDateString('zh-CN')}</time>
              </span>
              ${article.category ? `<span class="article-meta tags"><i class="fas fa-folder"></i> ${article.category.name}</span>` : ''}
            </div>
            <div class="content">${article.summary || article.content.substring(0, 100)}</div>
          </div>
        `;
        container.appendChild(div);
      });
      
      // 添加分页
      renderPagination(container, page, Math.ceil(total / pageSize));
    } catch (err) {
      console.error('加载文章失败:', err);
    }
  }
  
  function renderPagination(container, current, total) {
    if (total <= 1) return;
    
    const div = document.createElement('div');
    div.className = 'pagination';
    div.style.cssText = 'text-align:center;padding:20px;';
    
    let html = '';
    if (current > 1) {
      html += `<button onclick="loadArticles(${current - 1})">上一页</button> `;
    }
    html += `<span>第 ${current} / ${total} 页</span> `;
    if (current < total) {
      html += `<button onclick="loadArticles(${current + 1})">下一页</button>`;
    }
    
    div.innerHTML = html;
    container.appendChild(div);
  }
  
  window.loadArticles = loadArticles;
  
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => loadArticles(1));
  } else {
    loadArticles(1);
  }
})();
