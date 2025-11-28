# 陣列重複檢查與多層陣列存取指南

## 新增功能

### 1. 多層陣列存取

Parser 現在支援使用索引存取巢狀陣列:

```yaml
# 範例 YAML
apiconfig:
  routes:
    - path: /api/users
      middlewares:
        - name: auth
          priority: 1
```

可以使用以下路徑存取:
- `apiconfig.routes[0].path` → `/api/users`
- `apiconfig.routes[0].middlewares[0].name` → `auth`
- `apiconfig.routes[0].middlewares[0].priority` → `1`

### 2. 陣列欄位重複檢查

#### 2.1 單一欄位重複檢查 (`array_no_duplicates`)

檢查陣列中某個欄位是否有重複值。

**規則範例:**
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
  path: "apiconfig.routes"        # 陣列的路徑
  field: "path"                   # 要檢查的欄位
  message: "routes 中的 path 欄位不可重複"
```

**檢查結果:**
```yaml
routes:
  - path: /api/users    # ❌ 重複
    method: GET
  - path: /api/users    # ❌ 重複
    method: POST
  - path: /api/posts    # ✓ 沒有重複
    method: GET
```

#### 2.2 多欄位組合重複檢查 (`array_no_duplicates_combine`)

檢查陣列中多個欄位的組合是否重複。

**規則範例:**
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
  fields:                         # 要檢查組合的欄位列表
    - path
    - method
  message: "routes 中的 path+method 組合不可重複"
```

**檢查結果:**
```yaml
routes:
  - path: /api/users    # ✓ path+method = /api/users+GET (唯一)
    method: GET
  - path: /api/users    # ✓ path+method = /api/users+POST (唯一)
    method: POST
  - path: /api/users    # ❌ path+method = /api/users+GET (重複)
    method: GET
```

### 3. 巢狀陣列的重複檢查

雖然目前規則不直接支援巢狀陣列的重複檢查,但可以透過 Parser API 來實現:

```go
// 檢查 routes[2].middlewares 中 name 欄位的重複
duplicates, err := parser.CheckArrayDuplicates(
    "apiconfig.routes[2].middlewares",
    "name"
)
```

## 使用範例

### 範例 1: 檢查 API routes 的 path 不重複

```yaml
# rules/api/no-duplicate-path.yaml
id: api-001
name: "API path 不可重複"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/api*.yaml"

rule:
  type: array_no_duplicates
  path: "apiconfig.routes"
  field: "path"
  message: "API routes 中的 path 不可重複"
```

### 範例 2: 檢查資料庫連線名稱不重複

```yaml
# rules/database/no-duplicate-connection-name.yaml
id: db-001
name: "資料庫連線名稱不可重複"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/db*.yaml"

rule:
  type: array_no_duplicates
  path: "database.connections"
  field: "name"
  message: "資料庫連線名稱不可重複"
```

### 範例 3: 檢查使用者的 email+username 組合不重複

```yaml
# rules/users/no-duplicate-email-username.yaml
id: user-001
name: "使用者 email+username 組合不可重複"
enabled: true
severity: error

targets:
  file_patterns:
    - "**/users*.yaml"

rule:
  type: array_no_duplicates_combine
  path: "users"
  fields:
    - email
    - username
  message: "使用者的 email+username 組合不可重複"
```

## 測試

執行以下命令測試重複檢查功能:

```bash
# 測試有重複的檔案 (應該會報錯)
./validator testdata/api-test-duplicates.yaml

# 測試沒有重複的檔案 (應該通過)
./validator testdata/api-nested-arrays.yaml
```

## API 使用

### Parser API

```go
// 建立 parser
parser := parser.NewYAMLParser()
parser.ParseFile("config.yaml")

// 存取巢狀陣列
value, ok := parser.GetValue("routes[0].middlewares[1].name")

// 檢查單一欄位重複
duplicates, err := parser.CheckArrayDuplicates(
    "apiconfig.routes",  // 陣列路徑
    "path"               // 要檢查的欄位
)

// 檢查多欄位組合重複
duplicates, err := parser.CheckArrayMultiFieldDuplicates(
    "apiconfig.routes",           // 陣列路徑
    []string{"path", "method"}    // 要檢查組合的欄位
)

// 處理重複結果
for _, dup := range duplicates {
    fmt.Printf("重複值: %s, 出現在索引: %v\n", dup.Value, dup.Indices)
}
```

## 錯誤訊息格式

當發現重複時,錯誤訊息會包含:
- 重複的值
- 重複出現的所有索引位置

例如:
```
❌ [api-007] Route path 不可重複
   routes 中的 path 欄位不可重複 (重複值: /api/users)
   路徑: apiconfig.routes[0].path
❌ [api-007] Route path 不可重複
   routes 中的 path 欄位不可重複 (重複值: /api/users)
   路徑: apiconfig.routes[1].path
```
