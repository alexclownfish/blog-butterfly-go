// 动态加载侧边栏
(function() {
  const API = 'http://172.28.74.191:30083/api';
  
  async function loadSidebar() {
    try {
      const res = await fetch(`${API}/articles?page=1&page_size=5`);
      const json = await res.json();
      const articles = json.data || [];
      
      // 更新文章数量
      const countEls = document.querySelectorAll('.length-num');
      if (countEls.length > 0 && json.total) {
        countEls[0].textContent = json.total;
      }
      
      // 更新侧边栏最新文章
      const asideList = document.querySelector('.aside-list');
      if (asideList) {
        asideList.innerHTML = '';
        articles.forEach(article => {
          const div = document.createElement('div');
          div.className = 'aside-list-item';
          div.innerHTML = `
            <a class="thumbnail" href="/post.html?id=${article.id}" title="${article.title}">
              <img src="${article.cover_image || 'https://img.alexcld.com/img/girl3.jpg'}" alt="${article.title}"/>
            </a>
            <div class="content">
              <a class="title" href="/post.html?id=${article.id}" title="${article.title}">${article.title}</a>
              <time>${new Date(article.created_at).toLocaleDateString('zh-CN')}</time>
            </div>
          `;
          asideList.appendChild(div);
        });
      }
    } catch (err) {
      console.error('加载侧边栏失败:', err);
    }
  }
  
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', loadSidebar);
  } else {
    loadSidebar();
  }
})();
