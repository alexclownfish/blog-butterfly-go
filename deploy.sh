#!/bin/bash
set -e

echo "🚀 开始部署 Butterfly 博客..."

cd /root/blog-butterfly-go/backend
echo "📦 构建后端..."
docker build -t blog-butterfly-backend:latest .

cd /root/blog-butterfly-go/frontend
echo "📦 构建前端..."
docker build -t blog-butterfly-frontend:latest .

echo "☸️  部署到 K3s..."
cd /root/blog-butterfly-go/k8s
kubectl apply -f backend.yaml
kubectl apply -f frontend.yaml

echo "✅ 部署完成！"
echo "访问: http://172.28.74.191:30082"
