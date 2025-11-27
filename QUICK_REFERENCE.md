# Config Validator - 快速參考

## 最常用指令

```bash
# 驗證單個目錄
./validator /path/to/configs

# 驗證多個目錄
./validator configs/dev configs/staging configs/prod

# 驗證特定檔案
./validator api.yaml db.yaml

# 混合驗證（目錄 + 檔案）
./validator configs/ extra.yaml

# JSON 輸出
./validator --json /path/to/configs

# 測試專案
./validator testdata/valid     # 應該通過
./validator testdata/invalid   # 應該失敗
./validator testdata/mixed     # 多產品測試

# 多路徑測試
./validator testdata/valid testdata/invalid
```

## 新增產品的 3 步驟

```bash
# 1. 建立規則目錄
mkdir -p rules/your-product

# 2. 添加規則檔案
cat > rules/your-product/rule-001.yaml <<EOF
id: your-001
name: "規則名稱"
enabled: true
severity: error
targets:
  file_patterns:
    - "**/your-product*.yaml"
rule:
  type: required_field
  path: "field.name"
  message: "錯誤訊息"
EOF

# 3. 註冊到 products.yaml
# 添加以下內容到 products.yaml:
# - name: your-product
#   description: "你的產品描述"
#   rules_dir: rules/your-product
#   path_patterns:
#     - "**/your-product*.yaml"
```

## 7 種規則類型速查

| 類型 | 用途 | 範例 |
|------|------|------|
| `required_field` | 檢查單一欄位存在 | `path: "apiconfig"` |
| `required_fields` | 檢查多個欄位存在 | `fields: [host, port]` |
| `field_type` | 檢查欄位類型 | `expected_type: array` |
| `value_range` | 檢查數值範圍 | `min: 10, max: 100` |
| `array_item_required_fields` | 檢查陣列項目欄位 | `required_fields: [path, method]` |
| `array_item_field` | 檢查陣列項目值 | `allowed_values: [GET, POST]` |
| `pattern_match` | 正則表達式驗證 | `pattern: '^\$\{.*\}$'` |

## Severity 等級

- **error** - 必須修正，否則系統無法運作
- **warning** - 建議修正，系統可運作但不理想
- **info** - 提示訊息，可選擇性修正

## 常見問題解決

**Q: 配置檔未被識別？**
```bash
# 檢查 products.yaml 中的 path_patterns
# 或重命名檔案符合模式
mv myconfig.yaml api-config.yaml
```

**Q: 如何只顯示錯誤？**
```bash
./validator --json configs/ | jq '.results[] | select(.severity == "error")'
```

**Q: 如何在 CI/CD 中使用？**
```bash
# GitLab CI / GitHub Actions
docker run --rm -v $(pwd)/configs:/configs:ro config-validator /configs
```

## 檔案結構

```
config-validator/
├── products.yaml              # 產品映射配置
├── rules/                     # 規則目錄
│   ├── api/                  # API 產品規則
│   ├── database/             # Database 產品規則
│   └── your-product/         # 你的產品規則
└── validator                  # 執行檔
```

## 實用技巧

### 1. 監控配置變更
```bash
watch -n 2 './validator configs/'
```

### 2. Git Pre-commit Hook
```bash
# .git/hooks/pre-commit
#!/bin/bash
./validator configs/ || exit 1
```

### 3. 產生詳細報告
```bash
./validator --json configs/ > report.json
jq '.results[] | {file, rule_id, severity, message}' report.json
```

### 4. 統計錯誤數量
```bash
./validator --json configs/ | jq '.results | group_by(.severity) | map({severity: .[0].severity, count: length})'
```

### 5. 多環境批次驗證
```bash
for env in dev staging prod; do
  echo "驗證 $env..."
  ./validator configs/$env || exit 1
done
```

## 需要更多協助？

請參閱完整文檔：[README.md](./README.md)
