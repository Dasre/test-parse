# 萬用字元路徑完整指南

## 重要更新!

現在**所有**陣列相關的規則都支援萬用字元路徑 `[*]`,可以處理任意層級的巢狀陣列!

## 萬用字元語法

使用 `[*]` 表示"所有陣列項目":

```
routes[*]              → routes[0], routes[1], routes[2], ...
routes[*].middlewares  → routes[0].middlewares, routes[1].middlewares, ...
routes[*].middlewares[*].name → 所有 route 的所有 middleware 的 name
databases[*].connections[*].replicas[*].host → 三層巢狀!
```

## 支援萬用字元的規則

### 1. `array_item_required_fields` - 陣列項目必要欄位

**單層陣列 (原有用法):**
```yaml
rule:
  type: array_item_required_fields
  path: "apiconfig.routes"
  required_fields: ["path", "method", "handler"]
```

**巢狀陣列 (新!使用萬用字元):**
```yaml
rule:
  type: array_item_required_fields
  path: "apiconfig.routes[*].middlewares"    # [*] = 所有 routes
  required_fields: ["name", "priority"]
  message: "所有 middleware 都必須有 name 和 priority"
```

**三層巢狀:**
```yaml
rule:
  type: array_item_required_fields
  path: "databases[*].connections[*].replicas"    # 兩個 [*]!
  required_fields: ["host", "port"]
```

### 2. `array_item_field` - 陣列項目欄位驗證

**檢查所有巢狀項目的欄位值:**
```yaml
rule:
  type: array_item_field
  path: "apiconfig.routes[*].middlewares"
  field: "type"
  validation:
    type: enum
    allowed_values: ["auth", "logging", "cache"]
  message: "middleware type 必須是允許的值"
```

### 3. `array_no_duplicates` - 陣列不重複檢查

**原有用法 (檢查單一陣列):**
```yaml
rule:
  type: array_no_duplicates
  path: "apiconfig.routes"
  field: "path"
  message: "route path 不可重複"
```

**注意:** `array_no_duplicates` 檢查的是**同一個陣列內**的重複。
如果使用 `path: "routes[*].middlewares"`,它會檢查:
- routes[0].middlewares 內的重複
- routes[1].middlewares 內的重複
- ... (每個 route 的 middlewares 各自檢查)

但**不會**跨 route 檢查重複 (這是正確的行為!)

### 4. `field_type`, `value_range`, `pattern_match` 等

這些規則也可以使用萬用字元路徑:

```yaml
# 檢查所有 middleware 的 priority 是否為數字
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

## 完整範例

### YAML 配置

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
```

### 規則配置

```yaml
# 規則 1: 檢查所有 route 的必要欄位
- id: rule-001
  rule:
    type: array_item_required_fields
    path: "apiconfig.routes"
    required_fields: ["path", "method", "handler"]

# 規則 2: 檢查所有 middleware 的必要欄位 (使用萬用字元!)
- id: rule-002
  rule:
    type: array_item_required_fields
    path: "apiconfig.routes[*].middlewares"
    required_fields: ["name", "type", "priority"]

# 規則 3: 檢查所有 middleware 的 type 值 (使用萬用字元!)
- id: rule-003
  rule:
    type: array_item_field
    path: "apiconfig.routes[*].middlewares"
    field: "type"
    validation:
      type: enum
      allowed_values: ["auth", "logging", "cache"]

# 規則 4: 檢查所有 middleware 的 priority 類型 (雙層萬用字元!)
- id: rule-004
  rule:
    type: field_type
    path: "apiconfig.routes[*].middlewares[*].priority"
    expected_type: number
```

## 對比:兩種寫法

### 方法 A: 使用 `nested_*` 規則 (較明確)

```yaml
rule:
  type: nested_array_item_required_fields
  parent_path: "apiconfig.routes"
  child_path: "middlewares"
  required_fields: ["name", "priority"]
```

優點:
- 明確表達父子關係
- 語意清楚

缺點:
- 只能處理兩層關係
- 需要特殊的規則類型

### 方法 B: 使用萬用字元路徑 (更靈活)

```yaml
rule:
  type: array_item_required_fields
  path: "apiconfig.routes[*].middlewares"
  required_fields: ["name", "priority"]
```

優點:
- 可以處理任意層級 (三層、四層都可以!)
- 使用現有的規則類型
- 更靈活,可以混合使用 `[*]` 和 `[0]`

缺點:
- 語法可能不太直觀

**建議:** 使用方法 B (萬用字元路徑),因為它更強大且靈活!

## 混合使用範例

可以混合使用萬用字元 `[*]` 和特定索引 `[0]`:

```yaml
# 只檢查第一個 route 的所有 middlewares
rule:
  type: array_item_required_fields
  path: "apiconfig.routes[0].middlewares"
  required_fields: ["name"]

# 檢查所有 route 的第一個 middleware
rule:
  type: array_item_required_fields
  path: "apiconfig.routes[*].middlewares[0]"  # 混合!
  required_fields: ["name"]

# 檢查所有 route 的所有 middlewares
rule:
  type: array_item_required_fields
  path: "apiconfig.routes[*].middlewares[*]"  # 兩個萬用字元
  required_fields: ["name"]

# 等價於:
rule:
  type: array_item_required_fields
  path: "apiconfig.routes[*].middlewares"     # 最後一個可以省略 [*]
  required_fields: ["name"]
```

## 深層巢狀範例

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
```

```yaml
# 檢查所有 container 的必要欄位
rule:
  type: array_item_required_fields
  path: "infrastructure.regions[*].clusters[*].nodes[*].containers"
  required_fields: ["name", "ports"]
  message: "所有 container 都必須有 name 和 ports"
```

## 效能考量

使用萬用字元路徑時,系統會展開所有匹配的路徑。例如:
- `routes[*]` 有 100 個 routes → 展開成 100 個路徑
- `routes[*].middlewares[*]` 每個 route 有 5 個 middlewares → 展開成 500 個路徑

這對於大型配置檔可能會影響效能,但一般使用場景都沒問題。

## 總結

✅ **建議使用萬用字元路徑 `[*]`** - 更強大、更靈活!

所有陣列規則都支援:
- `array_item_required_fields`
- `array_item_field`
- `array_no_duplicates`
- `array_no_duplicates_combine`
- `field_type`
- `value_range`
- `pattern_match`
- 等等...

可以處理任意層級的巢狀陣列!
