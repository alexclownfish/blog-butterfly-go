#!/bin/bash
set -e

echo "🚀 开始部署 Butterfly 博客..."

echo "📦 构建后端..."
cd /root/blog-butterfly-go/backend
docker build -t blog-butterfly-backend:latest .

echo "📦 构建 web-vue 前台..."
cd /root/blog-butterfly-go/web-vue
docker build -t blog-butterfly-web-vue:latest .

echo "📦 构建 admin-vue 后台..."
cd /root/blog-butterfly-go/admin-vue
docker build -t blog-butterfly-admin-vue:latest .

echo "☸️  部署到 K3s..."
cd /root/blog-butterfly-go/k8s
kubectl apply -f backend.yaml
kubectl apply -f web-vue.yaml
kubectl apply -f admin-vue.yaml

echo "✅ 部署完成！"
echo "前台访问: http://172.28.74.191:31086"
echo "后台访问: http://172.28.74.191:31085"
echo "后端 API: http://172.28.74.191:31083/api"
