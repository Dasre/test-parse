package rule

import (
	"config-validator/internal/parser"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"regexp"
	"strings"

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
	case RuleTypeArrayNoDuplicates:
		return e.executeArrayNoDuplicates(rule, filePath)
	case RuleTypeArrayNoDuplicatesCombine:
		return e.executeArrayNoDuplicatesCombine(rule, filePath)
	case RuleTypeHashedValueCheck:
		return e.executeHashedValueCheck(rule, filePath)
	case RuleTypeContainsKeywords:
		return e.executeContainsKeywords(rule, filePath)
	case RuleTypeNoTrailingWhitespace:
		return e.executeNoTrailingWhitespace(rule, filePath)
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
// 支援萬用字元，例如 routes[*].path 會檢查每個 routes 項目是否都有 path 欄位
func (e *Executor) executeRequiredField(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail RequiredFieldRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	// 檢查是否包含萬用字元
	if strings.Contains(ruleDetail.Path, "[*]") {
		// 展開萬用字元路徑，檢查每個路徑是否存在
		return e.checkRequiredFieldsWithWildcard(ruleDetail.Path, rule, filePath, ruleDetail.Message)
	}

	// 單一路徑檢查
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

// checkRequiredFieldsWithWildcard 檢查萬用字元路徑中的必要欄位
// 例如: routes[*].path 會展開並檢查 routes[0].path, routes[1].path, ...
func (e *Executor) checkRequiredFieldsWithWildcard(path string, rule *ValidationRule, filePath, message string) []*ValidationResult {
	// 分析路徑，找出萬用字元的位置
	// 例如: routes[*].middlewares[*].name -> 需要處理兩層萬用字元
	parts := strings.Split(path, ".")

	// 找到第一個包含 [*] 的部分
	var parentPath string
	var fieldName string
	var remainingPath string

	for i, part := range parts {
		if strings.Contains(part, "[*]") {
			// 找到萬用字元
			if i > 0 {
				parentPath = strings.Join(parts[:i], ".")
			}
			fieldName = strings.TrimSuffix(part, "[*]")
			if i < len(parts)-1 {
				remainingPath = strings.Join(parts[i+1:], ".")
			}
			break
		}
	}

	// 獲取父陣列
	var parentArray []interface{}
	var exists bool

	if parentPath == "" {
		// 萬用字元在根層級，如 routes[*]
		parentArray, exists = e.parser.GetArray(fieldName)
	} else {
		// 萬用字元在嵌套路徑中，如 apiconfig.routes[*]
		fullArrayPath := parentPath + "." + fieldName
		parentArray, exists = e.parser.GetArray(fullArrayPath)
	}

	if !exists {
		// 父陣列不存在，返回 nil（不算錯誤）
		return nil
	}

	var results []*ValidationResult

	// 檢查每個陣列項目
	for i := range parentArray {
		var checkPath string
		if parentPath == "" {
			checkPath = fmt.Sprintf("%s[%d]", fieldName, i)
		} else {
			checkPath = fmt.Sprintf("%s.%s[%d]", parentPath, fieldName, i)
		}

		if remainingPath != "" {
			checkPath = checkPath + "." + remainingPath
		}

		// 遞迴檢查（處理多層萬用字元的情況）
		if strings.Contains(checkPath, "[*]") {
			results = append(results, e.checkRequiredFieldsWithWildcard(checkPath, rule, filePath, message)...)
		} else {
			// 檢查該路徑是否存在
			if !e.parser.HasField(checkPath) {
				results = append(results, &ValidationResult{
					File:     filePath,
					RuleID:   rule.ID,
					RuleName: rule.Name,
					Severity: rule.Severity,
					Message:  message,
					Path:     checkPath,
				})
			}
		}
	}

	return results
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
// 支援萬用字元，例如 routes[*].timeout 會檢查每個 routes 項目的 timeout 類型
func (e *Executor) executeFieldType(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail FieldTypeRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	expectedType := string(ruleDetail.ExpectedType)

	return e.processPathWithWildcard(ruleDetail.Path, func(actualPath string, value interface{}) *ValidationResult {
		// 檢查類型
		var actualType string
		switch value.(type) {
		case string:
			actualType = "string"
		case int, float64:
			actualType = "number"
		case bool:
			actualType = "boolean"
		case []interface{}:
			actualType = "array"
		case map[string]interface{}, map[interface{}]interface{}:
			actualType = "object"
		default:
			actualType = "unknown"
		}

		if actualType != expectedType {
			return &ValidationResult{
				File:          filePath,
				RuleID:        rule.ID,
				RuleName:      rule.Name,
				Severity:      rule.Severity,
				Message:       ruleDetail.Message,
				Path:          actualPath,
				ActualValue:   actualType,
				ExpectedValue: expectedType,
			}
		}
		return nil
	})
}

// executeValueRange 執行數值範圍檢查
// 支援萬用字元，例如 routes[*].timeout 會檢查每個 routes 項目的 timeout 範圍
func (e *Executor) executeValueRange(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail ValueRangeRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	return e.processPathWithWildcard(ruleDetail.Path, func(actualPath string, value interface{}) *ValidationResult {
		// 轉換為數字
		var numValue float64
		switch v := value.(type) {
		case int:
			numValue = float64(v)
		case float64:
			numValue = v
		default:
			// 不是數字，跳過檢查
			return nil
		}

		// 檢查範圍
		if numValue < ruleDetail.Min || numValue > ruleDetail.Max {
			return &ValidationResult{
				File:          filePath,
				RuleID:        rule.ID,
				RuleName:      rule.Name,
				Severity:      rule.Severity,
				Message:       ruleDetail.Message,
				Path:          actualPath,
				ActualValue:   fmt.Sprintf("%.0f", numValue),
				ExpectedValue: fmt.Sprintf("%.0f - %.0f", ruleDetail.Min, ruleDetail.Max),
			}
		}
		return nil
	})
}

// executeArrayItemRequiredFields 執行陣列項目必要欄位檢查
// 支援萬用字元路徑 [*]，例如: "routes[*].middlewares"
func (e *Executor) executeArrayItemRequiredFields(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail ArrayItemRequiredFieldsRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	// 檢查路徑是否包含萬用字元
	if strings.Contains(ruleDetail.Path, "[*]") {
		// 展開萬用字元路徑
		paths := e.parser.ExpandWildcardPath(ruleDetail.Path)
		if paths == nil || len(paths) == 0 {
			return nil
		}

		var results []*ValidationResult
		for _, pathInfo := range paths {
			// pathInfo.Value 應該是一個陣列
			arr, ok := pathInfo.Value.([]interface{})
			if !ok {
				continue
			}

			// 檢查陣列中的每個項目
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
							Path:     fmt.Sprintf("%s[%d].%s", pathInfo.Path, i, field),
						})
					}
				}
			}
		}
		return results
	}

	// 原有的非萬用字元邏輯
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
// 支援萬用字元，例如 routes[*].middlewares 會檢查所有 route 的 middlewares
func (e *Executor) executeArrayItemField(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail ArrayItemFieldRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	// 檢查路徑是否包含萬用字元
	if strings.Contains(ruleDetail.Path, "[*]") {
		// 展開萬用字元路徑
		paths := e.parser.ExpandWildcardPath(ruleDetail.Path)
		if paths == nil || len(paths) == 0 {
			return nil
		}

		var results []*ValidationResult
		for _, pathInfo := range paths {
			// pathInfo.Value 應該是一個陣列
			arr, ok := pathInfo.Value.([]interface{})
			if !ok {
				continue
			}

			// 檢查這個陣列的每個項目
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
							File:          filePath,
							RuleID:        rule.ID,
							RuleName:      rule.Name,
							Severity:      rule.Severity,
							Message:       ruleDetail.Message,
							Path:          fmt.Sprintf("%s[%d].%s", pathInfo.Path, i, ruleDetail.Field),
							ActualValue:   fieldStr,
							ExpectedValue: fmt.Sprintf("one of [%s]", strings.Join(ruleDetail.Validation.AllowedValues, ", ")),
						})
					}
				}
			}
		}
		return results
	}

	// 原有的單一路徑邏輯
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
					File:          filePath,
					RuleID:        rule.ID,
					RuleName:      rule.Name,
					Severity:      rule.Severity,
					Message:       ruleDetail.Message,
					Path:          fmt.Sprintf("%s[%d].%s", ruleDetail.Path, i, ruleDetail.Field),
					ActualValue:   fieldStr,
					ExpectedValue: fmt.Sprintf("one of [%s]", strings.Join(ruleDetail.Validation.AllowedValues, ", ")),
				})
			}
		}
	}

	return results
}

// executePatternMatch 執行正則表達式驗證
// 支援萬用字元，例如 routes[*].path 會檢查每個 routes 項目的 path 格式
func (e *Executor) executePatternMatch(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail PatternMatchRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	// 預先編譯正則表達式
	re, err := regexp.Compile(ruleDetail.Pattern)
	if err != nil {
		return makeErrorResult(rule, filePath, ruleDetail.Path, fmt.Sprintf("正則表達式錯誤: %v", err))
	}

	return e.processPathWithWildcard(ruleDetail.Path, func(actualPath string, value interface{}) *ValidationResult {
		// 轉換為字串
		strValue, ok := value.(string)
		if !ok {
			// 不是字串，跳過檢查
			return nil
		}

		// 檢查是否匹配
		if !re.MatchString(strValue) {
			return &ValidationResult{
				File:          filePath,
				RuleID:        rule.ID,
				RuleName:      rule.Name,
				Severity:      rule.Severity,
				Message:       ruleDetail.Message,
				Path:          actualPath,
				ActualValue:   strValue,
				ExpectedValue: fmt.Sprintf("pattern: %s", ruleDetail.Pattern),
			}
		}
		return nil
	})
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

// executeArrayNoDuplicates 執行陣列欄位不可重複檢查
// 支援萬用字元，例如 routes[*].middlewares 會檢查每個 route 的 middlewares 內部是否重複
func (e *Executor) executeArrayNoDuplicates(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail ArrayNoDuplicatesRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	// 檢查路徑是否包含萬用字元
	if strings.Contains(ruleDetail.Path, "[*]") {
		// 展開萬用字元路徑
		paths := e.parser.ExpandWildcardPath(ruleDetail.Path)
		if paths == nil || len(paths) == 0 {
			return nil
		}

		var results []*ValidationResult
		for _, pathInfo := range paths {
			// pathInfo.Value 應該是一個陣列
			_, ok := pathInfo.Value.([]interface{})
			if !ok {
				continue
			}

			// 檢查這個陣列內部的重複
			duplicates, err := e.parser.CheckArrayDuplicates(pathInfo.Path, ruleDetail.Field)
			if err != nil {
				continue
			}

			// 為每個重複的索引建立一個錯誤
			for _, dup := range duplicates {
				for _, idx := range dup.Indices {
					results = append(results, &ValidationResult{
						File:     filePath,
						RuleID:   rule.ID,
						RuleName: rule.Name,
						Severity: rule.Severity,
						Message:  fmt.Sprintf("%s (重複值: %s)", ruleDetail.Message, dup.Value),
						Path:     fmt.Sprintf("%s[%d].%s", pathInfo.Path, idx, ruleDetail.Field),
					})
				}
			}
		}
		return results
	}

	// 原有的單一路徑邏輯
	duplicates, err := e.parser.CheckArrayDuplicates(ruleDetail.Path, ruleDetail.Field)
	if err != nil {
		// 如果陣列不存在,不回報錯誤
		return nil
	}

	var results []*ValidationResult
	for _, dup := range duplicates {
		// 為每個重複的索引建立一個錯誤
		for _, idx := range dup.Indices {
			results = append(results, &ValidationResult{
				File:     filePath,
				RuleID:   rule.ID,
				RuleName: rule.Name,
				Severity: rule.Severity,
				Message:  fmt.Sprintf("%s (重複值: %s)", ruleDetail.Message, dup.Value),
				Path:     fmt.Sprintf("%s[%d].%s", ruleDetail.Path, idx, ruleDetail.Field),
			})
		}
	}

	return results
}

// executeArrayNoDuplicatesCombine 執行陣列多欄位組合不可重複檢查
func (e *Executor) executeArrayNoDuplicatesCombine(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail ArrayNoDuplicatesCombineRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	duplicates, err := e.parser.CheckArrayMultiFieldDuplicates(ruleDetail.Path, ruleDetail.Fields)
	if err != nil {
		// 如果陣列不存在,不回報錯誤
		return nil
	}

	var results []*ValidationResult
	for _, dup := range duplicates {
		// 為每個重複的索引建立一個錯誤
		for _, idx := range dup.Indices {
			results = append(results, &ValidationResult{
				File:     filePath,
				RuleID:   rule.ID,
				RuleName: rule.Name,
				Severity: rule.Severity,
				Message:  fmt.Sprintf("%s (重複組合: %s)", ruleDetail.Message, dup.Value),
				Path:     fmt.Sprintf("%s[%d]", ruleDetail.Path, idx),
			})
		}
	}

	return results
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

// processPathWithWildcard 通用的通配符處理函數
// 如果 path 包含 [*]，展開所有路徑並對每個路徑執行 checkFunc
// 否則直接對單個路徑執行 checkFunc
func (e *Executor) processPathWithWildcard(
	path string,
	checkFunc func(actualPath string, value interface{}) *ValidationResult,
) []*ValidationResult {
	if strings.Contains(path, "[*]") {
		// 展開通配符路徑
		paths := e.parser.ExpandWildcardPath(path)
		if paths == nil || len(paths) == 0 {
			return nil
		}

		var results []*ValidationResult
		for _, pathInfo := range paths {
			if result := checkFunc(pathInfo.Path, pathInfo.Value); result != nil {
				results = append(results, result)
			}
		}
		return results
	}

	// 單一路徑
	value, exists := e.parser.GetValue(path)
	if !exists {
		return nil
	}
	if result := checkFunc(path, value); result != nil {
		return []*ValidationResult{result}
	}
	return nil
}

// executeHashedValueCheck 執行雜湊值檢查
func (e *Executor) executeHashedValueCheck(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail HashedValueCheckRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	value, exists := e.parser.GetString(ruleDetail.Path)
	if !exists {
		return nil // 欄位不存在，不檢查
	}

	// 計算雜湊值
	var hasher hash.Hash
	switch strings.ToLower(ruleDetail.HashAlgorithm) {
	case "sha256":
		hasher = sha256.New()
	case "sha1":
		hasher = sha1.New()
	case "sha512":
		hasher = sha512.New()
	case "md5":
		hasher = md5.New()
	default:
		return makeErrorResult(rule, filePath, ruleDetail.Path, fmt.Sprintf("不支援的雜湊演算法: %s", ruleDetail.HashAlgorithm))
	}

	hasher.Write([]byte(value))
	hashValue := hex.EncodeToString(hasher.Sum(nil))

	// 檢查是否在列表中
	inList := false
	for _, h := range ruleDetail.HashList {
		if strings.EqualFold(hashValue, h) {
			inList = true
			break
		}
	}

	// 根據模式判斷是否違規
	violation := false
	if ruleDetail.Mode == "forbidden" && inList {
		violation = true // 禁止模式且在列表中
	} else if ruleDetail.Mode == "allowed" && !inList {
		violation = true // 允許模式但不在列表中
	}

	if violation {
		return []*ValidationResult{{
			File:     filePath,
			RuleID:   rule.ID,
			RuleName: rule.Name,
			Severity: rule.Severity,
			Message:  ruleDetail.Message,
			Path:     ruleDetail.Path,
		}}
	}

	return nil
}

// executeContainsKeywords 執行關鍵字檢查
func (e *Executor) executeContainsKeywords(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail ContainsKeywordsRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	// 檢查路徑是否包含萬用字元
	if strings.Contains(ruleDetail.Path, "[*]") {
		// 展開萬用字元路徑
		paths := e.parser.ExpandWildcardPath(ruleDetail.Path)
		if paths == nil || len(paths) == 0 {
			return nil
		}

		var results []*ValidationResult
		for _, pathInfo := range paths {
			value, ok := pathInfo.Value.(string)
			if !ok {
				continue
			}

			// 檢查此值
			if violation, matchedKeyword := e.checkKeywords(value, ruleDetail); violation {
				message := ruleDetail.Message
				if matchedKeyword != "" {
					message = fmt.Sprintf("%s (包含關鍵字: %s)", ruleDetail.Message, matchedKeyword)
				}

				results = append(results, &ValidationResult{
					File:     filePath,
					RuleID:   rule.ID,
					RuleName: rule.Name,
					Severity: rule.Severity,
					Message:  message,
					Path:     pathInfo.Path,
				})
			}
		}
		return results
	}

	// 原有的非萬用字元邏輯
	value, exists := e.parser.GetString(ruleDetail.Path)
	if !exists {
		return nil // 欄位不存在，不檢查
	}

	if violation, matchedKeyword := e.checkKeywords(value, ruleDetail); violation {
		message := ruleDetail.Message
		if matchedKeyword != "" {
			message = fmt.Sprintf("%s (包含關鍵字: %s)", ruleDetail.Message, matchedKeyword)
		}

		return []*ValidationResult{{
			File:     filePath,
			RuleID:   rule.ID,
			RuleName: rule.Name,
			Severity: rule.Severity,
			Message:  message,
			Path:     ruleDetail.Path,
		}}
	}

	return nil
}

// checkKeywords 檢查字串是否包含關鍵字
func (e *Executor) checkKeywords(value string, ruleDetail ContainsKeywordsRule) (bool, string) {
	// 處理大小寫
	checkValue := value
	keywords := make([]string, len(ruleDetail.Keywords))
	copy(keywords, ruleDetail.Keywords)

	if !ruleDetail.CaseSensitive {
		checkValue = strings.ToLower(value)
		for i := range keywords {
			keywords[i] = strings.ToLower(keywords[i])
		}
	}

	// 檢查是否包含關鍵字
	containsAny := false
	matchedKeyword := ""
	for _, keyword := range keywords {
		if strings.Contains(checkValue, keyword) {
			containsAny = true
			matchedKeyword = keyword
			break
		}
	}

	// 根據模式判斷是否違規
	violation := false
	if ruleDetail.Mode == "forbidden" && containsAny {
		violation = true // 禁止模式且包含關鍵字
	} else if ruleDetail.Mode == "required" && !containsAny {
		violation = true // 必須模式但不包含關鍵字
	}

	return violation, matchedKeyword
}

// executeNoTrailingWhitespace 執行 trailing whitespace 檢查
// 自動掃描整個 YAML 檔案中所有字串欄位
func (e *Executor) executeNoTrailingWhitespace(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail NoTrailingWhitespaceRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	// 獲取整個 YAML 文件的資料
	rootData, exists := e.parser.GetValue("")
	if !exists {
		return nil
	}

	var results []*ValidationResult

	// 遞迴檢查所有字串欄位
	e.scanForTrailingWhitespace(rootData, "", &results, rule, filePath, ruleDetail.Message)

	return results
}

// scanForTrailingWhitespace 遞迴掃描資料結構，檢查所有字串值
func (e *Executor) scanForTrailingWhitespace(data interface{}, currentPath string, results *[]*ValidationResult, rule *ValidationRule, filePath, message string) {
	switch v := data.(type) {
	case map[string]interface{}:
		// 遍歷物件的每個欄位
		for key, value := range v {
			newPath := key
			if currentPath != "" {
				newPath = currentPath + "." + key
			}
			e.scanForTrailingWhitespace(value, newPath, results, rule, filePath, message)
		}

	case map[interface{}]interface{}:
		// 處理 YAML 特殊的 map[interface{}]interface{} 格式
		for key, value := range v {
			keyStr, ok := key.(string)
			if !ok {
				continue
			}
			newPath := keyStr
			if currentPath != "" {
				newPath = currentPath + "." + keyStr
			}
			e.scanForTrailingWhitespace(value, newPath, results, rule, filePath, message)
		}

	case []interface{}:
		// 遍歷陣列的每個項目
		for i, item := range v {
			newPath := fmt.Sprintf("%s[%d]", currentPath, i)
			e.scanForTrailingWhitespace(item, newPath, results, rule, filePath, message)
		}

	case string:
		// 檢查字串值
		if wsType := e.checkTrailingWhitespace(v); wsType != "" {
			*results = append(*results, &ValidationResult{
				File:     filePath,
				RuleID:   rule.ID,
				RuleName: rule.Name,
				Severity: rule.Severity,
				Message:  fmt.Sprintf("%s (%s有空白字元)", message, wsType),
				Path:     currentPath,
			})
		}

	// 其他類型（數字、布林值等）不需要檢查
	}
}

// checkTrailingWhitespace 檢查字串是否有 trailing/leading whitespace
func (e *Executor) checkTrailingWhitespace(value string) string {
	trimmed := strings.TrimSpace(value)
	if value == trimmed {
		return "" // 沒有空白字元
	}

	// 判斷是 leading 還是 trailing
	var wsType string
	if strings.HasPrefix(value, " ") || strings.HasPrefix(value, "\t") {
		wsType = "開頭"
	}
	if strings.HasSuffix(value, " ") || strings.HasSuffix(value, "\t") {
		if wsType != "" {
			wsType = "開頭和結尾"
		} else {
			wsType = "結尾"
		}
	}

	return wsType
}
