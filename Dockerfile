# --- ステージ 1: ビルド ---
# Golangのビルド環境
FROM golang:1.21-alpine AS builder

WORKDIR /app

# ビルドに必要なソースコードとアセットを先にコピー
COPY main.go ./
COPY index.html ./

# go.mod ファイルが存在しない場合を考慮し、初期化する
# モジュール名は 'web-launcher' と仮定します
# '|| true' をつけて、既にファイルが存在する場合のエラーを無視します
RUN go mod init web-launcher || true

# main.go を解析し、必要な依存関係を go.mod と go.sum に反映させます
RUN go mod tidy

# 依存関係をダウンロードします（標準ライブラリのみ使用していますが、将来の拡張のため）
RUN go mod download

# アプリケーションをビルドします
# CGO_ENABLED=0: 静的リンクバイナリを生成
# -ldflags "-s -w": バイナリサイズを削減
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/web-launcher main.go

# --- ステージ 2: 実行 ---
# 軽量なAlpineイメージを使用
FROM alpine:latest

WORKDIR /app

# ビルドステージからビルド済みバイナリをコピー
COPY --from=builder /app/web-launcher /app/web-launcher
# SPAフロントエンド (index.html) をコピー
COPY --from=builder /app/index.html /app/index.html

# データ保存用ディレクトリ
# このディレクトリは、コンテナ外部からボリュームマウントすることを強く推奨します
RUN mkdir /data

# ポート8080を開放
EXPOSE 8080

# アプリケーションの実行
# /app/web-launcher を実行
CMD ["/app/web-launcher"]

