# 驗證規則完整參考手冊

> **最後更新：** 2025-11-29
> **系統版本：** v2.0 - 全功能通配符支持版本

---

## 📑 目錄

- [總覽](#總覽)
- [所有規則類型](#所有規則類型)
- [通配符支持說明](#通配符支持說明)
- [規則詳細說明](#規則詳細說明)
  - [基礎欄位檢查](#基礎欄位檢查)
  - [陣列項目檢查](#陣列項目檢查)
  - [重複值檢查](#重複值檢查)
  - [安全性檢查](#安全性檢查)
  - [資料品質檢查](#資料品質檢查)
- [規則撰寫範例](#規則撰寫範例)
- [最佳實踐](#最佳實踐)

---

## 總覽

本系統現在支持 **12 種驗證規則類型**，所有規則都經過以下改進：

### ✨ 功能亮點

1. **✅ 通配符全面支持** - 所有適用的規則類型都支持 `[*]` 通配符
2. **✅ 規則載入時驗證** - 規則寫錯會在載入時立即報錯
3. **✅ 詳細錯誤訊息** - 顯示實際值和期望值，便於快速修正
4. **✅ 簡化架構** - 移除冗餘功能，專注於配置驗證

### 規則分類

| 分類 | 規則數量 | 說明 |
|------|---------|------|
| 基礎欄位檢查 | 4 | required_field, required_fields, field_type, value_range |
| 陣列項目檢查 | 2 | array_item_required_fields, array_item_field |
| 重複值檢查 | 2 | array_no_duplicates, array_no_duplicates_combine |
| 安全性檢查 | 2 | hashed_value_check, contains_keywords |
| 資料品質檢查 | 2 | pattern_match, no_trailing_whitespace |

---

## 所有規則類型

### 完整列表

| # | 規則類型 | 通配符 | 主要用途 | Executor 函數 |
|---|---------|--------|---------|--------------|
| 1 | `required_field` | ✅ | 檢查單一必要欄位 | executeRequiredField |
| 2 | `required_fields` | - | 檢查多個必要欄位 | executeRequiredFields |
| 3 | `field_type` | ✅ | 檢查欄位類型 | executeFieldType |
| 4 | `value_range` | ✅ | 檢查數值範圍 | executeValueRange |
| 5 | `array_item_required_fields` | ✅ | 陣列項目必要欄位 | executeArrayItemRequiredFields |
| 6 | `array_item_field` | ✅ | 陣列項目欄位值驗證 | executeArrayItemField |
| 7 | `pattern_match` | ✅ | 正則表達式驗證 | executePatternMatch |
| 8 | `array_no_duplicates` | ✅ | 陣列欄位不重複 | executeArrayNoDuplicates |
| 9 | `array_no_duplicates_combine` | - | 多欄位組合不重複 | executeArrayNoDuplicatesCombine |
| 10 | `hashed_value_check` | - | SHA 雜湊值檢查 | executeHashedValueCheck |
| 11 | `contains_keywords` | ✅ | 關鍵字檢查 | executeContainsKeywords |
| 12 | `no_trailing_whitespace` | - | 空白字元檢查（全檔） | executeNoTrailingWhitespace |

---

## 通配符支持說明

### 什麼是通配符？

通配符 `[*]` 允許你在路徑中匹配所有陣列項目，無需逐一指定索引。

**範例：**
```yaml
# ❌ 舊方式：只檢查固定索引
path: "routes[0].timeout"  # 只檢查第一個 route

# ✅ 新方式：檢查所有項目
path: "routes[*].timeout"  # 檢查所有 routes
```

### 支持通配符的規則

以下規則類型**現已完全支持通配符**：

1. ✅ `required_field` - 可檢查 `routes[*].path` 是否存在
2. ✅ `field_type` - 可檢查 `routes[*].method` 的類型
3. ✅ `value_range` - 可檢查 `routes[*].timeout` 的範圍
4. ✅ `pattern_match` - 可檢查 `routes[*].path` 的格式
5. ✅ `array_item_required_fields` - 可使用 `routes[*].middlewares`
6. ✅ `contains_keywords` - 可檢查 `routes[*].description` 的關鍵字

### 多層通配符

支持任意層級的巢狀通配符：

```yaml
# 單層
routes[*].timeout

# 雙層
routes[*].middlewares[*].priority

# 三層
regions[*].clusters[*].nodes[*].cpu

# 混合使用
routes[0].middlewares[*].name  # 只檢查第一個 route 的所有 middlewares
routes[*].middlewares[0].name  # 檢查所有 routes 的第一個 middleware
```

---

## 規則詳細說明

### 基礎欄位檢查

#### 1. required_field

**功能：** 檢查單一必要欄位是否存在

**通配符支持：** ✅ 完全支持

**參數：**
| 參數 | 類型 | 必填 | 說明 |
|------|------|------|------|
| path | string | ✅ | 欄位路徑（支持通配符） |
| message | string | ✅ | 錯誤訊息 |

**使用範例：**

```yaml
# 基本用法
rule:
  type: required_field
  path: "apiconfig"
  message: "缺少 apiconfig 區塊"

# 使用通配符
rule:
  type: required_field
  path: "routes[*].method"
  message: "每個 route 都必須有 method 欄位"

# 多層通配符
rule:
  type: required_field
  path: "routes[*].middlewares[*].priority"
  message: "所有 middleware 都必須有 priority"
```

**驗證邏輯：**
- 檢查指定路徑的欄位是否存在
- 支持通配符，會展開所有陣列項目逐一檢查
- 欄位不存在時返回錯誤

---

#### 2. required_fields

**功能：** 檢查多個必要欄位

**通配符支持：** - （路徑不支持，但可檢查物件下的多個欄位）

**參數：**
| 參數 | 類型 | 必填 | 說明 |
|------|------|------|------|
| path | string | ✅ | 父路徑 |
| fields | []string | ✅ | 必要欄位列表 |
| message | string | ✅ | 錯誤訊息 |

**使用範例：**

```yaml
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

**驗證邏輯：**
- 先檢查父路徑是否存在
- 再檢查父路徑下的所有必要欄位
- 任一欄位缺失都會報錯

---

#### 3. field_type

**功能：** 檢查欄位類型

**通配符支持：** ✅ 完全支持

**參數：**
| 參數 | 類型 | 必填 | 說明 |
|------|------|------|------|
| path | string | ✅ | 欄位路徑（支持通配符） |
| expected_type | string | ✅ | 預期類型 |
| message | string | ✅ | 錯誤訊息 |

**支持的類型：**
- `string` - 字串
- `number` - 數字（int 或 float）
- `boolean` - 布林值
- `array` - 陣列
- `object` - 物件

**使用範例：**

```yaml
# 基本用法
rule:
  type: field_type
  path: "apiconfig.routes"
  expected_type: array
  message: "routes 必須是陣列格式"

# 使用通配符
rule:
  type: field_type
  path: "routes[*].timeout"
  expected_type: number
  message: "每個 route 的 timeout 必須是數字"

# 檢查巢狀結構
rule:
  type: field_type
  path: "routes[*].middlewares[*].priority"
  expected_type: number
  message: "所有 middleware 的 priority 必須是數字"
```

**驗證邏輯：**
- 獲取欄位值並判斷類型
- 與 expected_type 比對
- 類型不符時返回錯誤

---

#### 4. value_range

**功能：** 檢查數值範圍

**通配符支持：** ✅ 完全支持

**參數：**
| 參數 | 類型 | 必填 | 說明 |
|------|------|------|------|
| path | string | ✅ | 欄位路徑（支持通配符） |
| min | number | ✅ | 最小值（包含） |
| max | number | ✅ | 最大值（包含） |
| message | string | ✅ | 錯誤訊息 |

**使用範例：**

```yaml
# 基本用法
rule:
  type: value_range
  path: "apiconfig.timeout"
  min: 1000
  max: 30000
  message: "timeout 應在 1000-30000 ms 之間"

# 使用通配符檢查所有 routes
rule:
  type: value_range
  path: "routes[*].timeout"
  min: 1
  max: 100
  message: "每個 route 的 timeout 應在 1-100 秒之間"

# 檢查連接池大小
rule:
  type: value_range
  path: "databases[*].pool.maxConnections"
  min: 10
  max: 100
  message: "連接池大小應在 10-100 之間"
```

**驗證邏輯：**
- 獲取欄位值並轉換為數字
- 檢查是否在 [min, max] 範圍內
- 超出範圍時返回錯誤

---

### 陣列項目檢查

#### 5. array_item_required_fields

**功能：** 檢查陣列中每個項目的必要欄位

**通配符支持：** ✅ 完全支持

**參數：**
| 參數 | 類型 | 必填 | 說明 |
|------|------|------|------|
| path | string | ✅ | 陣列路徑（支持通配符） |
| required_fields | []string | ✅ | 必要欄位列表 |
| message | string | ✅ | 錯誤訊息 |

**使用範例：**

```yaml
# 基本用法
rule:
  type: array_item_required_fields
  path: "apiconfig.routes"
  required_fields:
    - path
    - method
    - handler
  message: "每個 route 必須包含 path, method, handler"

# 使用通配符檢查巢狀陣列
rule:
  type: array_item_required_fields
  path: "routes[*].middlewares"
  required_fields:
    - name
    - priority
  message: "每個 middleware 必須包含 name 和 priority"

# 多層巢狀
rule:
  type: array_item_required_fields
  path: "regions[*].clusters[*].nodes"
  required_fields:
    - name
    - ip
    - cpu
    - memory
  message: "每個 node 必須包含完整配置"
```

**驗證邏輯：**
- 獲取陣列中的每個項目
- 檢查每個項目是否包含所有必要欄位
- 缺少任一欄位時返回錯誤

---

#### 6. array_item_field

**功能：** 檢查陣列項目的欄位值（如枚舉驗證）

**通配符支持：** - （本身就是陣列操作）

**參數：**
| 參數 | 類型 | 必填 | 說明 |
|------|------|------|------|
| path | string | ✅ | 陣列路徑 |
| field | string | ✅ | 要檢查的欄位名 |
| validation | object | ✅ | 驗證規則 |
| message | string | ✅ | 錯誤訊息 |

**validation 物件：**
| 參數 | 類型 | 說明 |
|------|------|------|
| type | string | 驗證類型（目前只支持 "enum"） |
| allowed_values | []string | 允許的值列表 |

**使用範例：**

```yaml
# HTTP Method 驗證
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

# 狀態驗證
rule:
  type: array_item_field
  path: "services"
  field: "status"
  validation:
    type: enum
    allowed_values:
      - active
      - inactive
      - pending
  message: "status 必須是允許的值"
```

**驗證邏輯：**
- 遍歷陣列中的每個項目
- 檢查指定欄位的值是否在 allowed_values 中
- 不在列表中時返回錯誤

---

#### 7. pattern_match

**功能：** 正則表達式驗證

**通配符支持：** ✅ 完全支持

**參數：**
| 參數 | 類型 | 必填 | 說明 |
|------|------|------|------|
| path | string | ✅ | 欄位路徑（支持通配符） |
| pattern | string | ✅ | 正則表達式 |
| message | string | ✅ | 錯誤訊息 |

**使用範例：**

```yaml
# 環境變數格式檢查
rule:
  type: pattern_match
  path: "database.password"
  pattern: '^\$\{.*\}$'
  message: "password 必須使用環境變數，格式: ${VAR_NAME}"

# 使用通配符檢查所有 API 路徑格式
rule:
  type: pattern_match
  path: "routes[*].path"
  pattern: '^/api/[a-z0-9/-]+$'
  message: "API 路徑必須以 /api/ 開頭且只包含小寫字母、數字、斜線"

# Email 格式驗證
rule:
  type: pattern_match
  path: "users[*].email"
  pattern: '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
  message: "email 格式不正確"
```

**常用正則表達式：**

```yaml
# 環境變數：${VAR_NAME}
pattern: '^\$\{[A-Z_]+\}$'

# API 路徑：/api/xxx
pattern: '^/api/[a-z0-9/-]+$'

# IP 位址
pattern: '^(\d{1,3}\.){3}\d{1,3}$'

# URL (HTTP/HTTPS)
pattern: '^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}'

# UUID
pattern: '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
```

**驗證邏輯：**
- 獲取欄位值（字串類型）
- 使用正則表達式匹配
- 不匹配時返回錯誤
- **載入時會驗證正則表達式是否有效**

---

### 重複值檢查

#### 8. array_no_duplicates

**功能：** 檢查陣列中某個欄位的值不重複

**通配符支持：** - （用於巢狀場景，見下方說明）

**參數：**
| 參數 | 類型 | 必填 | 說明 |
|------|------|------|------|
| path | string | ✅ | 陣列路徑 |
| field | string | ✅ | 要檢查的欄位名 |
| message | string | ✅ | 錯誤訊息 |

**使用範例：**

```yaml
# 檢查 route path 不重複
rule:
  type: array_no_duplicates
  path: "apiconfig.routes"
  field: "path"
  message: "routes 中的 path 欄位不可重複"

# 檢查 ID 不重複
rule:
  type: array_no_duplicates
  path: "users"
  field: "id"
  message: "users 中的 id 不可重複"

# 巢狀陣列場景（每個 route 的 middlewares 中 name 不重複）
rule:
  type: array_no_duplicates
  path: "apiconfig.routes[*].middlewares"  # 使用通配符
  field: "name"
  message: "middleware 的 name 在同一個 route 中不可重複"
```

**驗證邏輯：**
- 遍歷陣列，收集所有項目的指定欄位值
- 找出重複的值及其索引
- 為每個重複項返回錯誤

---

#### 9. array_no_duplicates_combine

**功能：** 檢查陣列中多個欄位組合的值不重複

**通配符支持：** - （本身就是陣列操作）

**參數：**
| 參數 | 類型 | 必填 | 說明 |
|------|------|------|------|
| path | string | ✅ | 陣列路徑 |
| fields | []string | ✅ | 要組合檢查的欄位列表 |
| message | string | ✅ | 錯誤訊息 |

**使用範例：**

```yaml
# 檢查 path+method 組合不重複
rule:
  type: array_no_duplicates_combine
  path: "apiconfig.routes"
  fields:
    - path
    - method
  message: "routes 中的 path+method 組合不可重複"

# 檢查 name+version 組合不重複
rule:
  type: array_no_duplicates_combine
  path: "packages"
  fields:
    - name
    - version
  message: "packages 中的 name+version 組合不可重複"
```

**驗證邏輯：**
- 遍歷陣列，將指定欄位的值組合成字串（用 `|` 分隔）
- 找出重複的組合及其索引
- 為每個重複項返回錯誤

---

### 安全性檢查

#### 10. hashed_value_check

**功能：** SHA 雜湊值檢查（弱密碼檢測、授權碼驗證）

**通配符支持：** - （通常用於單一欄位）

**參數：**
| 參數 | 類型 | 必填 | 說明 |
|------|------|------|------|
| path | string | ✅ | 欄位路徑 |
| hash_algorithm | string | ✅ | 雜湊演算法 |
| mode | string | ✅ | 模式 |
| hash_list | []string | ✅ | 雜湊值列表 |
| message | string | ✅ | 錯誤訊息 |

**hash_algorithm 選項：**
- `sha1` - SHA-1
- `sha256` - SHA-256（推薦）
- `sha512` - SHA-512
- `md5` - MD5（不推薦用於安全場景）

**mode 選項：**
- `forbidden` - 禁止使用列表中的雜湊值（用於弱密碼檢測）
- `allowed` - 只允許列表中的雜湊值（用於授權碼驗證）

**使用範例：**

```yaml
# 弱密碼檢測
rule:
  type: hashed_value_check
  path: "admin.password"
  hash_algorithm: "sha256"
  mode: "forbidden"
  hash_list:
    - "240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9"  # admin123
    - "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"  # password
    - "8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918"  # admin
  message: "密碼不可使用常見弱密碼"

# 授權碼驗證
rule:
  type: hashed_value_check
  path: "license.key"
  hash_algorithm: "sha256"
  mode: "allowed"
  hash_list:
    - "abc123def456..."  # 合法授權碼 1
    - "xyz789uvw012..."  # 合法授權碼 2
  message: "授權碼不正確"
```

**生成雜湊值工具：**

```bash
# Bash
echo -n "your_password" | sha256sum

# Python
python3 -c "import hashlib; print(hashlib.sha256(b'your_password').hexdigest())"

# Node.js
node -e "console.log(require('crypto').createHash('sha256').update('your_password').digest('hex'))"
```

**驗證邏輯：**
- 獲取欄位值並計算雜湊
- 檢查雜湊值是否在 hash_list 中
- 根據 mode 判斷是否違規

---

#### 11. contains_keywords

**功能：** 關鍵字檢查（敏感字詞過濾、格式強制）

**通配符支持：** ✅ 完全支持

**參數：**
| 參數 | 類型 | 必填 | 說明 |
|------|------|------|------|
| path | string | ✅ | 欄位路徑（支持通配符） |
| mode | string | ✅ | 模式 |
| case_sensitive | bool | - | 是否區分大小寫（預設 false） |
| keywords | []string | ✅ | 關鍵字列表 |
| message | string | ✅ | 錯誤訊息 |

**mode 選項：**
- `forbidden` - 不可包含任何關鍵字（敏感字詞過濾）
- `required` - 必須包含至少一個關鍵字（格式強制）

**使用範例：**

```yaml
# 禁止敏感關鍵字（使用通配符）
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
  message: "API 路徑不應包含敏感關鍵字"

# 強制使用 HTTPS
rule:
  type: contains_keywords
  path: "api.baseUrl"
  mode: "required"
  case_sensitive: true
  keywords:
    - "https://"
  message: "API baseUrl 必須使用 HTTPS 協議"

# 檢查所有描述是否包含必要標籤
rule:
  type: contains_keywords
  path: "services[*].description"
  mode: "required"
  case_sensitive: false
  keywords:
    - "[stable]"
    - "[beta]"
    - "[alpha]"
  message: "服務描述必須包含版本標籤"
```

**驗證邏輯：**
- 獲取欄位值（字串類型）
- 根據 case_sensitive 處理大小寫
- 檢查是否包含關鍵字
- 根據 mode 判斷是否違規

---

### 資料品質檢查

#### 12. no_trailing_whitespace

**功能：** 檢查字串欄位前後是否有空白字元（全檔自動掃描）

**通配符支持：** - （自動掃描整個檔案）

**參數：**
| 參數 | 類型 | 必填 | 說明 |
|------|------|------|------|
| message | string | ✅ | 錯誤訊息 |

**特色：**
- ⭐ **全檔自動掃描** - 不需要指定 path
- 自動檢查整個 YAML 檔案中的**所有字串欄位**
- 同時檢查空格和 Tab 字元
- 檢查開頭（leading）和結尾（trailing）空白

**使用範例：**

```yaml
rule:
  type: no_trailing_whitespace
  message: "配置檔中的字串欄位前後不可有空白字元"
```

**錯誤訊息範例：**

```
⚠️  [api-013] 欄位前後空白檢查
   配置檔中的字串欄位前後不可有空白字元 (結尾有空白字元)
   路徑: apiconfig.routes[0].path

⚠️  [api-013] 欄位前後空白檢查
   配置檔中的字串欄位前後不可有空白字元 (開頭有空白字元)
   路徑: admin.username

⚠️  [api-013] 欄位前後空白檢查
   配置檔中的字串欄位前後不可有空白字元 (開頭和結尾有空白字元)
   路徑: database.description
```

**驗證邏輯：**
- 遞迴掃描整個 YAML 資料結構
- 對所有字串值進行 trim 檢查
- 發現空白時返回詳細的錯誤位置

---

## 規則撰寫範例

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

### 範例 1：檢查所有 routes 的 timeout

```yaml
id: api-timeout
name: "Route Timeout 檢查"
enabled: true
severity: warning

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: value_range
  path: "routes[*].timeout"  # 使用通配符
  min: 1
  max: 100
  message: "每個 route 的 timeout 應在 1-100 秒之間"
```

### 範例 2：檢查所有 middlewares 的必要欄位

```yaml
id: api-middleware
name: "Middleware 必要欄位"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: array_item_required_fields
  path: "routes[*].middlewares"  # 通配符展開所有 routes
  required_fields:
    - name
    - priority
    - enabled
  message: "所有 middleware 必須包含 name, priority, enabled"
```

### 範例 3：檢查 API 路徑格式

```yaml
id: api-path-format
name: "API 路徑格式檢查"
enabled: true
severity: warning

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: pattern_match
  path: "routes[*].path"  # 檢查所有 routes 的 path
  pattern: '^/api/v\d+/[a-z0-9/-]+$'
  message: "API 路徑必須符合格式: /api/v1/resource"
```

### 範例 4：多層巢狀檢查

```yaml
id: infra-cpu-limit
name: "Container CPU 限制檢查"
enabled: true
severity: warning

targets:
  file_patterns:
    - "**/infrastructure*.yaml"

rule:
  type: value_range
  path: "regions[*].clusters[*].nodes[*].containers[*].cpu"
  min: 0.1
  max: 8.0
  message: "Container CPU 應在 0.1-8.0 核之間"
```

---

## 最佳實踐

### 1. 規則命名規範

```
<產品代碼>-<編號>-<功能描述>.yaml

範例：
- api-001-required-fields.yaml
- api-002-routes-structure.yaml
- db-001-connection-check.yaml
- fe-001-theme-validation.yaml
```

### 2. 規則 ID 編號建議

- **001-099**：基礎結構驗證（必要欄位、型別檢查）
- **100-199**：數值範圍、格式驗證
- **200-299**：業務邏輯驗證
- **300-399**：安全性檢查
- **900-999**：實驗性或臨時規則

### 3. 善用通配符

對於陣列相關的檢查，優先使用通配符路徑：

```yaml
# ✅ 推薦：使用通配符（簡潔、直觀）
rule:
  type: value_range
  path: "routes[*].timeout"
  min: 1
  max: 100

# ⚠️ 不推薦：使用 nested 規則（複雜）
# 除非有特殊需求才使用
```

### 4. 錯誤訊息要清楚

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

### 5. 嚴重程度分級

| 級別 | 使用時機 | 範例 |
|------|---------|------|
| **error** | 配置錯誤會導致系統無法運作 | 缺少必要欄位、資料型別錯誤 |
| **warning** | 配置不理想但系統仍可運作 | 數值超出建議範圍、命名不符規範 |
| **info** | 提示性訊息 | 建議添加的欄位、優化建議 |

### 6. 測試規則

在部署規則前，務必測試：

```bash
# 1. 測試有效配置（應該通過）
go run cmd/validator/main.go testdata/valid/

# 2. 測試無效配置（應該報錯）
go run cmd/validator/main.go testdata/invalid/

# 3. 查看 JSON 輸出
go run cmd/validator/main.go --json testdata/ > report.json
```

### 7. 規則載入時驗證

系統會在載入規則時自動驗證：

**會檢查的內容：**
- ✅ 必要欄位是否完整
- ✅ severity 是否合法（error/warning/info）
- ✅ expected_type 是否合法（string/number/boolean/array/object）
- ✅ 正則表達式是否有效
- ✅ hash_algorithm 是否合法（sha1/sha256/sha512/md5）
- ✅ mode 是否合法（forbidden/allowed 或 required）

**範例錯誤訊息：**

```
❌ 載入規則失敗: 規則 api-003 配置錯誤: pattern 正則表達式無效: error parsing regexp: missing closing ): `^(abc$`
```

---

## 快速查詢

### 我想要...

| 需求 | 使用規則類型 | 通配符 |
|------|------------|--------|
| 檢查某個欄位是否存在 | `required_field` | ✅ |
| 檢查多個欄位是否存在 | `required_fields` | - |
| 檢查欄位類型 | `field_type` | ✅ |
| 檢查數值範圍 | `value_range` | ✅ |
| 檢查陣列每個項目的欄位 | `array_item_required_fields` | ✅ |
| 檢查陣列項目的值是否符合列舉 | `array_item_field` | - |
| 檢查字串格式（email、URL） | `pattern_match` | ✅ |
| 檢查陣列中某欄位不重複 | `array_no_duplicates` | - |
| 檢查陣列中多欄位組合不重複 | `array_no_duplicates_combine` | - |
| 檢查巢狀陣列 | 使用 `[*]` 通配符 | ✅ |
| 檢測弱密碼 | `hashed_value_check` | - |
| 禁止或要求特定關鍵字 | `contains_keywords` | ✅ |
| 檢查字串前後空白 | `no_trailing_whitespace` | - |

---

## 相關指令

```bash
# 驗證配置檔
validator <path>

# 驗證多個路徑
validator <path1> <path2> <path3>

# JSON 輸出
validator --json <path>
```

---

**維護者註記：**
- 所有 executor 函數位於 `internal/rule/executor.go`
- 所有規則類型定義位於 `internal/rule/types.go`
- 規則驗證函數位於 `internal/rule/loader.go`
