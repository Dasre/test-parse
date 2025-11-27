#!/bin/bash
# Config Validator 專案驗證腳本
# 用途：快速驗證專案所有功能是否正常

set -e
cd "$(dirname "$0")"

echo "========================================"
echo "  Config Validator 專案驗證"
echo "========================================"
echo ""

# 1. 編譯
echo "→ 編譯專案..."
go build -o validator ./cmd/validator
echo "  ✓ 編譯成功"
echo ""

# 2. 測試有效配置
echo "→ 測試有效配置..."
./validator testdata/valid > /dev/null 2>&1
echo "  ✓ 有效配置驗證通過"

# 3. 測試無效配置
echo "→ 測試無效配置..."
if ./validator testdata/invalid > /dev/null 2>&1; then
  echo "  ✗ 無效配置應該要失敗"
  exit 1
fi
echo "  ✓ 無效配置正確檢測"

# 4. 測試混合產品
echo "→ 測試混合產品..."
./validator testdata/mixed > /dev/null 2>&1
echo "  ✓ 混合產品驗證通過"

# 5. 測試多路徑
echo "→ 測試多路徑支援..."
./validator testdata/valid testdata/mixed > /dev/null 2>&1
echo "  ✓ 多路徑驗證通過"

# 6. 測試多檔案
echo "→ 測試多檔案驗證..."
./validator testdata/valid/api-config.yaml testdata/valid/db-config.yaml > /dev/null 2>&1
echo "  ✓ 多檔案驗證通過"

# 7. 測試 JSON 輸出
echo "→ 測試 JSON 輸出..."
./validator --json testdata/valid > /tmp/test.json 2>&1
if ! grep -q '"total"' /tmp/test.json; then
  echo "  ✗ JSON 格式錯誤"
  exit 1
fi
echo "  ✓ JSON 輸出正確"

# 8. 統計資訊
echo ""
echo "========================================"
echo "  專案統計"
echo "========================================"
API_RULES=$(find rules/api -name "*.yaml" 2>/dev/null | wc -l)
DB_RULES=$(find rules/database -name "*.yaml" 2>/dev/null | wc -l)
echo "• API 規則: $API_RULES 條"
echo "• Database 規則: $DB_RULES 條"
echo "• 總計: $((API_RULES + DB_RULES)) 條規則"
echo ""

# 完成
echo "========================================"
echo "  ✅ 所有驗證通過！"
echo "========================================"
echo ""
echo "專案已就緒，可以開始使用："
echo ""
echo "  ./validator testdata/valid"
echo "  ./validator configs/"
echo "  ./validator --json configs/ > report.json"
echo ""
