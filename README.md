# Config Validator

可配置的 YAML 配置檔驗證工具，專為 CI/CD pipeline 設計。

## 目錄

- [專案簡介](#專案簡介)
- [核心特性](#核心特性)
- [快速開始](#快速開始)
- [完整使用教學](#完整使用教學)
  - [從零開始使用](#從零開始使用)
  - [常見使用場景](#常見使用場景)
  - [新增自訂產品](#新增自訂產品)
  - [自訂規則範例](#自訂規則範例)
  - [疑難排解](#疑難排解)
  - [最佳實踐](#最佳實踐)
- [專案結構](#專案結構)
- [多產品架構](#多產品架構)
- [使用說明](#使用說明)
- [規則系統](#規則系統)
- [CI/CD 整合](#cicd-整合)
- [開發指南](#開發指南)

## 專案簡介

Config Validator 是一個靈活的配置檔驗證工具，通過 YAML 格式的規則檔案來定義驗證邏輯。每條規則都是一個獨立的檔案，支援多種驗證類型，適合在 CI/CD 流程中自動化檢查配置檔的正確性。

### 核心特性

- 🎯 **規則即檔案**：每條規則是一個 YAML 檔案，易於管理和版本控制
- 🏢 **多產品支援**：根據檔案路徑自動選擇對應產品的驗證規則
- 🔧 **靈活配置**：支援 7 種規則類型，涵蓋常見驗證場景
- 🐳 **Docker 友好**：一行指令完成驗證，輕鬆整合進 CI/CD
- 📊 **多種輸出**：支援終端友好和 JSON 格式輸出
- ⚡ **快速執行**：Go 語言實作，效能優異

## 30 秒快速上手

```bash
# 1. Clone 並編譯
git clone <repository-url> && cd config-validator
go build -o validator ./cmd/validator

# 2. 驗證你的配置檔（支援多個路徑）
./validator /path/to/your/configs
# 或
./validator configs/dev configs/staging api.yaml

# 3. 完成！查看驗證結果
# ✅ 所有驗證通過
# 或
# ❌ 2 個錯誤 (顯示詳細錯誤資訊)
```

**這個工具能做什麼？**
- 自動檢測配置檔類型（API、Database、Frontend 等）
- 驗證配置檔的結構、類型、數值範圍等
- 整合到 Git Hook、CI/CD pipeline
- 產生 JSON 格式的驗證報告

## 快速開始

### 使用 Docker（推薦）
```bash
# 建置映像
docker build -t config-validator .

# 驗證配置檔
docker run --rm -v $(pwd)/configs:/configs:ro config-validator /configs

# JSON 格式輸出
docker run --rm -v $(pwd)/configs:/configs:ro config-validator --json /configs

# 測試範例：驗證專案內的測試數據
docker run --rm -v $(pwd)/testdata:/configs:ro config-validator /configs/valid
docker run --rm -v $(pwd)/testdata:/configs:ro config-validator /configs/invalid
```

### 本地執行
```bash
# 安裝依賴
go mod download

# 編譯
go build -o validator ./cmd/validator

# 驗證配置檔（終端輸出）
./validator ./configs

# JSON 格式輸出
./validator --json ./configs

# 測試有效配置
./validator testdata/valid
# 輸出：📋 載入了 7 條規則
#       ✅ 所有驗證通過

# 測試無效配置
./validator testdata/invalid
# 輸出：📋 載入了 7 條規則
#
#       📄 testdata/invalid/api-bad-config.yaml
#         ❌ [api-003] HTTP Method 驗證
#            method 必須是合法的 HTTP 動詞
#            路徑: apiconfig.routes[1].method
#         ❌ [api-004] Route 必要欄位
#            每個 route 必須包含 path, method, handler
#            路徑: apiconfig.routes[0].method
#         ⚠️  [api-005] Timeout 範圍檢查
#            timeout 應在 1000-30000 ms 之間
#            路徑: apiconfig.timeout
#
#       ==================================================
#       ❌ 2 個錯誤
#       ⚠️  1 個警告
```

## 完整使用教學

### 從零開始使用

#### 步驟 1：準備配置檔

假設你有以下配置檔需要驗證：

```bash
# 建立配置檔目錄
mkdir -p my-configs

# 建立 API 配置檔
cat > my-configs/api-service.yaml <<EOF
apiconfig:
  routes:
    - path: /api/users
      method: GET
      handler: getUsersHandler
  timeout: 5000
EOF

# 建立資料庫配置檔
cat > my-configs/database.yaml <<EOF
database:
  host: localhost
  port: 5432
  username: myapp
  database: production
  password: \${DB_PASSWORD}
EOF
```

#### 步驟 2：執行驗證

**方式一：使用本地編譯**
```bash
# 1. Clone 專案
git clone <repository-url>
cd config-validator

# 2. 編譯
go build -o validator ./cmd/validator

# 3. 驗證配置檔
./validator ../my-configs

# 輸出：
# 📋 載入了 2 個產品的規則：
#    • api: 5 條規則
#    • database: 2 條規則
#
# ✅ 所有驗證通過
```

**方式二：使用 Docker**
```bash
# 1. 建置 Docker 映像
docker build -t config-validator .

# 2. 驗證配置檔
docker run --rm -v $(pwd)/my-configs:/configs:ro config-validator /configs
```

#### 步驟 3：查看驗證結果

**成功範例：**
```
📋 載入了 2 個產品的規則：
   • api: 5 條規則
   • database: 2 條規則

📋 載入了 7 條規則

✅ 所有驗證通過
```

**失敗範例：**
```
📋 載入了 1 個產品的規則：
   • api: 5 條規則

📄 my-configs/api-service.yaml
  ❌ [api-003] HTTP Method 驗證
     method 必須是合法的 HTTP 動詞
     路徑: apiconfig.routes[0].method

==================================================
❌ 1 個錯誤
```

### 常見使用場景

#### 場景 1：本地開發驗證

開發時即時檢查配置檔是否正確：

```bash
# 編輯配置檔後立即驗證
vim configs/api-config.yaml
./validator configs/

# 使用 watch 自動監控
watch -n 2 './validator configs/'
```

#### 場景 2：Git Pre-commit Hook

在提交前自動驗證配置檔：

```bash
# .git/hooks/pre-commit
#!/bin/bash
echo "驗證配置檔..."
./validator configs/
if [ $? -ne 0 ]; then
  echo "❌ 配置檔驗證失敗，請修正後再提交"
  exit 1
fi
echo "✅ 配置檔驗證通過"
```

#### 場景 3：CI/CD Pipeline

**GitLab CI：**
```yaml
# .gitlab-ci.yml
validate-configs:
  stage: validate
  image: config-validator:latest
  script:
    - validator /configs --json > validation-report.json
  artifacts:
    reports:
      junit: validation-report.json
    when: always
  only:
    changes:
      - configs/**/*.yaml
```

**GitHub Actions：**
```yaml
# .github/workflows/validate.yml
name: Validate Configs

on:
  pull_request:
    paths:
      - 'configs/**/*.yaml'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build validator
        run: docker build -t config-validator .

      - name: Validate configs
        run: |
          docker run --rm \
            -v ${{ github.workspace }}/configs:/configs:ro \
            config-validator /configs
```

**Jenkins：**
```groovy
// Jenkinsfile
pipeline {
    agent any

    stages {
        stage('Validate Configs') {
            when {
                changeset "configs/**/*.yaml"
            }
            steps {
                sh 'docker build -t config-validator:${BUILD_NUMBER} .'
                sh '''
                    docker run --rm \
                        -v ${WORKSPACE}/configs:/configs:ro \
                        config-validator:${BUILD_NUMBER} /configs
                '''
            }
        }
    }
}
```

#### 場景 4：多環境配置驗證

驗證不同環境的配置檔：

```bash
# 目錄結構
configs/
├── dev/
│   ├── api-config.yaml
│   └── db-config.yaml
├── staging/
│   ├── api-config.yaml
│   └── db-config.yaml
└── production/
    ├── api-config.yaml
    └── db-config.yaml

# 驗證所有環境
for env in dev staging production; do
  echo "驗證 $env 環境..."
  ./validator configs/$env
  if [ $? -ne 0 ]; then
    echo "❌ $env 環境配置有誤"
    exit 1
  fi
done
echo "✅ 所有環境配置正確"
```

#### 場景 5：生成驗證報告

產生詳細的 JSON 報告並分析：

```bash
# 產生 JSON 報告
./validator --json configs/ > report.json

# 使用 jq 分析報告
# 統計錯誤數量
jq '.results | map(select(.severity == "error")) | length' report.json

# 列出所有錯誤檔案
jq -r '.results[] | select(.severity == "error") | .file' report.json | sort -u

# 按規則 ID 分組統計
jq -r '.results[] | .rule_id' report.json | sort | uniq -c
```

### 新增自訂產品

假設你要為「前端配置」添加驗證規則：

#### 步驟 1：建立規則目錄

```bash
mkdir -p rules/frontend
```

#### 步驟 2：建立驗證規則

```bash
cat > rules/frontend/fe-001-theme-check.yaml <<EOF
id: fe-001
name: "主題配置必要欄位"
enabled: true
severity: error
description: "確保主題配置包含所有必要欄位"

targets:
  file_patterns:
    - "**/theme*.yaml"
    - "**/frontend*.yaml"

rule:
  type: required_fields
  path: "theme"
  fields:
    - primaryColor
    - secondaryColor
    - fontFamily
  message: "主題配置缺少必要欄位"
EOF
```

#### 步驟 3：註冊產品

編輯 `products.yaml`，添加：

```yaml
products:
  # ... 現有產品 ...

  - name: frontend
    description: "前端配置驗證"
    rules_dir: rules/frontend
    path_patterns:
      - "**/frontend/**/*.yaml"
      - "**/theme*.yaml"
      - "**/ui*.yaml"
```

#### 步驟 4：測試驗證

```bash
# 建立測試配置
cat > test-frontend.yaml <<EOF
theme:
  primaryColor: "#007bff"
  secondaryColor: "#6c757d"
  fontFamily: "Arial, sans-serif"
EOF

# 執行驗證
./validator test-frontend.yaml

# 輸出：
# 📋 載入了 1 個產品的規則：
#    • frontend: 1 條規則
#
# ✅ 所有驗證通過
```

### 自訂規則範例

#### 範例 1：檢查 API 端點路徑格式

```yaml
# rules/api/api-006-path-format.yaml
id: api-006
name: "API 路徑格式檢查"
enabled: true
severity: warning

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: pattern_match
  path: "apiconfig.routes[*].path"
  pattern: '^/api/[a-z0-9/-]+$'
  message: "API 路徑必須以 /api/ 開頭且只包含小寫字母、數字、斜線和連字號"
```

#### 範例 2：檢查連接池大小範圍

```yaml
# rules/database/db-003-pool-size.yaml
id: db-003
name: "連接池大小檢查"
enabled: true
severity: warning

targets:
  file_patterns:
    - "**/db*.yaml"

rule:
  type: value_range
  path: "database.pool.maxConnections"
  min: 10
  max: 100
  message: "連接池大小應在 10-100 之間"
```

#### 範例 3：檢查環境變數格式

```yaml
# rules/shared/env-var-format.yaml
id: shared-001
name: "環境變數格式檢查"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/*.yaml"

rule:
  type: pattern_match
  path: "*.password"  # 匹配任何名為 password 的欄位
  pattern: '^\$\{[A-Z_]+\}$'
  message: "敏感資訊必須使用環境變數，格式: ${VAR_NAME}（全大寫加底線）"
```

### 疑難排解

#### 問題 1：找不到產品配置

```
錯誤：載入產品配置失敗: 讀取產品配置失敗: open ./products.yaml: no such file or directory
```

**解決方式：**
```bash
# 確認 products.yaml 在專案根目錄
ls -la products.yaml

# 如果在子目錄執行，使用絕對路徑
cd config-validator
./validator /absolute/path/to/configs
```

#### 問題 2：配置檔未被識別

```
⚠️  無法識別配置檔 my-config.yaml 的產品類型，跳過驗證
```

**解決方式：**
```bash
# 檢查檔案命名是否符合 products.yaml 中的 path_patterns
# 或者修改 products.yaml 添加新的匹配模式

# 例如：將檔案重命名為符合模式的名稱
mv my-config.yaml api-config.yaml
```

#### 問題 3：規則目錄不存在

```
錯誤：載入產品 api 的規則失敗: 規則目錄不存在: ./rules/api
```

**解決方式：**
```bash
# 確認規則目錄結構
ls -la rules/

# 確保 products.yaml 中的 rules_dir 路徑正確
# 並且目錄內有 .yaml 規則檔案
```

### 最佳實踐

#### 1. 規則命名規範

```
<product>-<number>-<description>.yaml

範例：
- api-001-required-fields.yaml
- db-001-connection-check.yaml
- fe-001-theme-validation.yaml
```

#### 2. 規則組織策略

```
rules/
├── api/              # API 產品規則
│   ├── api-001-*.yaml
│   ├── api-002-*.yaml
│   └── ...
├── database/         # 資料庫規則
├── frontend/         # 前端規則
└── shared/           # 共用規則（如果需要）
```

#### 3. Severity 使用指南

- **error**：配置錯誤會導致系統無法運作，必須修正
  - 例如：缺少必要欄位、資料類型錯誤

- **warning**：配置不理想但系統仍可運作，建議修正
  - 例如：數值超出建議範圍、命名不符合規範

- **info**：提示性訊息，可選擇性修正
  - 例如：建議添加的欄位、優化建議

#### 4. CI/CD 整合建議

```bash
# 在 CI/CD 中設定嚴格模式
# 將 warning 也視為失敗
./validator configs/
EXIT_CODE=$?
if [ $EXIT_CODE -ne 0 ]; then
  echo "配置驗證失敗"
  exit 1
fi

# 或者只在有 error 時失敗
./validator --json configs/ > report.json
ERROR_COUNT=$(jq '.results | map(select(.severity == "error")) | length' report.json)
if [ $ERROR_COUNT -gt 0 ]; then
  echo "發現 $ERROR_COUNT 個錯誤"
  exit 1
fi
```

## 專案結構
```
config-validator/
├── cmd/
│   └── validator/
│       └── main.go                    # 程式入口
│
├── internal/
│   ├── rule/
│   │   ├── types.go                   # Rule 結構定義
│   │   ├── loader.go                  # 規則載入器
│   │   └── executor.go                # 規則執行引擎
│   ├── parser/
│   │   └── yaml.go                    # YAML 解析器
│   ├── product/                        # 產品檢測模組 ⭐
│   │   ├── types.go                   # 產品配置結構
│   │   └── detector.go                # 產品自動檢測器
│   └── reporter/
│       └── reporter.go                # 結果輸出器
│
├── rules/                              # 規則定義資料夾（按產品分類）⭐
│   ├── api/                           # API 產品規則
│   │   ├── api-001-required-fields.yaml
│   │   ├── api-002-routes-structure.yaml
│   │   ├── api-003-method-validation.yaml
│   │   ├── api-004-route-required-fields.yaml
│   │   └── api-005-timeout-range.yaml
│   └── database/                      # Database 產品規則
│       ├── db-001-required-fields.yaml
│       └── db-002-password-check.yaml
│
├── testdata/                           # 測試配置檔
│   ├── valid/
│   │   ├── api-config.yaml
│   │   └── db-config.yaml
│   ├── invalid/
│   │   └── api-bad-config.yaml
│   └── mixed/                          # 混合多產品測試
│       ├── api-config.yaml
│       └── db-config.yaml
│
├── products.yaml                       # 產品與規則映射配置 ⭐
├── Dockerfile
├── .dockerignore
├── go.mod
├── go.sum
└── README.md
```

## 多產品架構

### 產品自動檢測

驗證器會根據配置檔路徑自動判斷產品類型，並載入對應的驗證規則。這通過 `products.yaml` 配置檔實現。

### 產品配置檔格式

`products.yaml` 定義了產品與規則的映射關係：

```yaml
products:
  - name: api                          # 產品名稱
    description: "API 配置驗證"        # 產品描述
    rules_dir: rules/api               # 規則目錄
    path_patterns:                     # 路徑匹配模式
      - "**/api/**/*.yaml"             # 匹配 api 目錄下的所有 YAML
      - "**/api*.yaml"                 # 匹配檔名以 api 開頭的 YAML
      - "**/routes*.yaml"              # 匹配檔名包含 routes 的 YAML

  - name: database
    description: "資料庫配置驗證"
    rules_dir: rules/database
    path_patterns:
      - "**/db/**/*.yaml"
      - "**/database*.yaml"
```

### 新增產品

要為新產品添加驗證規則：

1. **建立規則目錄**
   ```bash
   mkdir -p rules/frontend
   ```

2. **添加驗證規則**
   在 `rules/frontend/` 中建立規則檔案

3. **更新產品配置**
   在 `products.yaml` 中添加產品定義：
   ```yaml
   products:
     - name: frontend
       description: "前端配置驗證"
       rules_dir: rules/frontend
       path_patterns:
         - "**/frontend/**/*.yaml"
         - "**/theme*.yaml"
   ```

4. **測試驗證**
   ```bash
   ./validator ./configs/frontend
   ```

### 工作原理

1. **掃描配置檔**：遍歷指定目錄下的所有 YAML 檔案
2. **檢測產品類型**：根據檔案路徑匹配 `products.yaml` 中的模式
3. **載入規則**：根據產品類型載入對應目錄的驗證規則
4. **執行驗證**：使用載入的規則驗證配置檔
5. **輸出結果**：顯示每個產品的驗證結果統計

### 範例：混合產品驗證

```bash
# 目錄結構
configs/
├── api-service.yaml          # 自動使用 api 規則
├── db-connection.yaml        # 自動使用 database 規則
└── routes-config.yaml        # 自動使用 api 規則

# 執行驗證
./validator configs/

# 輸出
📋 載入了 2 個產品的規則：
   • api: 5 條規則
   • database: 2 條規則

✅ 所有驗證通過
```

## 使用說明

### 基本語法
```bash
validator <path1> [path2] [path3] ... [--json]
```

**參數說明：**
- `<path>`：配置檔或目錄路徑（必填，可指定多個）
  - 可以是單一檔案：`api-config.yaml`
  - 可以是目錄：`configs/`
  - 可以混合使用：`configs/ extra.yaml`
- `--json`：輸出 JSON 格式（可選）

**退出碼：**
- `0`：驗證通過
- `1`：驗證失敗（有 error 級別的問題）

### 使用範例

#### 驗證單個目錄
```bash
# 驗證 configs 目錄下的所有 YAML 檔案
./validator ./configs
```

#### 驗證多個目錄
```bash
# 一次驗證多個目錄
./validator configs/dev configs/staging configs/prod

# 範例輸出會包含所有目錄的檔案
```

#### 驗證指定檔案
```bash
# 只驗證特定檔案
./validator configs/api-config.yaml configs/db-config.yaml

# 支援相對路徑和絕對路徑
./validator /etc/myapp/config.yaml ./local-config.yaml
```

#### 混合驗證（目錄 + 檔案）
```bash
# 驗證整個目錄加上額外的單獨檔案
./validator configs/ extra-settings.yaml special/override.yaml
```

#### 驗證並輸出 JSON
```bash
# 適合用於 CI/CD 或程式化處理
./validator --json ./configs > validation-report.json

# 多路徑 JSON 輸出
./validator --json configs/ testdata/ > full-report.json
```

#### 檢查退出碼
```bash
# 在腳本中使用
./validator ./configs
if [ $? -eq 0 ]; then
  echo "✅ 配置驗證通過"
else
  echo "❌ 配置驗證失敗"
  exit 1
fi
```

#### 使用 Docker
```bash
# 驗證當前目錄下的 configs 資料夾
docker run --rm -v $(pwd)/configs:/configs:ro config-validator /configs

# JSON 輸出
docker run --rm -v $(pwd)/configs:/configs:ro config-validator --json /configs
```

### 輸出格式

#### 終端輸出（預設）
```
📋 載入了 5 條規則

📄 configs/api-config.yaml
  ❌ [api-003] HTTP Method 驗證
     method 必須是合法的 HTTP 動詞
     路徑: apiconfig.routes[0].method
  ⚠️  [api-005] Timeout 範圍檢查
     timeout 應在 1000-30000 ms 之間
     路徑: apiconfig.timeout

==================================================
❌ 1 個錯誤
⚠️  1 個警告
```

#### JSON 輸出
```json
{
  "total": 2,
  "results": [
    {
      "file": "configs/api-config.yaml",
      "rule_id": "api-003",
      "rule_name": "HTTP Method 驗證",
      "severity": "error",
      "message": "method 必須是合法的 HTTP 動詞",
      "path": "apiconfig.routes[0].method"
    },
    {
      "file": "configs/api-config.yaml",
      "rule_id": "api-005",
      "rule_name": "Timeout 範圍檢查",
      "severity": "warning",
      "message": "timeout 應在 1000-30000 ms 之間",
      "path": "apiconfig.timeout"
    }
  ]
}
```

## 規則系統

### 支援的規則類型

| 規則類型 | 說明 | 使用場景 |
|---------|------|---------|
| `required_field` | 檢查必要欄位是否存在 | 確保關鍵配置不遺漏 |
| `required_fields` | 檢查多個必要欄位 | 批次檢查多個必要欄位 |
| `field_type` | 檢查欄位型別 | 確保資料型別正確 |
| `value_range` | 檢查數值範圍 | 驗證數值在合理範圍內 |
| `array_item_required_fields` | 檢查陣列項目的必要欄位 | 驗證陣列中每個物件的結構 |
| `array_item_field` | 檢查陣列項目的欄位值 | 驗證陣列項目的枚舉值 |
| `pattern_match` | 正則表達式驗證 | 驗證字串格式 |

### 規則檔案格式

每個規則檔案包含以下欄位：
```yaml
id: string              # 規則唯一識別碼（必填）
name: string            # 規則名稱（必填）
enabled: boolean        # 是否啟用（必填）
severity: string        # error/warning/info（必填）
description: string     # 規則描述（可選）
targets:                # 適用目標（必填）
  file_patterns:        # 檔案匹配 pattern 陣列
    - string
rule:                   # 驗證邏輯（必填）
  type: string          # 規則類型
  # ... 其他參數依規則類型而異
```

### 規則範例

#### 範例 1：必要欄位檢查
```yaml
# rules/api-001-required-fields.yaml
id: api-001
name: "API Config 必要欄位檢查"
enabled: true
severity: error
description: "確保 apiconfig 區塊存在"

targets:
  file_patterns:
    - "**/api*.yaml"
    - "**/routes*.yaml"

rule:
  type: required_field
  path: "apiconfig"
  message: "缺少 apiconfig 區塊"
```

#### 範例 2：型別檢查
```yaml
# rules/api-002-routes-structure.yaml
id: api-002
name: "Routes 必須是陣列"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: field_type
  path: "apiconfig.routes"
  expected_type: array
  message: "routes 必須是陣列格式"
```

#### 範例 3：枚舉值驗證
```yaml
# rules/api-003-method-validation.yaml
id: api-003
name: "HTTP Method 驗證"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: array_item_field
  path: "apiconfig.routes"
  field: "method"
  validation:
    type: enum
    allowed_values:
      - GET
      - POST
      - PUT
      - DELETE
      - PATCH
  message: "method 必須是合法的 HTTP 動詞"
```

#### 範例 4：陣列項目必要欄位
```yaml
# rules/api-004-route-required-fields.yaml
id: api-004
name: "Route 必要欄位"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: array_item_required_fields
  path: "apiconfig.routes"
  required_fields:
    - path
    - method
    - handler
  message: "每個 route 必須包含 path, method, handler"
```

#### 範例 5：數值範圍檢查
```yaml
# rules/api-005-timeout-range.yaml
id: api-005
name: "Timeout 範圍檢查"
enabled: true
severity: warning

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: value_range
  path: "apiconfig.timeout"
  min: 1000
  max: 30000
  message: "timeout 應在 1000-30000 ms 之間"
```

#### 範例 6：多個必要欄位
```yaml
# rules/db-001-required-fields.yaml
id: db-001
name: "Database 必要欄位"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/db*.yaml"
    - "**/database*.yaml"

rule:
  type: required_fields
  path: "database"
  fields:
    - host
    - port
    - username
    - database
  message: "database 缺少必要欄位"
```

#### 範例 7：正則驗證
```yaml
# rules/db-002-password-check.yaml
id: db-002
name: "密碼不應 hardcode"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/db*.yaml"

rule:
  type: pattern_match
  path: "database.password"
  pattern: '^\$\{.*\}$'
  message: "password 必須使用環境變數，格式: ${VAR_NAME}"
```

## 配置檔範例

### 有效的 API 配置
```yaml
# testdata/valid/api-config.yaml
apiconfig:
  routes:
    - path: /api/users
      method: GET
      handler: getUsersHandler
    - path: /api/users/:id
      method: POST
      handler: createUserHandler
  timeout: 5000
```

### 無效的 API 配置
```yaml
# testdata/invalid/api-bad-config.yaml
apiconfig:
  routes:
    - path: /api/users
      # ❌ 缺少 method
      handler: getUsersHandler
    - path: /api/posts
      method: INVALID_METHOD  # ❌ 不合法的 method
      handler: getPostsHandler
  timeout: 50000  # ❌ 超過範圍
```

### 有效的 Database 配置
```yaml
# testdata/valid/db-config.yaml
database:
  host: localhost
  port: 5432
  username: dbuser
  database: myapp
  password: ${DB_PASSWORD}  # ✅ 使用環境變數
  pool:
    maxConnections: 50
```

## CI/CD 整合

### GitLab CI
```yaml
# .gitlab-ci.yml
stages:
  - validate

validate-configs:
  stage: validate
  image: config-validator:latest
  script:
    - validator /configs --json
  artifacts:
    reports:
      junit: validation-results.json
    when: always
  rules:
    - changes:
        - configs/**/*.yaml
```

### GitHub Actions
```yaml
# .github/workflows/validate.yml
name: Validate Configs

on:
  pull_request:
    paths:
      - 'configs/**/*.yaml'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Build validator
        run: docker build -t config-validator .
      
      - name: Validate configs
        run: |
          docker run --rm \
            -v ${{ github.workspace }}/configs:/configs:ro \
            config-validator /configs
```

### Jenkins
```groovy
// Jenkinsfile
pipeline {
    agent any
    
    stages {
        stage('Build Validator') {
            steps {
                sh 'docker build -t config-validator:${BUILD_NUMBER} .'
            }
        }
        
        stage('Validate Configs') {
            steps {
                sh '''
                    docker run --rm \
                        -v ${WORKSPACE}/configs:/configs:ro \
                        config-validator:${BUILD_NUMBER} /configs
                '''
            }
        }
    }
    
    post {
        failure {
            echo 'Config validation failed!'
        }
    }
}
```

## 實際測試

### 快速驗證專案

專案已包含測試數據，可以立即測試：

```bash
# 1. 下載依賴並編譯
go mod download
go build -o validator ./cmd/validator

# 2. 測試有效配置（預期通過）
./validator testdata/valid
```

**預期輸出：**
```
📋 載入了 2 個產品的規則：
   • api: 5 條規則
   • database: 2 條規則

📋 載入了 7 條規則

✅ 所有驗證通過
```

```bash
# 3. 測試無效配置（預期失敗）
./validator testdata/invalid
```

**預期輸出：**
```
📋 載入了 1 個產品的規則：
   • api: 5 條規則

📋 載入了 5 條規則

📄 testdata/invalid/api-bad-config.yaml
  ❌ [api-003] HTTP Method 驗證
     method 必須是合法的 HTTP 動詞
     路徑: apiconfig.routes[1].method
  ❌ [api-004] Route 必要欄位
     每個 route 必須包含 path, method, handler
     路徑: apiconfig.routes[0].method
  ⚠️  [api-005] Timeout 範圍檢查
     timeout 應在 1000-30000 ms 之間
     路徑: apiconfig.timeout

==================================================
❌ 2 個錯誤
⚠️  1 個警告
```

```bash
# 4. 測試 JSON 輸出
./validator --json testdata/invalid
```

**預期輸出：**
```json
{
  "results": [
    {
      "file": "testdata/invalid/api-bad-config.yaml",
      "rule_id": "api-003",
      "rule_name": "HTTP Method 驗證",
      "severity": "error",
      "message": "method 必須是合法的 HTTP 動詞",
      "path": "apiconfig.routes[1].method"
    },
    {
      "file": "testdata/invalid/api-bad-config.yaml",
      "rule_id": "api-004",
      "rule_name": "Route 必要欄位",
      "severity": "error",
      "message": "每個 route 必須包含 path, method, handler",
      "path": "apiconfig.routes[0].method"
    },
    {
      "file": "testdata/invalid/api-bad-config.yaml",
      "rule_id": "api-005",
      "rule_name": "Timeout 範圍檢查",
      "severity": "warning",
      "message": "timeout 應在 1000-30000 ms 之間",
      "path": "apiconfig.timeout"
    }
  ],
  "total": 3
}
```

### 混合產品測試

```bash
# 5. 測試混合產品配置（包含多種產品類型）
./validator testdata/mixed
```

**預期輸出：**
```
📋 載入了 2 個產品的規則：
   • api: 5 條規則
   • database: 2 條規則

📋 載入了 7 條規則

✅ 所有驗證通過
```

這展示了驗證器如何自動識別不同產品的配置檔並套用對應的驗證規則。

### 檢查退出碼

```bash
# 成功時退出碼為 0
./validator testdata/valid
echo $?  # 輸出：0

# 失敗時退出碼為 1
./validator testdata/invalid
echo $?  # 輸出：1
```

## 開發指南

### 新增規則

1. 在 `rules/` 資料夾建立新的 YAML 檔案
2. 定義規則內容（參考上方範例）
3. 重新建置 Docker 映像
```bash
# 範例：新增自訂規則
cat > rules/custom-001.yaml <<EOF
id: custom-001
name: "自訂規則"
enabled: true
severity: warning
targets:
  file_patterns: ["**/*.yaml"]
rule:
  type: required_field
  path: "version"
  message: "建議加上 version 欄位"
EOF

# 重新建置
docker build -t config-validator .
```

### 停用規則

將規則檔案中的 `enabled` 設為 `false`：
```yaml
id: api-005
name: "Timeout 範圍檢查"
enabled: false  # 停用此規則
severity: warning
# ...
```

### 本地開發測試
```bash
# 下載依賴
go mod download

# 編譯專案
go build -o validator ./cmd/validator

# 測試有效配置（應該通過）
./validator testdata/valid
# 預期輸出：✅ 所有驗證通過

# 測試無效配置（應該報錯）
./validator testdata/invalid
# 預期輸出：顯示錯誤和警告

# 測試 JSON 輸出
./validator --json testdata/invalid

# 執行驗證（開發模式，不需編譯）
go run ./cmd/validator ./testdata/valid

# 執行測試（如果有寫測試）
go test ./...

# 檢查程式碼品質
go vet ./...
gofmt -s -w .
```

## Dockerfile
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o validator ./cmd/validator

FROM alpine:3.19
COPY --from=builder /app/validator /validator
COPY rules /rules
WORKDIR /workspace
ENTRYPOINT ["/validator"]
```

## 技術規格

### 依賴項目

- **Go**: 1.21+
- **gopkg.in/yaml.v3**: YAML 解析

### 支援的配置檔格式

- YAML (`.yaml`, `.yml`)

### 路徑表達式

- 支援點號分隔：`database.pool.maxConnections`
- 支援陣列索引顯示：`routes[0].method`（僅用於輸出）

### 型別對應

| YAML 型別 | Go 型別 | 驗證器型別 |
|----------|---------|-----------|
| 字串 | `string` | `string` |
| 數字 | `int/float64` | `number` |
| 布林 | `bool` | `boolean` |
| 陣列 | `[]interface{}` | `array` |
| 物件 | `map[string]interface{}` | `object` |

## 常見問題

### Q: 如何驗證多層巢狀結構？

A: 使用點號分隔的路徑，例如：`database.pool.maxConnections`

### Q: 規則的優先順序如何？

A: 規則按檔案名稱字母順序載入和執行，但彼此獨立，無優先順序關係。

### Q: 可以在規則中使用正則表達式嗎？

A: 可以，使用 `pattern_match` 類型規則，參考範例 7。

### Q: 如何處理可選欄位？

A: 不要為可選欄位建立 `required_field` 規則即可。

### Q: 能否自訂錯誤訊息？

A: 可以，每條規則都有 `message` 欄位可自訂錯誤訊息。

## 擴展性

### 未來可能新增的功能

- [ ] 條件驗證（if-then 邏輯）
- [ ] 欄位間的參照檢查
- [ ] 自訂腳本規則（Lua/JavaScript）
- [ ] 規則標籤分類
- [ ] 自動修復建議
- [ ] Web UI 介面

### 貢獻指南

歡迎提交 Issue 和 Pull Request！

1. Fork 本專案
2. 建立功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交變更 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 開啟 Pull Request

## 授權

MIT License

## 聯絡方式

如有問題或建議，請開啟 Issue。