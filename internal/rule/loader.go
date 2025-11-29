package rule

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Loader 規則載入器
type Loader struct {
	rulesDir string
}

// NewLoader 建立新的規則載入器
func NewLoader(rulesDir string) *Loader {
	return &Loader{
		rulesDir: rulesDir,
	}
}

// LoadRules 載入所有規則（遞迴載入子目錄）
func (l *Loader) LoadRules() ([]*ValidationRule, error) {
	var rules []*ValidationRule

	// 檢查規則目錄是否存在
	if _, err := os.Stat(l.rulesDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("規則目錄不存在: %s", l.rulesDir)
	}

	// 收集所有規則檔案路徑（遞迴）
	var ruleFiles []string
	err := filepath.Walk(l.rulesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳過目錄
		if info.IsDir() {
			return nil
		}

		// 只處理 YAML 檔案
		if strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml") {
			ruleFiles = append(ruleFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("掃描規則目錄失敗: %w", err)
	}

	// 按檔案路徑排序
	sort.Strings(ruleFiles)

	// 載入每個規則檔案
	for _, ruleFile := range ruleFiles {
		rule, err := l.loadRuleFile(ruleFile)
		if err != nil {
			return nil, fmt.Errorf("載入規則檔案 %s 失敗: %w", ruleFile, err)
		}

		// 只載入啟用的規則
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

// loadRuleFile 載入單個規則檔案
func (l *Loader) loadRuleFile(filePath string) (*ValidationRule, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("讀取檔案失敗: %w", err)
	}

	var rule ValidationRule
	if err := yaml.Unmarshal(content, &rule); err != nil {
		return nil, fmt.Errorf("解析 YAML 失敗: %w", err)
	}

	// 驗證必要欄位
	if err := l.validateRule(&rule); err != nil {
		return nil, err
	}

	return &rule, nil
}

// validateRule 驗證規則的完整性
func (l *Loader) validateRule(rule *ValidationRule) error {
	if rule.ID == "" {
		return fmt.Errorf("規則缺少 id 欄位")
	}
	if rule.Name == "" {
		return fmt.Errorf("規則 %s 缺少 name 欄位", rule.ID)
	}
	if rule.Severity != SeverityError && rule.Severity != SeverityWarning && rule.Severity != SeverityInfo {
		return fmt.Errorf("規則 %s 的 severity 必須是 error、warning 或 info", rule.ID)
	}
	if len(rule.Targets.FilePatterns) == 0 {
		return fmt.Errorf("規則 %s 缺少 file_patterns", rule.ID)
	}
	if rule.Rule.Type == "" {
		return fmt.Errorf("規則 %s 缺少 rule.type", rule.ID)
	}

	// 驗證規則類型的詳細配置
	if err := l.validateRuleDetails(rule); err != nil {
		return fmt.Errorf("規則 %s 配置錯誤: %w", rule.ID, err)
	}

	return nil
}

// validateRuleDetails 驗證規則類型的詳細配置
func (l *Loader) validateRuleDetails(rule *ValidationRule) error {
	switch rule.Rule.Type {
	case RuleTypeRequiredField:
		return validateRequiredFieldRule(rule.Rule.RawRule)
	case RuleTypeRequiredFields:
		return validateRequiredFieldsRule(rule.Rule.RawRule)
	case RuleTypeFieldType:
		return validateFieldTypeRule(rule.Rule.RawRule)
	case RuleTypeValueRange:
		return validateValueRangeRule(rule.Rule.RawRule)
	case RuleTypeArrayItemRequiredFields:
		return validateArrayItemRequiredFieldsRule(rule.Rule.RawRule)
	case RuleTypeArrayItemField:
		return validateArrayItemFieldRule(rule.Rule.RawRule)
	case RuleTypePatternMatch:
		return validatePatternMatchRule(rule.Rule.RawRule)
	case RuleTypeArrayNoDuplicates:
		return validateArrayNoDuplicatesRule(rule.Rule.RawRule)
	case RuleTypeArrayNoDuplicatesCombine:
		return validateArrayNoDuplicatesCombineRule(rule.Rule.RawRule)
	case RuleTypeHashedValueCheck:
		return validateHashedValueCheckRule(rule.Rule.RawRule)
	case RuleTypeContainsKeywords:
		return validateContainsKeywordsRule(rule.Rule.RawRule)
	case RuleTypeNoTrailingWhitespace:
		return validateNoTrailingWhitespaceRule(rule.Rule.RawRule)
	default:
		return fmt.Errorf("不支援的規則類型: %s", rule.Rule.Type)
	}
}

// validateRequiredFieldRule 驗證 required_field 規則
func validateRequiredFieldRule(rawRule map[string]interface{}) error {
	path, ok := rawRule["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("required_field 規則必須包含 path 欄位")
	}
	message, ok := rawRule["message"].(string)
	if !ok || message == "" {
		return fmt.Errorf("required_field 規則必須包含 message 欄位")
	}
	return nil
}

// validateRequiredFieldsRule 驗證 required_fields 規則
func validateRequiredFieldsRule(rawRule map[string]interface{}) error {
	path, ok := rawRule["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("required_fields 規則必須包含 path 欄位")
	}
	fields, ok := rawRule["fields"].([]interface{})
	if !ok || len(fields) == 0 {
		return fmt.Errorf("required_fields 規則必須包含非空的 fields 欄位")
	}
	message, ok := rawRule["message"].(string)
	if !ok || message == "" {
		return fmt.Errorf("required_fields 規則必須包含 message 欄位")
	}
	return nil
}

// validateFieldTypeRule 驗證 field_type 規則
func validateFieldTypeRule(rawRule map[string]interface{}) error {
	path, ok := rawRule["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("field_type 規則必須包含 path 欄位")
	}
	expectedType, ok := rawRule["expected_type"].(string)
	if !ok || expectedType == "" {
		return fmt.Errorf("field_type 規則必須包含 expected_type 欄位")
	}
	// 驗證 expected_type 是否合法
	validTypes := []string{"string", "number", "boolean", "array", "object"}
	valid := false
	for _, t := range validTypes {
		if expectedType == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("expected_type 必須是以下之一: %v", validTypes)
	}
	message, ok := rawRule["message"].(string)
	if !ok || message == "" {
		return fmt.Errorf("field_type 規則必須包含 message 欄位")
	}
	return nil
}

// validateValueRangeRule 驗證 value_range 規則
func validateValueRangeRule(rawRule map[string]interface{}) error {
	path, ok := rawRule["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("value_range 規則必須包含 path 欄位")
	}
	_, hasMin := rawRule["min"]
	_, hasMax := rawRule["max"]
	if !hasMin || !hasMax {
		return fmt.Errorf("value_range 規則必須包含 min 和 max 欄位")
	}
	message, ok := rawRule["message"].(string)
	if !ok || message == "" {
		return fmt.Errorf("value_range 規則必須包含 message 欄位")
	}
	return nil
}

// validateArrayItemRequiredFieldsRule 驗證 array_item_required_fields 規則
func validateArrayItemRequiredFieldsRule(rawRule map[string]interface{}) error {
	path, ok := rawRule["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("array_item_required_fields 規則必須包含 path 欄位")
	}
	fields, ok := rawRule["required_fields"].([]interface{})
	if !ok || len(fields) == 0 {
		return fmt.Errorf("array_item_required_fields 規則必須包含非空的 required_fields 欄位")
	}
	message, ok := rawRule["message"].(string)
	if !ok || message == "" {
		return fmt.Errorf("array_item_required_fields 規則必須包含 message 欄位")
	}
	return nil
}

// validateArrayItemFieldRule 驗證 array_item_field 規則
func validateArrayItemFieldRule(rawRule map[string]interface{}) error {
	path, ok := rawRule["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("array_item_field 規則必須包含 path 欄位")
	}
	field, ok := rawRule["field"].(string)
	if !ok || field == "" {
		return fmt.Errorf("array_item_field 規則必須包含 field 欄位")
	}
	_, ok = rawRule["validation"]
	if !ok {
		return fmt.Errorf("array_item_field 規則必須包含 validation 欄位")
	}
	message, ok := rawRule["message"].(string)
	if !ok || message == "" {
		return fmt.Errorf("array_item_field 規則必須包含 message 欄位")
	}
	return nil
}

// validatePatternMatchRule 驗證 pattern_match 規則
func validatePatternMatchRule(rawRule map[string]interface{}) error {
	path, ok := rawRule["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("pattern_match 規則必須包含 path 欄位")
	}
	pattern, ok := rawRule["pattern"].(string)
	if !ok || pattern == "" {
		return fmt.Errorf("pattern_match 規則必須包含 pattern 欄位")
	}
	// 驗證正則表達式是否有效
	if _, err := regexp.Compile(pattern); err != nil {
		return fmt.Errorf("pattern 正則表達式無效: %w", err)
	}
	message, ok := rawRule["message"].(string)
	if !ok || message == "" {
		return fmt.Errorf("pattern_match 規則必須包含 message 欄位")
	}
	return nil
}

// validateArrayNoDuplicatesRule 驗證 array_no_duplicates 規則
func validateArrayNoDuplicatesRule(rawRule map[string]interface{}) error {
	path, ok := rawRule["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("array_no_duplicates 規則必須包含 path 欄位")
	}
	field, ok := rawRule["field"].(string)
	if !ok || field == "" {
		return fmt.Errorf("array_no_duplicates 規則必須包含 field 欄位")
	}
	message, ok := rawRule["message"].(string)
	if !ok || message == "" {
		return fmt.Errorf("array_no_duplicates 規則必須包含 message 欄位")
	}
	return nil
}

// validateArrayNoDuplicatesCombineRule 驗證 array_no_duplicates_combine 規則
func validateArrayNoDuplicatesCombineRule(rawRule map[string]interface{}) error {
	path, ok := rawRule["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("array_no_duplicates_combine 規則必須包含 path 欄位")
	}
	fields, ok := rawRule["fields"].([]interface{})
	if !ok || len(fields) == 0 {
		return fmt.Errorf("array_no_duplicates_combine 規則必須包含非空的 fields 欄位")
	}
	message, ok := rawRule["message"].(string)
	if !ok || message == "" {
		return fmt.Errorf("array_no_duplicates_combine 規則必須包含 message 欄位")
	}
	return nil
}

// validateHashedValueCheckRule 驗證 hashed_value_check 規則
func validateHashedValueCheckRule(rawRule map[string]interface{}) error {
	path, ok := rawRule["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("hashed_value_check 規則必須包含 path 欄位")
	}
	hashAlgorithm, ok := rawRule["hash_algorithm"].(string)
	if !ok || hashAlgorithm == "" {
		return fmt.Errorf("hashed_value_check 規則必須包含 hash_algorithm 欄位")
	}
	// 驗證 hash_algorithm 是否合法
	validAlgos := []string{"sha1", "sha256", "sha512", "md5"}
	valid := false
	for _, algo := range validAlgos {
		if strings.ToLower(hashAlgorithm) == algo {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("hash_algorithm 必須是以下之一: %v", validAlgos)
	}
	mode, ok := rawRule["mode"].(string)
	if !ok || mode == "" {
		return fmt.Errorf("hashed_value_check 規則必須包含 mode 欄位")
	}
	if mode != "forbidden" && mode != "allowed" {
		return fmt.Errorf("mode 必須是 forbidden 或 allowed")
	}
	hashList, ok := rawRule["hash_list"].([]interface{})
	if !ok || len(hashList) == 0 {
		return fmt.Errorf("hashed_value_check 規則必須包含非空的 hash_list 欄位")
	}
	message, ok := rawRule["message"].(string)
	if !ok || message == "" {
		return fmt.Errorf("hashed_value_check 規則必須包含 message 欄位")
	}
	return nil
}

// validateContainsKeywordsRule 驗證 contains_keywords 規則
func validateContainsKeywordsRule(rawRule map[string]interface{}) error {
	path, ok := rawRule["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("contains_keywords 規則必須包含 path 欄位")
	}
	mode, ok := rawRule["mode"].(string)
	if !ok || mode == "" {
		return fmt.Errorf("contains_keywords 規則必須包含 mode 欄位")
	}
	if mode != "forbidden" && mode != "required" {
		return fmt.Errorf("mode 必須是 forbidden 或 required")
	}
	keywords, ok := rawRule["keywords"].([]interface{})
	if !ok || len(keywords) == 0 {
		return fmt.Errorf("contains_keywords 規則必須包含非空的 keywords 欄位")
	}
	message, ok := rawRule["message"].(string)
	if !ok || message == "" {
		return fmt.Errorf("contains_keywords 規則必須包含 message 欄位")
	}
	return nil
}

// validateNoTrailingWhitespaceRule 驗證 no_trailing_whitespace 規則
func validateNoTrailingWhitespaceRule(rawRule map[string]interface{}) error {
	message, ok := rawRule["message"].(string)
	if !ok || message == "" {
		return fmt.Errorf("no_trailing_whitespace 規則必須包含 message 欄位")
	}
	return nil
}

// MatchRules 根據檔案路徑匹配適用的規則
func MatchRules(rules []*ValidationRule, filePath string) []*ValidationRule {
	var matched []*ValidationRule

	fileName := filepath.Base(filePath)

	for _, rule := range rules {
		if matchFilePatterns(fileName, rule.Targets.FilePatterns) {
			matched = append(matched, rule)
		}
	}

	return matched
}

// matchFilePatterns 檢查檔案名稱是否匹配任一 pattern
func matchFilePatterns(fileName string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, fileName); matched {
			return true
		}

		// 支援 ** 通配符（簡化版本，只檢查檔案名稱）
		if strings.Contains(pattern, "**") {
			// 提取實際的檔案 pattern
			parts := strings.Split(pattern, "/")
			filePattern := parts[len(parts)-1]
			if matched, _ := filepath.Match(filePattern, fileName); matched {
				return true
			}
		}
	}

	return false
}
