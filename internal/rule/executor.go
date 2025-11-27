package rule

import (
	"config-validator/internal/parser"
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Executor 規則執行引擎
type Executor struct {
	parser *parser.YAMLParser
}

// NewExecutor 建立新的執行引擎
func NewExecutor(p *parser.YAMLParser) *Executor {
	return &Executor{
		parser: p,
	}
}

// Execute 執行規則驗證
func (e *Executor) Execute(rule *ValidationRule, filePath string) []*ValidationResult {
	switch rule.Rule.Type {
	case RuleTypeRequiredField:
		return e.executeRequiredField(rule, filePath)
	case RuleTypeRequiredFields:
		return e.executeRequiredFields(rule, filePath)
	case RuleTypeFieldType:
		return e.executeFieldType(rule, filePath)
	case RuleTypeValueRange:
		return e.executeValueRange(rule, filePath)
	case RuleTypeArrayItemRequiredFields:
		return e.executeArrayItemRequiredFields(rule, filePath)
	case RuleTypeArrayItemField:
		return e.executeArrayItemField(rule, filePath)
	case RuleTypePatternMatch:
		return e.executePatternMatch(rule, filePath)
	default:
		return []*ValidationResult{
			{
				File:     filePath,
				RuleID:   rule.ID,
				RuleName: rule.Name,
				Severity: SeverityError,
				Message:  fmt.Sprintf("不支援的規則類型: %s", rule.Rule.Type),
				Path:     "",
			},
		}
	}
}

// executeRequiredField 執行必要欄位檢查
func (e *Executor) executeRequiredField(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail RequiredFieldRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	if !e.parser.HasField(ruleDetail.Path) {
		return []*ValidationResult{
			{
				File:     filePath,
				RuleID:   rule.ID,
				RuleName: rule.Name,
				Severity: rule.Severity,
				Message:  ruleDetail.Message,
				Path:     ruleDetail.Path,
			},
		}
	}

	return nil
}

// executeRequiredFields 執行多個必要欄位檢查
func (e *Executor) executeRequiredFields(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail RequiredFieldsRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	// 先檢查父路徑是否存在
	if !e.parser.HasField(ruleDetail.Path) {
		return []*ValidationResult{
			{
				File:     filePath,
				RuleID:   rule.ID,
				RuleName: rule.Name,
				Severity: rule.Severity,
				Message:  ruleDetail.Message,
				Path:     ruleDetail.Path,
			},
		}
	}

	// 檢查每個必要欄位
	var results []*ValidationResult
	for _, field := range ruleDetail.Fields {
		fieldPath := ruleDetail.Path + "." + field
		if !e.parser.HasField(fieldPath) {
			results = append(results, &ValidationResult{
				File:     filePath,
				RuleID:   rule.ID,
				RuleName: rule.Name,
				Severity: rule.Severity,
				Message:  ruleDetail.Message,
				Path:     fieldPath,
			})
		}
	}

	return results
}

// executeFieldType 執行欄位類型檢查
func (e *Executor) executeFieldType(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail FieldTypeRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	if !e.parser.HasField(ruleDetail.Path) {
		return nil // 欄位不存在，不檢查類型
	}

	actualType := e.parser.GetType(ruleDetail.Path)
	expectedType := string(ruleDetail.ExpectedType)

	if actualType != expectedType {
		return []*ValidationResult{
			{
				File:     filePath,
				RuleID:   rule.ID,
				RuleName: rule.Name,
				Severity: rule.Severity,
				Message:  ruleDetail.Message,
				Path:     ruleDetail.Path,
			},
		}
	}

	return nil
}

// executeValueRange 執行數值範圍檢查
func (e *Executor) executeValueRange(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail ValueRangeRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	value, exists := e.parser.GetNumber(ruleDetail.Path)
	if !exists {
		return nil // 欄位不存在或不是數字，不檢查範圍
	}

	if value < ruleDetail.Min || value > ruleDetail.Max {
		return []*ValidationResult{
			{
				File:     filePath,
				RuleID:   rule.ID,
				RuleName: rule.Name,
				Severity: rule.Severity,
				Message:  ruleDetail.Message,
				Path:     ruleDetail.Path,
			},
		}
	}

	return nil
}

// executeArrayItemRequiredFields 執行陣列項目必要欄位檢查
func (e *Executor) executeArrayItemRequiredFields(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail ArrayItemRequiredFieldsRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	arr, exists := e.parser.GetArray(ruleDetail.Path)
	if !exists {
		return nil // 陣列不存在，不檢查
	}

	var results []*ValidationResult
	for i, item := range arr {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			// 嘗試轉換 map[interface{}]interface{}
			if m, ok2 := item.(map[interface{}]interface{}); ok2 {
				itemMap = make(map[string]interface{})
				for k, v := range m {
					if strKey, ok3 := k.(string); ok3 {
						itemMap[strKey] = v
					}
				}
			} else {
				continue
			}
		}

		// 檢查每個必要欄位
		for _, field := range ruleDetail.RequiredFields {
			if _, exists := itemMap[field]; !exists {
				results = append(results, &ValidationResult{
					File:     filePath,
					RuleID:   rule.ID,
					RuleName: rule.Name,
					Severity: rule.Severity,
					Message:  ruleDetail.Message,
					Path:     fmt.Sprintf("%s[%d].%s", ruleDetail.Path, i, field),
				})
			}
		}
	}

	return results
}

// executeArrayItemField 執行陣列項目欄位驗證
func (e *Executor) executeArrayItemField(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail ArrayItemFieldRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	arr, exists := e.parser.GetArray(ruleDetail.Path)
	if !exists {
		return nil // 陣列不存在，不檢查
	}

	var results []*ValidationResult
	for i, item := range arr {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			// 嘗試轉換 map[interface{}]interface{}
			if m, ok2 := item.(map[interface{}]interface{}); ok2 {
				itemMap = make(map[string]interface{})
				for k, v := range m {
					if strKey, ok3 := k.(string); ok3 {
						itemMap[strKey] = v
					}
				}
			} else {
				continue
			}
		}

		fieldValue, exists := itemMap[ruleDetail.Field]
		if !exists {
			continue
		}

		// 目前只支援 enum 驗證
		if ruleDetail.Validation.Type == "enum" {
			fieldStr, ok := fieldValue.(string)
			if !ok {
				continue
			}

			valid := false
			for _, allowed := range ruleDetail.Validation.AllowedValues {
				if fieldStr == allowed {
					valid = true
					break
				}
			}

			if !valid {
				results = append(results, &ValidationResult{
					File:     filePath,
					RuleID:   rule.ID,
					RuleName: rule.Name,
					Severity: rule.Severity,
					Message:  ruleDetail.Message,
					Path:     fmt.Sprintf("%s[%d].%s", ruleDetail.Path, i, ruleDetail.Field),
				})
			}
		}
	}

	return results
}

// executePatternMatch 執行正則表達式驗證
func (e *Executor) executePatternMatch(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail PatternMatchRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	value, exists := e.parser.GetString(ruleDetail.Path)
	if !exists {
		return nil // 欄位不存在或不是字串，不檢查
	}

	matched, err := regexp.MatchString(ruleDetail.Pattern, value)
	if err != nil {
		return makeErrorResult(rule, filePath, ruleDetail.Path, fmt.Sprintf("正則表達式錯誤: %v", err))
	}

	if !matched {
		return []*ValidationResult{
			{
				File:     filePath,
				RuleID:   rule.ID,
				RuleName: rule.Name,
				Severity: rule.Severity,
				Message:  ruleDetail.Message,
				Path:     ruleDetail.Path,
			},
		}
	}

	return nil
}

// unmarshalRule 將 RawRule 解析為具體的規則結構
func unmarshalRule(rawRule map[string]interface{}, target interface{}) error {
	// 將 map 轉換為 YAML，再解析回結構
	data, err := yaml.Marshal(rawRule)
	if err != nil {
		return fmt.Errorf("序列化規則失敗: %w", err)
	}

	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("解析規則失敗: %w", err)
	}

	return nil
}

// makeErrorResult 建立錯誤結果
func makeErrorResult(rule *ValidationRule, filePath, path, message string) []*ValidationResult {
	return []*ValidationResult{
		{
			File:     filePath,
			RuleID:   rule.ID,
			RuleName: rule.Name,
			Severity: SeverityError,
			Message:  message,
			Path:     path,
		},
	}
}
