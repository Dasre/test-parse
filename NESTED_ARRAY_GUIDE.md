# 巢狀陣列驗證完整指南

## 概述

本系統現在完整支援多層巢狀陣列的驗證,可以自動檢查父陣列中所有子陣列的內容。

## 支援的巢狀陣列規則類型

### 1. `nested_array_no_duplicates` - 巢狀陣列重複檢查

自動檢查父陣列中**每個項目**的子陣列是否有重複值。

**使用場景:**
檢查所有 routes 中的 middlewares 是否有重複的 name

**規則範例:**
```yaml
id: api-009
name: "每個 route 的 middlewares 中 name 不可重複"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: nested_array_no_duplicates
  parent_path: "apiconfig.routes"      # 父陣列路徑
  child_path: "middlewares"            # 子陣列欄位名稱
  field: "name"                        # 要檢查重複的欄位
  message: "middleware 的 name 在同一個 route 中不可重複"
```

**測試資料:**
```yaml
apiconfig:
  routes:
    - path: /api/users
      middlewares:
        - name: auth
        - name: logging
        - name: auth        # ❌ 重複！

    - path: /api/posts
      middlewares:
        - name: auth        # ✓ 不同 route,不算重複
        - name: cache
        - name: cache       # ❌ 重複！
```

**也支援多欄位組合檢查:**
```yaml
rule:
  type: nested_array_no_duplicates
  parent_path: "apiconfig.routes"
  child_path: "middlewares"
  fields:                             # 使用 fields 代替 field
    - name
    - version
  message: "middleware 的 name+version 組合不可重複"
```

---

### 2. `nested_array_item_required_fields` - 巢狀陣列必要欄位檢查

自動檢查父陣列中**每個項目**的子陣列項目是否有必要欄位。

**使用場景:**
確保所有 routes 的所有 middlewares 都有 name 和 priority 欄位

**規則範例:**
```yaml
id: api-010
name: "Middleware 必要欄位"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: nested_array_item_required_fields
  parent_path: "apiconfig.routes"
  child_path: "middlewares"
  required_fields:                    # 必要欄位列表
    - name
    - priority
  message: "每個 middleware 必須包含 name 和 priority 欄位"
```

**測試資料:**
```yaml
apiconfig:
  routes:
    - path: /api/users
      middlewares:
        - name: auth
          priority: 1                 # ✓ 都有
        - name: logging               # ❌ 缺少 priority
        - priority: 3                 # ❌ 缺少 name

    - path: /api/posts
      middlewares:
        - name: cache
          priority: 1                 # ✓ 都有
```

---

### 3. `nested_array_item_field` - 巢狀陣列欄位值驗證

自動驗證父陣列中**每個項目**的子陣列項目的欄位值。

**使用場景:**
檢查所有 routes 的所有 middlewares 的 type 欄位是否符合允許值

**規則範例:**
```yaml
id: api-011
name: "Middleware type 驗證"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: nested_array_item_field
  parent_path: "apiconfig.routes"
  child_path: "middlewares"
  field: "type"                       # 要驗證的欄位
  validation:
    type: enum
    allowed_values:
      - auth
      - logging
      - cache
      - ratelimit
  message: "middleware type 必須是允許的值"
```

**測試資料:**
```yaml
apiconfig:
  routes:
    - path: /api/users
      middlewares:
        - name: auth-middleware
          type: auth                  # ✓ 允許的值
        - name: log-middleware
          type: logging               # ✓ 允許的值

    - path: /api/posts
      middlewares:
        - name: custom-middleware
          type: custom                # ❌ 不在允許值中
```

---

## 對照:一般規則 vs 巢狀陣列規則

### 一般陣列規則 (只檢查指定層級)

```yaml
# 只檢查 routes 這一層
rule:
  type: array_item_required_fields
  path: "apiconfig.routes"
  required_fields:
    - path
    - method
```

### 巢狀陣列規則 (自動檢查所有子陣列)

```yaml
# 自動檢查所有 routes[*].middlewares[*]
rule:
  type: nested_array_item_required_fields
  parent_path: "apiconfig.routes"
  child_path: "middlewares"
  required_fields:
    - name
    - priority
```

---

## 完整範例

### YAML 配置檔案

```yaml
apiconfig:
  routes:
    - path: /api/users
      method: GET
      handler: getUsersHandler
      middlewares:
        - name: auth
          type: auth
          priority: 1
        - name: logging
          type: logging
          priority: 2

    - path: /api/posts
      method: GET
      handler: getPostsHandler
      middlewares:
        - name: cache
          type: cache
          priority: 1
        - name: auth
          type: auth
          priority: 2

  timeout: 5000
```

### 規則配置

```yaml
# 規則 1: 檢查 middlewares 中 name 不重複
- id: check-001
  rule:
    type: nested_array_no_duplicates
    parent_path: "apiconfig.routes"
    child_path: "middlewares"
    field: "name"
    message: "middleware name 在同一 route 中不可重複"

# 規則 2: 檢查 middlewares 必要欄位
- id: check-002
  rule:
    type: nested_array_item_required_fields
    parent_path: "apiconfig.routes"
    child_path: "middlewares"
    required_fields: ["name", "type", "priority"]
    message: "middleware 必須有 name, type, priority"

# 規則 3: 檢查 middlewares 的 type 值
- id: check-003
  rule:
    type: nested_array_item_field
    parent_path: "apiconfig.routes"
    child_path: "middlewares"
    field: "type"
    validation:
      type: enum
      allowed_values: ["auth", "logging", "cache"]
    message: "middleware type 必須是允許的值"
```

---

## 更複雜的巢狀結構

本系統也支援更深層的巢狀,例如:

```yaml
databases:
  - name: primary
    connections:
      - host: db1.example.com
        replicas:
          - host: replica1.example.com
          - host: replica2.example.com
```

可以使用:
```yaml
rule:
  type: nested_array_item_required_fields
  parent_path: "databases[0].connections"  # 可以指定父陣列的索引
  child_path: "replicas"
  required_fields: ["host"]
```

或者檢查所有 databases 的所有 connections:
```yaml
# 先檢查第一層
rule:
  type: nested_array_item_required_fields
  parent_path: "databases"
  child_path: "connections"
  required_fields: ["host", "replicas"]

# 再檢查第二層 (需要分別為每個 database 設定)
# 這部分可能需要多個規則或更進階的功能
```

---

## 測試指令

```bash
# 測試重複檢查
./validator testdata/api-nested-duplicate.yaml

# 測試必要欄位
./validator testdata/api-nested-missing-fields.yaml

# 測試所有功能
./validator testdata/
```

---

## 錯誤訊息格式

當發現問題時,錯誤訊息會清楚指出位置:

```
❌ [api-009] 每個 route 的 middlewares 中 name 不可重複
   middleware 的 name 在同一個 route 中不可重複 (重複值: auth)
   路徑: apiconfig.routes[0].middlewares[0].name

❌ [api-009] 每個 route 的 middlewares 中 name 不可重複
   middleware 的 name 在同一個 route 中不可重複 (重複值: auth)
   路徑: apiconfig.routes[0].middlewares[2].name
```

路徑格式:`parent_path[parent_index].child_path[child_index].field`
- `routes[0]` = 第一個 route
- `middlewares[2]` = 該 route 的第三個 middleware
- `name` = 重複的欄位

---

## 總結

現在支援的所有規則類型:

### 一般規則
- `required_field` - 單一必要欄位
- `required_fields` - 多個必要欄位
- `field_type` - 欄位類型檢查
- `value_range` - 數值範圍檢查
- `pattern_match` - 正則表達式匹配

### 陣列規則
- `array_item_required_fields` - 陣列項目必要欄位
- `array_item_field` - 陣列項目欄位驗證
- `array_no_duplicates` - 陣列欄位不重複
- `array_no_duplicates_combine` - 陣列多欄位組合不重複

### 巢狀陣列規則 (新增!)
- `nested_array_no_duplicates` - 巢狀陣列重複檢查
- `nested_array_item_required_fields` - 巢狀陣列項目必要欄位
- `nested_array_item_field` - 巢狀陣列項目欄位驗證
