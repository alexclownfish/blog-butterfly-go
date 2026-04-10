const API_BASE = 'http://172.28.74.191:30083/api';

async function fetchArticles() {
  const res = await fetch(`${API_BASE}/articles`);
  const data = await res.json();
  return data.data;
}

async function fetchArticle(id) {
  const res = await fetch(`${API_BASE}/articles/${id}`);
  const data = await res.json();
  return data.data;
}

window.BlogAPI = { fetchArticles, fetchArticle };
