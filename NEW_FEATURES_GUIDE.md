# 新增功能指南

## 概述

新增三種安全驗證規則類型，提升配置檔的安全性和資料品質：

1. **`hashed_value_check`** - SHA 雜湊值檢查（弱密碼檢測）
2. **`contains_keywords`** - 關鍵字檢查（敏感字詞檢測）
3. **`no_trailing_whitespace`** - Trailing/Leading 空白字元檢查

---

## 1️⃣ SHA 雜湊值檢查 (`hashed_value_check`)

### 功能說明

將配置檔中的字串欄位進行 SHA 雜湊運算，然後與預先定義的雜湊值列表比對，可用於：
- 檢測弱密碼（禁止使用常見密碼）
- 驗證授權碼（只允許特定的 hash 值）
- 敏感資料比對（不需明文儲存在規則中）

### 規則格式

```yaml
id: xxx-001
name: "規則名稱"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/*.yaml"

rule:
  type: hashed_value_check
  path: "admin.password"              # 要檢查的欄位路徑
  hash_algorithm: "sha256"            # 雜湊演算法：sha1, sha256, sha512, md5
  mode: "forbidden"                   # forbidden（禁止） 或 allowed（允許）
  hash_list:                          # 雜湊值列表
    - "240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9"  # admin123
    - "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"  # password
  message: "密碼不可使用常見弱密碼"
```

### 使用範例

#### 範例 1：禁止使用弱密碼

```yaml
id: sec-001
name: "弱密碼檢查"
enabled: true
severity: error

targets:
  file_patterns:
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
  message: "密碼不可使用常見弱密碼 (admin123, password, admin)"
```

**配置檔範例：**
```yaml
admin:
  username: admin
  password: admin123  # ❌ SHA256 在禁用列表中
```

#### 範例 2：驗證授權碼（允許模式）

```yaml
id: license-001
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
  mode: "allowed"                     # 只允許這些 hash
  hash_list:
    - "abc123def456..."               # 合法授權碼的 hash
    - "xyz789uvw012..."               # 另一個合法授權碼的 hash
  message: "授權碼不正確"
```

### 生成 SHA256 雜湊值

```bash
# 在終端機中生成密碼的 SHA256 hash
echo -n "admin123" | sha256sum

# 或使用 Python
python3 -c "import hashlib; print(hashlib.sha256(b'admin123').hexdigest())"
```

### 支援的雜湊演算法

- `sha1` - SHA-1（不建議用於安全場景）
- `sha256` - SHA-256（**推薦**）
- `sha512` - SHA-512
- `md5` - MD5（不建議用於安全場景）

---

## 2️⃣ 關鍵字檢查 (`contains_keywords`)

### 功能說明

檢查字串欄位是否包含（或必須包含）特定關鍵字，可用於：
- 禁止敏感關鍵字（internal、admin、debug）
- 強制使用特定格式（必須包含 https://）
- 過濾不當內容

**特色：**
- ✅ 支援萬用字元路徑 `[*]`
- ✅ 可設定是否區分大小寫
- ✅ 兩種模式：`forbidden`（禁止） / `required`（必須包含）

### 規則格式

```yaml
id: xxx-002
name: "規則名稱"
enabled: true
severity: warning

targets:
  file_patterns:
    - "**/*.yaml"

rule:
  type: contains_keywords
  path: "api.endpoint"                # 要檢查的欄位路徑
  mode: "forbidden"                   # forbidden（禁止） 或 required（必須包含）
  case_sensitive: false               # 是否區分大小寫
  keywords:                           # 關鍵字列表
    - "password"
    - "secret"
    - "internal"
  message: "API endpoint 不應包含敏感關鍵字"
```

### 使用範例

#### 範例 1：禁止 API 路徑包含敏感關鍵字

```yaml
id: api-012
name: "敏感關鍵字檢查"
enabled: true
severity: warning

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: contains_keywords
  path: "apiconfig.routes[*].path"    # 檢查所有 routes
  mode: "forbidden"
  case_sensitive: false
  keywords:
    - "internal"
    - "private"
    - "admin"
    - "debug"
    - "test"
  message: "API 路徑不應包含敏感關鍵字"
```

**配置檔範例：**
```yaml
apiconfig:
  routes:
    - path: /api/users           # ✓ 正常
      method: GET
    - path: /api/internal/debug  # ❌ 包含 "internal" 和 "debug"
      method: POST
```

#### 範例 2：強制使用 HTTPS（必須包含模式）

```yaml
id: api-013
name: "強制使用 HTTPS"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: contains_keywords
  path: "api.baseUrl"
  mode: "required"                    # 必須包含
  case_sensitive: true
  keywords:
    - "https://"
  message: "API baseUrl 必須使用 HTTPS 協議"
```

**配置檔範例：**
```yaml
api:
  baseUrl: "http://api.example.com"   # ❌ 不包含 "https://"
```

#### 範例 3：禁止敏感檔案路徑

```yaml
id: file-001
name: "敏感檔案路徑檢查"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/*.yaml"

rule:
  type: contains_keywords
  path: "config.logPath"
  mode: "forbidden"
  case_sensitive: true              # 區分大小寫
  keywords:
    - "/etc/passwd"
    - "/etc/shadow"
    - "C:\\Windows\\System32"
    - "/root/"
  message: "日誌路徑不應指向系統敏感目錄"
```

---

## 3️⃣ Trailing/Leading 空白字元檢查 (`no_trailing_whitespace`)

### 功能說明

自動掃描整個 YAML 檔案中的**所有字串欄位**，檢查前後是否有多餘的空白字元（空格、Tab），避免：
- 配置錯誤（路徑、名稱前後有空格）
- 資料品質問題
- 難以察覺的輸入錯誤

**特色：**
- ⭐ **全檔自動掃描** - 不需要指定 `path`，自動檢查所有字串值
- ✅ 涵蓋所有可能保留空白的類型（包括數字、布林值寫成字串形式）
- ✅ 自動偵測是開頭、結尾或兩者都有空白
- ✅ 同時檢查空格和 Tab

### 規則格式

```yaml
id: xxx-003
name: "規則名稱"
enabled: true
severity: warning

targets:
  file_patterns:
    - "**/*.yaml"

rule:
  type: no_trailing_whitespace
  message: "配置檔中的字串欄位前後不可有空白字元"  # 不需要 path 參數
```

### 使用範例

#### 範例 1：API 配置檔全檔檢查

```yaml
id: api-013
name: "欄位前後不可有空白"
enabled: true
severity: warning

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: no_trailing_whitespace
  message: "配置檔中的字串欄位前後不可有空白字元"
```

**配置檔範例：**
```yaml
apiconfig:
  routes:
    - path: "/api/users "         # ❌ 結尾有空白
      handler: getUsersHandler
    - path: /api/posts
      handler: " createPost "     # ❌ 前後都有空白
  timeout: 5000
  description: "	API Config	"  # ❌ 開頭和結尾有 Tab

admin:
  username: "admin "              # ❌ 結尾有空白
  password: MyStrongPassword      # ✓ 正常
  port: "5432 "                   # ❌ 數字寫成字串形式，結尾有空白
```

**錯誤訊息：**
```
⚠️  [api-013] 欄位前後不可有空白
   配置檔中的字串欄位前後不可有空白字元 (結尾有空白字元)
   路徑: apiconfig.routes[0].path

⚠️  [api-013] 欄位前後不可有空白
   配置檔中的字串欄位前後不可有空白字元 (開頭和結尾有空白字元)
   路徑: apiconfig.routes[1].handler

⚠️  [api-013] 欄位前後不可有空白
   配置檔中的字串欄位前後不可有空白字元 (開頭和結尾有空白字元)
   路徑: apiconfig.description

⚠️  [api-013] 欄位前後不可有空白
   配置檔中的字串欄位前後不可有空白字元 (結尾有空白字元)
   路徑: admin.username

⚠️  [api-013] 欄位前後不可有空白
   配置檔中的字串欄位前後不可有空白字元 (結尾有空白字元)
   路徑: admin.port
```

#### 範例 2：資料庫配置檔全檔檢查

```yaml
id: db-004
name: "欄位前後不可有空白"
enabled: true
severity: warning

targets:
  file_patterns:
    - "**/db*.yaml"

rule:
  type: no_trailing_whitespace
  message: "配置檔中的字串欄位前後不可有空白字元"
```

#### 範例 3：通用配置檔檢查

```yaml
id: general-001
name: "欄位前後不可有空白"
enabled: true
severity: warning

targets:
  file_patterns:
    - "**/*.yaml"  # 所有 YAML 檔案

rule:
  type: no_trailing_whitespace
  message: "配置檔中的字串欄位前後不可有空白字元"
```

---

## 完整使用範例

### 配置檔：`testdata/valid/api-security-good.yaml`

```yaml
apiconfig:
  routes:
    - path: /api/users
      method: GET
      handler: getUsersHandler      # ✓ 無空白
    - path: /api/posts
      method: POST
      handler: createPostHandler    # ✓ 無敏感關鍵字
  timeout: 5000

admin:
  username: admin
  password: MyStr0ngP@ssw0rd!2024   # ✓ 強密碼，hash 不在禁用列表
```

**驗證結果：**
```bash
$ ./validator testdata/valid/api-security-good.yaml
✅ 所有驗證通過
```

---

### 配置檔：`testdata/invalid/api-security-bad.yaml`

```yaml
apiconfig:
  routes:
    - path: /api/users
      method: GET
      handler: getUsersHandler
    - path: /api/internal/debug      # ❌ 包含敏感關鍵字
      method: POST
      handler: " debugHandler "      # ❌ 前後有空白
    - path: /api/admin/secret        # ❌ 包含敏感關鍵字
      method: DELETE
      handler: deleteSecretHandler
  timeout: 5000

admin:
  username: admin
  password: admin123                 # ❌ 弱密碼
```

**驗證結果：**
```bash
$ ./validator testdata/invalid/api-security-bad.yaml

📋 載入了 1 個產品的規則：
   • api: 12 條規則

📄 testdata/invalid/api-security-bad.yaml
  ❌ [api-011] 弱密碼檢查
     密碼不可使用常見弱密碼 (admin123, password, admin, qwerty 等)
     路徑: admin.password

  ⚠️  [api-012] 敏感關鍵字檢查
     API 路徑不應包含敏感關鍵字 (包含關鍵字: internal)
     路徑: apiconfig.routes[1].path

  ⚠️  [api-012] 敏感關鍵字檢查
     API 路徑不應包含敏感關鍵字 (包含關鍵字: admin)
     路徑: apiconfig.routes[2].path

  ⚠️  [api-013] 欄位前後不可有空白
     handler 名稱前後不可有空白字元 (開頭和結尾有空白字元)
     路徑: apiconfig.routes[1].handler

==================================================
❌ 1 個錯誤
⚠️  3 個警告
```

---

## 規則檔案位置

新增的規則檔案位於：

```
rules/api/
├── api-011-password-weak-check.yaml       # SHA 雜湊值檢查（弱密碼）
├── api-012-sensitive-keywords.yaml        # 關鍵字檢查（敏感字詞）
└── api-013-no-trailing-whitespace.yaml    # Trailing space 檢查
```

---

## 技術細節

### 萬用字元路徑支援

三個新規則都支援萬用字元路徑 `[*]`，可以自動檢查陣列中的所有項目：

```yaml
# 檢查所有 routes 的所有 middlewares 的 name
path: "apiconfig.routes[*].middlewares[*].name"

# 檢查所有 users 的 email
path: "users[*].email"

# 檢查所有 databases 的所有 connections 的 host
path: "databases[*].connections[*].host"
```

### 實作檔案

- **規則類型定義：** `internal/rule/types.go`
- **執行器實作：** `internal/rule/executor.go`
  - `executeHashedValueCheck()` - 雜湊值檢查
  - `executeContainsKeywords()` - 關鍵字檢查（支援萬用字元）
  - `executeNoTrailingWhitespace()` - 空白字元檢查（支援萬用字元）

---

## 常見問題

### Q1：如何生成密碼的 SHA256 雜湊值？

```bash
# 方法 1：使用 echo 和 sha256sum
echo -n "your_password" | sha256sum

# 方法 2：使用 Python
python3 -c "import hashlib; print(hashlib.sha256(b'your_password').hexdigest())"

# 方法 3：使用線上工具
# https://emn178.github.io/online-tools/sha256.html
```

### Q2：`forbidden` 和 `required` 模式的差異？

- **`forbidden` 模式**：不可包含列表中的任何項目
  - hash_list / keywords 中的值都是**禁止**的
  - 適用場景：弱密碼檢查、敏感關鍵字過濾

- **`required` 模式**：必須包含列表中的至少一個項目
  - hash_list / keywords 中的值是**允許**的
  - 適用場景：授權碼驗證、強制使用 HTTPS

### Q3：`case_sensitive` 是否區分大小寫？

- `case_sensitive: false`（預設）：不區分大小寫
  - `"Admin"` 會匹配關鍵字 `"admin"`

- `case_sensitive: true`：區分大小寫
  - `"Admin"` 不會匹配關鍵字 `"admin"`

### Q4：Trailing whitespace 檢查會檢查哪些字元？

- 空格（` `）
- Tab（`\t`）
- 同時檢查開頭（leading）和結尾（trailing）

### Q5：如何停用某個規則？

在規則檔案中設定 `enabled: false`：

```yaml
id: api-012
name: "敏感關鍵字檢查"
enabled: false    # 停用此規則
severity: warning
# ...
```

---

## 升級注意事項

### 向下相容性

✅ **完全相容** - 新增的規則不影響現有功能，現有配置檔和規則都能正常運作。

### 編譯要求

需要重新編譯驗證器：

```bash
go build -o validator ./cmd/validator
```

### 規則數量變化

- 原有規則：9 條（api-001 到 api-010，跳過 api-006）
- 新增規則：3 條（api-011, api-012, api-013）
- **總計：12 條規則**

---

## 總結

新增的三種規則類型大幅提升了配置檔的安全性和資料品質檢查能力：

| 規則類型 | 主要用途 | 特色 | 推薦嚴重程度 |
|---------|---------|------|------------|
| `hashed_value_check` | 弱密碼檢測、授權碼驗證 | 指定欄位檢查 | error |
| `contains_keywords` | 敏感字詞過濾、格式強制 | 支援萬用字元 `[*]` | warning/error |
| `no_trailing_whitespace` | 資料品質檢查 | ⭐ 全檔自動掃描 | warning |

**建議使用場景：**
- 🔒 所有涉及密碼的配置檔都應加上 `hashed_value_check`
- 🚫 所有 API 路徑都應加上 `contains_keywords` 過濾敏感字詞
- ✨ 所有配置檔都應加上 `no_trailing_whitespace` 確保資料品質（自動檢查所有字串欄位）
