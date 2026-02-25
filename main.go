package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// Link はランチャーのリンク項目を表します
type Link struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Color       string `json:"color,omitempty"`       // 追加 (例: "#ffffff")
	Description string `json:"description,omitempty"` // 追加
	Emoji       string `json:"emoji,omitempty"`       // 追加
}

var (
	// データファイルのパス。コンテナ内での使用を想定
	dataPath = "/data"
	dataFile = filepath.Join(dataPath, "links.json")
	mutex    = &sync.Mutex{}
)

// ensureDataDir はデータディレクトリが存在することを確認します
func ensureDataDir() {
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}
	// データファイルが存在しない場合は空のJSON配列で作成
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		log.Println("links.json not found, creating...")
		if err := os.WriteFile(dataFile, []byte("[]"), 0644); err != nil {
			log.Fatalf("Failed to create links.json: %v", err)
		}
	}
}

// loadLinks はJSONファイルからリンクデータを読み込みます
func loadLinks() ([]Link, error) {
	mutex.Lock()
	defer mutex.Unlock()

	data, err := os.ReadFile(dataFile)
	if err != nil {
		// ファイルが存在しない場合も考慮（ensureDataDirで作成されるはずだが念のため）
		if os.IsNotExist(err) {
			return []Link{}, nil
		}
		return nil, err
	}

	var links []Link
	if err := json.Unmarshal(data, &links); err != nil {
		return nil, err
	}
	return links, nil
}

// saveLinks はリンクデータをJSONファイルに保存します
func saveLinks(links []Link) error {
	mutex.Lock()
	defer mutex.Unlock()

	data, err := json.MarshalIndent(links, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(dataFile, data, 0644)
}

// linksHandler は /api/links へのリクエストを処理します
func linksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		links, err := loadLinks()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(links)

	case http.MethodPost:
		var links []Link
		if err := json.NewDecoder(r.Body).Decode(&links); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := saveLinks(links); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(links) // 保存した内容をそのまま返す

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	// データディレクトリとファイルの初期化
	ensureDataDir()

	// APIハンドラ
	http.HandleFunc("/api/links", linksHandler)

	// SPA (index.html) の配信
	// ルートパス "/" は index.html を提供
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
