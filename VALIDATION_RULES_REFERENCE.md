# 驗證規則完整參考手冊

## 📋 目錄

- [驗證功能總覽](#驗證功能總覽)
- [API 產品驗證規則](#api-產品驗證規則)
- [Database 產品驗證規則](#database-產品驗證規則)
- [規則類型速查表](#規則類型速查表)
- [YAML 規則模板](#yaml-規則模板)
- [進階功能](#進階功能)
- [使用建議](#使用建議)

---

## 驗證功能總覽

目前專案包含 **2 個產品** 共 **14 條驗證規則**：

| 產品 | 規則數量 | 規則目錄 |
|------|---------|---------|
| API | 12 條 | `rules/api/` |
| Database | 2 條 | `rules/database/` |

**檔案配對規則：**
- API 產品：`**/api*.yaml`、`**/routes*.yaml`
- Database 產品：`**/db*.yaml`、`**/database*.yaml`

---

## API 產品驗證規則

### api-001：API Config 必要欄位檢查

**檔案：** `rules/api/api-001-required-fields.yaml`

```yaml
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

**驗證內容：** 確保配置檔中存在 `apiconfig` 根節點

**適用檔案：** API 配置檔

---

### api-002：Routes 必須是陣列

**檔案：** `rules/api/api-002-routes-structure.yaml`

```yaml
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

**驗證內容：** 確保 `routes` 欄位是陣列類型，而非字串或物件

**適用檔案：** API 配置檔

---

### api-003：HTTP Method 驗證

**檔案：** `rules/api/api-003-method-validation.yaml`

```yaml
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

**驗證內容：** 每個 route 的 `method` 必須是以下值之一：
- GET
- POST
- PUT
- DELETE
- PATCH

**適用檔案：** API 配置檔

---

### api-004：Route 必要欄位

**檔案：** `rules/api/api-004-route-required-fields.yaml`

```yaml
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

**驗證內容：** 每個 route 必須包含以下欄位：
- `path`：API 路徑
- `method`：HTTP 方法
- `handler`：處理函式

**適用檔案：** API 配置檔

---

### api-005：Timeout 範圍檢查

**檔案：** `rules/api/api-005-timeout-range.yaml`

```yaml
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

**驗證內容：** timeout 數值必須在 1000-30000 毫秒之間

**嚴重程度：** 警告（warning）- 超出範圍仍可運作但不建議

**適用檔案：** API 配置檔

---

### api-007：Route path 不可重複

**檔案：** `rules/api/api-007-no-duplicate-path.yaml`

```yaml
id: api-007
name: "Route path 不可重複"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: array_no_duplicates
  path: "apiconfig.routes"
  field: "path"
  message: "routes 中的 path 欄位不可重複"
```

**驗證內容：** 檢查所有 routes 中的 `path` 欄位是否有重複值

**範例錯誤：**
```yaml
routes:
  - path: /api/users  # ❌ 重複
    method: GET
  - path: /api/users  # ❌ 重複
    method: POST
```

**適用檔案：** API 配置檔

---

### api-008：Route path+method 組合不可重複

**檔案：** `rules/api/api-008-no-duplicate-path-method.yaml`

```yaml
id: api-008
name: "Route path+method 組合不可重複"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: array_no_duplicates_combine
  path: "apiconfig.routes"
  fields:
    - path
    - method
  message: "routes 中的 path+method 組合不可重複"
```

**驗證內容：** 檢查 `path` 和 `method` 的組合是否唯一

**範例：**
```yaml
routes:
  - path: /api/users
    method: GET      # ✓ /api/users+GET（唯一）
  - path: /api/users
    method: POST     # ✓ /api/users+POST（唯一）
  - path: /api/users
    method: GET      # ❌ /api/users+GET（重複）
```

**適用檔案：** API 配置檔

---

### api-009：Middleware name 不可重複

**檔案：** `rules/api/api-009-middleware-no-duplicate.yaml`

```yaml
id: api-009
name: "每個 route 的 middlewares 中 name 不可重複"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: array_no_duplicates
  path: "apiconfig.routes[*].middlewares"
  field: "name"
  message: "middleware 的 name 在同一個 route 中不可重複"
```

**驗證內容：** 檢查同一個 route 的 middlewares 中，`name` 欄位不可重複

**萬用字元說明：** `routes[*].middlewares` 表示檢查所有 routes 的 middlewares

**範例錯誤：**
```yaml
routes:
  - path: /api/users
    middlewares:
      - name: auth      # ✓
      - name: logging   # ✓
      - name: auth      # ❌ 同一 route 中重複
```

**適用檔案：** API 配置檔

---

### api-010：Middleware 必要欄位

**檔案：** `rules/api/api-010-middleware-required-fields.yaml`

```yaml
id: api-010
name: "Middleware 必要欄位"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: array_item_required_fields
  path: "apiconfig.routes[*].middlewares"
  required_fields:
    - name
    - priority
  message: "每個 middleware 必須包含 name 和 priority 欄位"
```

**驗證內容：** 所有 routes 的所有 middlewares 必須包含：
- `name`：中介軟體名稱
- `priority`：優先順序

**萬用字元說明：** `routes[*].middlewares` 自動檢查所有 route 的所有 middleware

**適用檔案：** API 配置檔

---

### api-011：弱密碼檢查

**檔案：** `rules/api/api-011-password-weak-check.yaml`

```yaml
id: api-011
name: "弱密碼檢查"
enabled: true
severity: error
description: "檢查密碼是否為常見弱密碼（使用 SHA256 雜湊比對）"

targets:
  file_patterns:
    - "**/api*.yaml"
    - "**/admin*.yaml"
    - "**/user*.yaml"

rule:
  type: hashed_value_check
  path: "admin.password"
  hash_algorithm: "sha256"
  mode: "forbidden"
  hash_list:
    - "240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9"  # admin123
    - "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"  # password
    - "8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918"  # admin
  message: "密碼不可使用常見弱密碼 (admin123, password, admin, qwerty 等)"
```

**驗證內容：** 將密碼進行 SHA256 雜湊運算，檢查是否在常見弱密碼列表中

**雜湊演算法：** sha1, sha256, sha512, md5

**模式說明：**
- `forbidden`：禁止使用列表中的雜湊值（弱密碼檢測）
- `allowed`：只允許列表中的雜湊值（授權碼驗證）

**適用檔案：** API、Admin、User 配置檔

---

### api-012：敏感關鍵字檢查

**檔案：** `rules/api/api-012-sensitive-keywords.yaml`

```yaml
id: api-012
name: "敏感關鍵字檢查"
enabled: true
severity: warning
description: "檢查 API 路徑是否包含敏感關鍵字"

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: contains_keywords
  path: "apiconfig.routes[*].path"
  mode: "forbidden"
  case_sensitive: false
  keywords:
    - "internal"
    - "private"
    - "admin"
    - "debug"
    - "test"
    - "secret"
  message: "API 路徑不應包含敏感關鍵字"
```

**驗證內容：** 檢查字串欄位是否包含（或必須包含）特定關鍵字

**支援萬用字元：** ✅ 支援 `[*]` 路徑

**模式說明：**
- `forbidden`：不可包含列表中的關鍵字（敏感字詞過濾）
- `required`：必須包含列表中至少一個關鍵字（強制格式）

**大小寫設定：**
- `case_sensitive: false`：不區分大小寫（預設）
- `case_sensitive: true`：區分大小寫

**適用檔案：** API 配置檔

---

### api-013：欄位前後不可有空白

**檔案：** `rules/api/api-013-no-trailing-whitespace.yaml`

```yaml
id: api-013
name: "欄位前後不可有空白"
enabled: true
severity: warning
description: "自動檢查整個配置檔中所有字串欄位前後是否有多餘的空白字元"

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: no_trailing_whitespace
  message: "配置檔中的字串欄位前後不可有空白字元"
```

**驗證內容：** 自動掃描整個 YAML 檔案中的**所有字串欄位**，檢查前後是否有空格或 Tab

**特色：** ⭐ **全檔自動掃描** - 不需要指定 `path`，會自動檢查所有字串值

**檢查類型：**
- 開頭空白（Leading whitespace）
- 結尾空白（Trailing whitespace）
- 同時檢查空格和 Tab 字元

**錯誤訊息範例：**
```
⚠️  [api-013] 欄位前後不可有空白
   配置檔中的字串欄位前後不可有空白字元 (結尾有空白字元)
   路徑: apiconfig.routes[0].path

⚠️  [api-013] 欄位前後不可有空白
   配置檔中的字串欄位前後不可有空白字元 (開頭有空白字元)
   路徑: admin.username
```

**適用檔案：** API 配置檔

---

## Database 產品驗證規則

### db-001：Database 必要欄位

**檔案：** `rules/database/db-001-required-fields.yaml`

```yaml
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

**驗證內容：** database 配置必須包含以下欄位：
- `host`：資料庫主機位址
- `port`：連接埠號
- `username`：使用者名稱
- `database`：資料庫名稱

**適用檔案：** Database 配置檔

---

### db-002：密碼不應 hardcode

**檔案：** `rules/database/db-002-password-check.yaml`

```yaml
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

**驗證內容：** 密碼必須使用環境變數格式，不可直接寫入明碼

**正確格式：** `${DB_PASSWORD}`、`${DATABASE_PASSWORD}`

**錯誤格式：** `my_password123`、`admin`

**適用檔案：** Database 配置檔

---

## 規則類型速查表

| 規則類型 | 功能說明 | 參數 | 使用場景 |
|---------|---------|------|---------|
| `required_field` | 檢查單一必要欄位是否存在 | `path`, `message` | 確保關鍵配置不遺漏 |
| `required_fields` | 檢查多個必要欄位 | `path`, `fields`, `message` | 批次檢查多個必要欄位 |
| `field_type` | 檢查欄位型別 | `path`, `expected_type`, `message` | 確保資料型別正確 |
| `value_range` | 檢查數值範圍 | `path`, `min`, `max`, `message` | 驗證數值在合理範圍內 |
| `array_item_required_fields` | 檢查陣列項目的必要欄位 | `path`, `required_fields`, `message` | 驗證陣列中每個物件的結構 |
| `array_item_field` | 檢查陣列項目的欄位值 | `path`, `field`, `validation`, `message` | 驗證陣列項目的列舉值 |
| `pattern_match` | 正則表達式驗證 | `path`, `pattern`, `message` | 驗證字串格式 |
| `array_no_duplicates` | 檢查陣列欄位不重複 | `path`, `field`, `message` | 確保陣列中某欄位值唯一 |
| `array_no_duplicates_combine` | 檢查陣列多欄位組合不重複 | `path`, `fields`, `message` | 確保多欄位組合唯一 |
| `hashed_value_check` | SHA 雜湊值檢查 | `path`, `hash_algorithm`, `mode`, `hash_list`, `message` | 弱密碼檢測、授權碼驗證 |
| `contains_keywords` | 關鍵字檢查 | `path`, `mode`, `case_sensitive`, `keywords`, `message` | 敏感字詞過濾、格式強制 |
| `no_trailing_whitespace` | Trailing/Leading 空白檢查（全檔掃描） | `message` | 資料品質檢查 |

### 支援的資料型別

| YAML 型別 | expected_type 值 |
|----------|-----------------|
| 字串 | `string` |
| 數字 | `number` |
| 布林值 | `boolean` |
| 陣列 | `array` |
| 物件 | `object` |

---

## YAML 規則模板

### 基本規則結構

```yaml
id: xxx-001                    # 規則唯一識別碼（必填）
name: "規則名稱"                # 規則顯示名稱（必填）
enabled: true                  # 是否啟用（必填）
severity: error                # error/warning/info（必填）
description: "規則詳細說明"     # 規則描述（可選）

targets:                       # 適用目標（必填）
  file_patterns:               # 檔案匹配模式
    - "**/api*.yaml"
    - "**/config*.yaml"

rule:                          # 驗證邏輯（必填）
  type: rule_type              # 規則類型
  # ... 其他參數依規則類型而異
```

---

### 模板 1：必要欄位檢查

#### 單一必要欄位

```yaml
id: xxx-001
name: "檢查 xxx 欄位存在"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/*.yaml"

rule:
  type: required_field
  path: "config.section"
  message: "缺少 section 區塊"
```

#### 多個必要欄位

```yaml
id: xxx-002
name: "檢查多個必要欄位"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/*.yaml"

rule:
  type: required_fields
  path: "config"
  fields:
    - field1
    - field2
    - field3
  message: "config 缺少必要欄位"
```

---

### 模板 2：欄位型別檢查

```yaml
id: xxx-003
name: "檢查欄位型別"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/*.yaml"

rule:
  type: field_type
  path: "config.items"
  expected_type: array  # string, number, boolean, array, object
  message: "items 必須是陣列格式"
```

---

### 模板 3：數值範圍檢查

```yaml
id: xxx-004
name: "檢查數值範圍"
enabled: true
severity: warning

targets:
  file_patterns:
    - "**/*.yaml"

rule:
  type: value_range
  path: "config.timeout"
  min: 1000
  max: 30000
  message: "timeout 應在 1000-30000 ms 之間"
```

---

### 模板 4：陣列項目必要欄位

#### 一般陣列

```yaml
id: xxx-005
name: "陣列項目必要欄位"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/*.yaml"

rule:
  type: array_item_required_fields
  path: "config.items"
  required_fields:
    - id
    - name
    - type
  message: "每個項目必須包含 id, name, type"
```

#### 巢狀陣列（使用萬用字元）

```yaml
id: xxx-006
name: "巢狀陣列項目必要欄位"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/*.yaml"

rule:
  type: array_item_required_fields
  path: "config.routes[*].middlewares"  # [*] = 所有 routes
  required_fields:
    - name
    - priority
  message: "所有 middleware 必須包含 name 和 priority"
```

---

### 模板 5：陣列項目欄位值驗證

```yaml
id: xxx-007
name: "陣列項目欄位值驗證"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/*.yaml"

rule:
  type: array_item_field
  path: "config.items"
  field: "status"
  validation:
    type: enum
    allowed_values:
      - active
      - inactive
      - pending
  message: "status 必須是允許的值"
```

---

### 模板 6：正則表達式驗證

```yaml
id: xxx-008
name: "正則表達式驗證"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/*.yaml"

rule:
  type: pattern_match
  path: "config.email"
  pattern: '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
  message: "email 格式不正確"
```

**常用正則範例：**

```yaml
# 環境變數格式：${VAR_NAME}
pattern: '^\$\{[A-Z_]+\}$'

# API 路徑格式：/api/...
pattern: '^/api/[a-z0-9/-]+$'

# Email 格式
pattern: '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'

# IP 位址格式
pattern: '^(\d{1,3}\.){3}\d{1,3}$'

# URL 格式
pattern: '^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}'
```

---

### 模板 7：陣列欄位不重複檢查

#### 單一欄位不重複

```yaml
id: xxx-009
name: "陣列欄位不重複"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/*.yaml"

rule:
  type: array_no_duplicates
  path: "config.items"
  field: "id"
  message: "items 中的 id 欄位不可重複"
```

#### 多欄位組合不重複

```yaml
id: xxx-010
name: "陣列多欄位組合不重複"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/*.yaml"

rule:
  type: array_no_duplicates_combine
  path: "config.items"
  fields:
    - name
    - version
  message: "items 中的 name+version 組合不可重複"
```

#### 巢狀陣列不重複（使用萬用字元）

```yaml
id: xxx-011
name: "巢狀陣列欄位不重複"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/*.yaml"

rule:
  type: array_no_duplicates
  path: "config.routes[*].middlewares"  # 檢查每個 route 的 middlewares
  field: "name"
  message: "middleware 的 name 在同一個 route 中不可重複"
```

---

### 模板 8：SHA 雜湊值檢查

#### 弱密碼檢測（禁止模式）

```yaml
id: xxx-012
name: "弱密碼檢查"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/*.yaml"

rule:
  type: hashed_value_check
  path: "admin.password"
  hash_algorithm: "sha256"          # sha1, sha256, sha512, md5
  mode: "forbidden"                 # 禁止使用列表中的雜湊值
  hash_list:
    - "240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9"  # admin123
    - "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"  # password
  message: "密碼不可使用常見弱密碼"
```

#### 授權碼驗證（允許模式）

```yaml
id: xxx-013
name: "授權碼驗證"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/license*.yaml"

rule:
  type: hashed_value_check
  path: "license.key"
  hash_algorithm: "sha256"
  mode: "allowed"                   # 只允許列表中的雜湊值
  hash_list:
    - "abc123def456..."             # 合法授權碼的 hash
    - "xyz789uvw012..."
  message: "授權碼不正確"
```

**生成 SHA256 雜湊值：**
```bash
# 終端機
echo -n "your_password" | sha256sum

# Python
python3 -c "import hashlib; print(hashlib.sha256(b'your_password').hexdigest())"
```

---

### 模板 9：關鍵字檢查

#### 禁止敏感關鍵字（forbidden 模式）

```yaml
id: xxx-014
name: "敏感關鍵字檢查"
enabled: true
severity: warning

targets:
  file_patterns:
    - "**/*.yaml"

rule:
  type: contains_keywords
  path: "apiconfig.routes[*].path"  # 支援萬用字元
  mode: "forbidden"                 # 禁止包含關鍵字
  case_sensitive: false             # 不區分大小寫
  keywords:
    - "internal"
    - "private"
    - "admin"
    - "debug"
  message: "API 路徑不應包含敏感關鍵字"
```

#### 強制使用 HTTPS（required 模式）

```yaml
id: xxx-015
name: "強制使用 HTTPS"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: contains_keywords
  path: "api.baseUrl"
  mode: "required"                  # 必須包含關鍵字
  case_sensitive: true              # 區分大小寫
  keywords:
    - "https://"
  message: "API baseUrl 必須使用 HTTPS 協議"
```

---

### 模板 10：Trailing/Leading 空白檢查

```yaml
id: xxx-016
name: "欄位前後空白檢查"
enabled: true
severity: warning

targets:
  file_patterns:
    - "**/*.yaml"

rule:
  type: no_trailing_whitespace
  message: "配置檔中的字串欄位前後不可有空白字元"
```

**特色：**
- ⭐ **全檔自動掃描** - 不需要指定 `path`
- 自動檢查整個 YAML 檔案中的**所有字串欄位**
- 同時檢查空格和 Tab 字元

**錯誤訊息範例：**
```
⚠️  [xxx-016] 欄位前後空白檢查
   配置檔中的字串欄位前後不可有空白字元 (結尾有空白字元)
   路徑: apiconfig.routes[0].path

⚠️  [xxx-016] 欄位前後空白檢查
   配置檔中的字串欄位前後不可有空白字元 (開頭有空白字元)
   路徑: admin.username

⚠️  [xxx-016] 欄位前後空白檢查
   配置檔中的字串欄位前後不可有空白字元 (開頭和結尾有空白字元)
   路徑: database.description
```

---

## 進階功能

### 萬用字元路徑 `[*]`

使用 `[*]` 可以自動展開所有陣列項目，處理任意層級的巢狀陣列。

#### 語法說明

```
routes[*]                           → routes[0], routes[1], routes[2], ...
routes[*].middlewares               → 所有 route 的 middlewares
routes[*].middlewares[*].name       → 所有 route 的所有 middleware 的 name
databases[*].connections[*].host    → 兩層巢狀
```

#### 使用範例

```yaml
# 檢查所有 route 的所有 middlewares 的 priority 是否為數字
rule:
  type: field_type
  path: "apiconfig.routes[*].middlewares[*].priority"
  expected_type: number
  message: "priority 必須是數字"

# 檢查所有 connection 的 timeout 範圍
rule:
  type: value_range
  path: "databases[*].connections[*].timeout"
  min: 1000
  max: 30000
  message: "timeout 必須在 1000-30000 之間"
```

#### 混合使用

可以混合使用萬用字元 `[*]` 和特定索引 `[0]`：

```yaml
# 只檢查第一個 route 的所有 middlewares
path: "apiconfig.routes[0].middlewares"

# 檢查所有 route 的第一個 middleware
path: "apiconfig.routes[*].middlewares[0]"

# 檢查所有 route 的所有 middlewares（兩種寫法相同）
path: "apiconfig.routes[*].middlewares[*]"
path: "apiconfig.routes[*].middlewares"      # 最後一個可省略 [*]
```

### 深層巢狀範例

```yaml
# 四層巢狀結構
infrastructure:
  regions:
    - name: us-east-1
      clusters:
        - name: prod-cluster
          nodes:
            - name: node-1
              containers:
                - name: app
                  ports: [80, 443]

# 檢查所有 container 的必要欄位
rule:
  type: array_item_required_fields
  path: "infrastructure.regions[*].clusters[*].nodes[*].containers"
  required_fields: ["name", "ports"]
  message: "所有 container 都必須有 name 和 ports"
```

---

### 嚴重程度分級

| 級別 | 說明 | 使用時機 | 範例 |
|------|------|---------|------|
| **error** | 錯誤 | 配置錯誤會導致系統無法運作，必須修正 | 缺少必要欄位、資料型別錯誤 |
| **warning** | 警告 | 配置不理想但系統仍可運作，建議修正 | 數值超出建議範圍、命名不符合規範 |
| **info** | 資訊 | 提示性訊息，可選擇性修正 | 建議添加的欄位、優化建議 |

---

### 檔案匹配模式

使用 glob 模式匹配檔案：

```yaml
targets:
  file_patterns:
    # 精確匹配
    - "config.yaml"

    # 匹配檔名開頭
    - "api*.yaml"
    - "db*.yaml"

    # 匹配任意目錄
    - "**/api*.yaml"          # 任何目錄下的 api*.yaml
    - "**/config/*.yaml"      # 任何 config 目錄下的 .yaml

    # 匹配特定目錄
    - "configs/production/*.yaml"

    # 多個模式
    - "**/api*.yaml"
    - "**/routes*.yaml"
    - "**/gateway*.yaml"
```

---

## 使用建議

### 1. 規則命名規範

```
<產品代碼>-<編號>-<功能描述>.yaml

範例：
- api-001-required-fields.yaml
- api-002-routes-structure.yaml
- db-001-required-fields.yaml
- db-002-password-check.yaml
```

### 2. 規則組織策略

```
rules/
├── api/              # API 產品規則
│   ├── api-001-*.yaml
│   ├── api-002-*.yaml
│   └── ...
├── database/         # 資料庫規則
│   ├── db-001-*.yaml
│   └── db-002-*.yaml
├── frontend/         # 前端規則（如有需要）
└── shared/           # 共用規則（如有需要）
```

### 3. 規則 ID 編號建議

- **001-099**：基礎結構驗證（必要欄位、型別檢查）
- **100-199**：數值範圍、格式驗證
- **200-299**：業務邏輯驗證
- **300-399**：安全性檢查
- **900-999**：實驗性或臨時規則

### 4. 錯誤訊息撰寫原則

好的錯誤訊息應該：
- ✅ 清楚說明問題
- ✅ 提供修正方向
- ✅ 使用具體的欄位名稱

```yaml
# ✅ 好的訊息
message: "password 必須使用環境變數，格式: ${VAR_NAME}"
message: "每個 route 必須包含 path, method, handler 欄位"
message: "timeout 應在 1000-30000 ms 之間"

# ❌ 不好的訊息
message: "驗證失敗"
message: "格式錯誤"
message: "缺少欄位"
```

### 5. 優先使用萬用字元路徑

對於巢狀陣列驗證：

```yaml
# ✅ 推薦：使用萬用字元（更靈活）
rule:
  type: array_item_required_fields
  path: "apiconfig.routes[*].middlewares"
  required_fields: ["name", "priority"]

# ⚠️  可用但較複雜：使用 nested 規則
rule:
  type: nested_array_item_required_fields
  parent_path: "apiconfig.routes"
  child_path: "middlewares"
  required_fields: ["name", "priority"]
```

### 6. 測試規則的方法

```bash
# 測試單一檔案
./validator testdata/api-config.yaml

# 測試整個目錄
./validator testdata/valid

# 預期失敗的測試
./validator testdata/invalid

# JSON 輸出（方便程式化處理）
./validator --json testdata/ > report.json
```

### 7. 規則開發流程

1. **定義規則** - 建立 YAML 規則檔案
2. **建立測試資料** - 在 `testdata/` 建立測試檔案
3. **執行驗證** - 測試規則是否正確運作
4. **調整規則** - 根據結果調整規則內容
5. **文件更新** - 更新本參考手冊

---

## 相關文件

- **DUPLICATE_CHECK_GUIDE.md** - 陣列重複檢查與多層陣列存取指南
- **NESTED_ARRAY_GUIDE.md** - 巢狀陣列驗證完整指南
- **WILDCARD_PATH_GUIDE.md** - 萬用字元路徑完整指南
- **README.md** - 專案完整說明文件
- **QUICK_REFERENCE.md** - 快速參考指南

---

## 快速查詢

### 我想要...

| 需求 | 使用規則類型 | 參考章節 |
|------|------------|---------|
| 檢查某個欄位是否存在 | `required_field` | 模板 1 |
| 檢查多個欄位是否存在 | `required_fields` | 模板 1 |
| 檢查欄位型別 | `field_type` | 模板 2 |
| 檢查數值範圍 | `value_range` | 模板 3 |
| 檢查陣列每個項目的欄位 | `array_item_required_fields` | 模板 4 |
| 檢查陣列項目的值是否符合列舉 | `array_item_field` | 模板 5 |
| 檢查字串格式（如 email、URL） | `pattern_match` | 模板 6 |
| 檢查陣列中某欄位不重複 | `array_no_duplicates` | 模板 7 |
| 檢查陣列中多欄位組合不重複 | `array_no_duplicates_combine` | 模板 7 |
| 檢查巢狀陣列 | 使用 `[*]` 萬用字元 | 進階功能 |
| 檢測弱密碼或授權碼 | `hashed_value_check` | 模板 8 |
| 禁止或要求特定關鍵字 | `contains_keywords` | 模板 9 |
| 檢查字串前後是否有空白 | `no_trailing_whitespace` | 模板 10 |

---

**最後更新：** 2025-11-28
