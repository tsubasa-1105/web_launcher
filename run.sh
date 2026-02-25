# Dockerイメージのビルド
docker build -t web-launcher .

# 現在のディレクトリに 'data' ディレクトリを作成
mkdir -p ./data

# コンテナを実行 (ポート 8080 をホストの 8080 にマッピング)
docker run -d \
  --restart unless-stopped \
  -p 80:8080 \
  -v $(pwd)/data:/data \
  --name web-launcher \
  web-launcher
