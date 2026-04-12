# 🎯 Sales Training AI — 銷售培訓 AI 對話系統

一個基於 AI 的銷售培訓對話生成與分析系統，透過 LLM（大型語言模型）自動生成逼真的銷售對話場景，並提供專業的技巧評分與改進建議。採用 WhatsApp 風格的聊天介面，讓學習體驗更加沉浸與直觀。

---

## 📁 專案結構

```
self improve cc/
├── backend/
│   ├── main.go              # Go 後端 API 伺服器
│   ├── go.mod               # Go 模組依賴
│   ├── go.sum               # Go 依賴校驗
│   ├── Dockerfile           # 後端 Docker 映像
│   ├── openllm/
│   │   └── client.go        # OpenAI 相容 API 客戶端
│   └── sales/
│       └── prompt.go        # 銷售培訓 Prompt 生成邏輯
├── frontend/
│   ├── index.html           # Vue.js 前端（WhatsApp 風格 UI）
│   ├── nginx.conf           # Nginx 反向代理設定
│   └── Dockerfile           # 前端 Docker 映像
├── docker-compose.yml       # Docker Compose 編排
└── README.md
```

---

## ✨ 功能特色

### 🤖 AI 銷售對話生成
- 自動生成完整的銷售電話對話（開場 → 需求探索 → 產品介紹 → 異議處理 → 促成交易）
- 支援**英文**與**中文**雙語生成
- 包含對話雙方的**心理策略分析**（💡 氣泡提示）
- 隨機生成成功成交或交易失敗的結局

### 📊 專業評分與分析
- **六大維度評分**：開場破冰、需求探索、產品介紹、異議處理、促成交易、整體表現（1-10 分）
- **優點分析**：自動識別對話中運用得好的銷售技巧
- **改進建議**：針對性地提出可改善之處
- **關鍵學習點**：提取最重要的收穫

### 🔄 迭代改進
- 根據改進建議**重新生成**對話
- 新舊對話**評分比較**（進步/退步視覺化）
- 支援多個對話結果之間的**導航切換**

### 🔊 語音朗讀（TTS）
- 使用瀏覽器原生 Web Speech API
- 支援單條訊息朗讀與全文播放
- 中英文語音自動切換

### ⚙️ n8n 工作流程整合
- 內建 n8n 自動化平台
- 可設定 Webhook 工作流程
- 帳號：`admin` / 密碼：`admin123`

---

## 🛠️ 系統需求

### 必要
- **Docker Desktop**（[下載連結](https://www.docker.com/products/docker-desktop)）
  - 確保 Docker Desktop 已啟動並運行
- **Git**（選用，用於版本控制）

### 本地開發（不用 Docker）
- **Go 1.21+**（[下載連結](https://go.dev/dl/)）
- **現代瀏覽器**（Chrome、Edge、Firefox 等）
- **LLM API 服務**（需支援 OpenAI 相容介面）

---

## 🚀 快速開始

### 方法一：Docker Compose（推薦 ✅）

這是最簡單的方式，一鍵啟動所有服務：

```bash
# 1. 進入專案目錄
cd "c:\Users\User\Desktop\self improve cc"

# 2. 建置並啟動所有服務
docker-compose up --build

# 或在背景運行
docker-compose up --build -d
```

啟動後，開啟瀏覽器前往：
- 🌐 **前端介面**：http://localhost
- 🔧 **後端 API**：http://localhost:8080
- 🤖 **n8n 工作流程**：http://localhost:5678（帳號：`admin` / 密碼：`admin123`）

### 方法二：本地運行（不用 Docker）

#### 步驟 1：啟動後端

```bash
# 進入後端目錄
cd backend

# 安裝依賴
go mod download

# 啟動後端伺服器
go run main.go
```

後端將運行在 http://localhost:8080

#### 步驟 2：開啟前端

直接用瀏覽器開啟 `frontend/index.html`，或使用靜態檔案伺服器：

```bash
# 方法 A：Python
cd frontend
python -m http.server 3000
# 然後開啟 http://localhost:3000

# 方法 B：Node.js
npx serve frontend
# 然後開啟顯示的網址
```

> ⚠️ **注意**：若本地運行前端，API 請求需要指向 `http://localhost:8080`。前端程式碼中的 API 請求可能需要調整。

---

## ⚙️ LLM 設定指南

本系統需要連接一個** OpenAI 相容的 LLM API** 才能生成銷售對話。支援的服務包括：

### 相容的 LLM 服務

| 服務 | Base URL 範例 | 備註 |
|------|--------------|------|
| **OpenAI** | `https://api.openai.com/v1` | 官方 API |
| **Ollama** | `http://localhost:11434/v1` | 本地運行，免費 |
| **LM Studio** | `http://localhost:1234/v1` | 本地運行，免費 |
| **Groq** | `https://api.groq.com/openai/v1` | 免費額度 |
| **Together AI** | `https://api.together.xyz/v1` | 免費額度 |
| **任何 OpenAI 相容 API** | 依服務提供 | — |

### 設定步驟

1. **啟動應用程式**後，開啟 http://localhost
2. 點擊右上角的 **「⚙️ 設定」** 按鈕
3. 填入以下資訊：

| 欄位 | 說明 | 範例 |
|------|------|------|
| **Base URL** | LLM API 的基礎網址 | `https://api.openai.com/v1` |
| **Model** | 使用的模型名稱 | `gpt-4o`、`gpt-3.5-turbo`、`llama3` |
| **API Key** | API 金鑰（密碼欄位） | `sk-...` |

4. 點擊 **「測試連接」** 確認連線成功（綠色圓點 = 已連接）
5. 點擊 **「儲存設定」**

### 推薦的免費本地方案（Ollama）

如果你沒有 OpenAI API Key，可以使用 Ollama 在本地免費運行：

```bash
# 1. 安裝 Ollama（https://ollama.com）
# 2. 下載模型
ollama pull llama3

# 3. 啟動 Ollama（預設運行在 localhost:11434）
ollama serve
```

然後在設定中填入：
- **Base URL**：`http://localhost:11434/v1`
- **Model**：`llama3`
- **API Key**：（留空或填 `ollama`）

> ⚠️ 若使用 Docker 運行本系統，Ollama 的 Base URL 需改為 `http://host.docker.internal:11434/v1`（Windows/Mac）或 `http://172.17.0.1:11434/v1`（Linux），以便容器能存取主機上的 Ollama 服務。

---

## 📖 使用教學

### 生成銷售對話

1. **確認 LLM 已連接**（狀態列顯示綠色圓點）
2. 在 **「情境設定」** 區塊中設定：
   - **產品/服務**：例如「健身房年卡」
   - **客戶類型**：例如「猶豫型」
   - **銷售目標**：例如「促成年度會員銷售」
3. 點擊 **「🟢 生成對話」** 按鈕
4. 等待 AI 生成（可能需要 10-30 秒）

### 查看分析結果

對話生成後，下方的 **「📊 分析報告」** 區塊會顯示：
- **評分矩陣**：六個維度的分數
- **做得好的地方**：綠色卡片
- **可改進的地方**：橙色卡片
- **關鍵學習點**：紫色卡片

### 改進與重新生成

1. 查看分析報告中的 **「可改進的地方」**
2. 點擊 **「🔄 根據改進建議重新生成」** 按鈕
3. 系統會自動將改進建議發送給 LLM，生成更好的對話
4. 可在新舊對話之間切換，比較評分變化

### 切換語言

點擊右上角的 **EN / ZH** 按鈕即可切換英文/中文介面及對話生成語言。

---

## 📡 API 端點

| 方法 | 端點 | 說明 |
|------|------|------|
| `GET` | `/api/health` | 健康檢查 |
| `GET` | `/api/openllm/config` | 取得目前的 LLM 設定 |
| `POST` | `/api/openllm/config` | 更新 LLM 設定 |
| `POST` | `/api/openllm/test` | 測試 LLM 連線 |
| `POST` | `/api/sales/generate` | 生成銷售對話 |
| `POST` | `/api/sales/regenerate` | 根據改進建議重新生成 |

### API 請求範例

**生成銷售對話：**
```bash
curl -X POST http://localhost:8080/api/sales/generate \
  -H "Content-Type: application/json" \
  -d '{
    "product": "健身房年卡",
    "customerType": "猶豫型",
    "goal": "促成年度會員銷售",
    "language": "zh"
  }'
```

**測試 LLM 連線：**
```bash
curl -X POST http://localhost:8080/api/openllm/test
```

---

## 🔧 開發指南

### 修改前端

1. 編輯 `frontend/index.html`
2. 重新建置：`docker-compose up --build frontend`

### 修改後端

1. 編輯 `backend/main.go` 或其他 Go 檔案
2. 重新建置：`docker-compose up --build backend`

### 修改 Dockerfile

- 後端 Dockerfile 需要包含 `openllm/` 和 `sales/` 子目錄（目前只 COPY `main.go`，如需完整建置請更新 Dockerfile）

> ⚠️ **已知問題**：目前的 `backend/Dockerfile` 只複製了 `main.go`，但程式碼引用了 `openllm/` 和 `sales/` 子目錄。在 Docker 建置前，需更新 Dockerfile 以複製所有必要的 Go 原始碼。修正方式見下方。

### 修正後端 Dockerfile

將 `backend/Dockerfile` 中的 COPY 行改為：

```dockerfile
# Copy all source code
COPY . .
```

或者更具體地：

```dockerfile
COPY main.go ./
COPY openllm/ ./openllm/
COPY sales/ ./sales/
COPY go.sum ./
```

---

## 🛑 停止應用程式

```bash
# 停止所有容器
docker-compose down

# 停止並清除資料卷（會刪除 n8n 資料）
docker-compose down -v

# 查看運行中的容器
docker-compose ps

# 查看日誌
docker-compose logs -f
```

---

## ❓ 常見問題

### Q: 生成對話時一直轉圈圈（Loading）
**A:** 可能是 LLM API 連線問題。請檢查：
1. 設定中的 Base URL 是否正確
2. API Key 是否有效
3. 模型名稱是否正確
4. 點擊「測試連接」確認連線狀態

### Q: Docker 建置後端失敗
**A:** 確保 `backend/Dockerfile` 已複製所有必要的子目錄（`openllm/`、`sales/`、`go.sum`）。參考上方「修正後端 Dockerfile」段落。

### Q: 前端無法連接後端
**A:** 若使用 Docker Compose，Nginx 會自動代理 API 請求到後端。若本地運行，需確認後端運行在 `localhost:8080`。

### Q: Ollama 連不上
**A:** 確保 Ollama 正在運行（`ollama serve`），且 Base URL 使用 `http://host.docker.internal:11434/v1`（Docker 環境）或 `http://localhost:11434/v1`（本地環境）。

### Q: n8n 如何使用？
**A:** 開啟 http://localhost:5678，用帳號 `admin`、密碼 `admin123` 登入。可以建立自動化工作流程，例如自動生成銷售對話並發送通知。

---

## 📦 技術棧

| 層級 | 技術 |
|------|------|
| **前端** | Vue.js 3（CDN）、HTML5、CSS3、JavaScript、Web Speech API |
| **後端** | Go 1.21+、OpenAI Go SDK v3 |
| **LLM 整合** | OpenAI 相容 API（OpenAI / Ollama / Groq / 等） |
| **容器化** | Docker、Docker Compose |
| **Web 伺服器** | Nginx（Alpine）— 前端靜態檔案 + API 反向代理 |
| **自動化** | n8n（工作流程引擎） |

---

## 📄 授權

MIT License