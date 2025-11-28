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
	case RuleTypeNestedArrayNoDuplicates:
		return e.executeNestedArrayNoDuplicates(rule, filePath)
	case RuleTypeNestedArrayItemRequiredFields:
		return e.executeNestedArrayItemRequiredFields(rule, filePath)
	case RuleTypeNestedArrayItemField:
		return e.executeNestedArrayItemField(rule, filePath)
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

// executeArrayNoDuplicates 執行陣列欄位不可重複檢查
func (e *Executor) executeArrayNoDuplicates(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail ArrayNoDuplicatesRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

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

// executeNestedArrayNoDuplicates 執行巢狀陣列欄位不可重複檢查
// 檢查父陣列中每個項目的子陣列是否有重複
func (e *Executor) executeNestedArrayNoDuplicates(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail NestedArrayNoDuplicatesRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	// 取得父陣列
	parentArr, exists := e.parser.GetArray(ruleDetail.ParentPath)
	if !exists {
		return nil // 父陣列不存在,不檢查
	}

	var results []*ValidationResult

	// 遍歷父陣列的每個項目
	for parentIdx := range parentArr {
		// 構建子陣列的完整路徑，如 "apiconfig.routes[0].middlewares"
		childArrayPath := fmt.Sprintf("%s[%d].%s", ruleDetail.ParentPath, parentIdx, ruleDetail.ChildPath)

		var duplicates []*parser.DuplicateInfo
		var err error

		// 根據是單一欄位還是多欄位組合來檢查重複
		if ruleDetail.Field != "" {
			// 單一欄位檢查
			duplicates, err = e.parser.CheckArrayDuplicates(childArrayPath, ruleDetail.Field)
		} else if len(ruleDetail.Fields) > 0 {
			// 多欄位組合檢查
			duplicates, err = e.parser.CheckArrayMultiFieldDuplicates(childArrayPath, ruleDetail.Fields)
		} else {
			return makeErrorResult(rule, filePath, "", "必須指定 field 或 fields")
		}

		if err != nil {
			// 子陣列不存在或其他錯誤,繼續檢查下一個
			continue
		}

		// 將找到的重複項目加入結果
		for _, dup := range duplicates {
			for _, childIdx := range dup.Indices {
				var path string
				if ruleDetail.Field != "" {
					path = fmt.Sprintf("%s[%d].%s[%d].%s", ruleDetail.ParentPath, parentIdx, ruleDetail.ChildPath, childIdx, ruleDetail.Field)
				} else {
					path = fmt.Sprintf("%s[%d].%s[%d]", ruleDetail.ParentPath, parentIdx, ruleDetail.ChildPath, childIdx)
				}

				results = append(results, &ValidationResult{
					File:     filePath,
					RuleID:   rule.ID,
					RuleName: rule.Name,
					Severity: rule.Severity,
					Message:  fmt.Sprintf("%s (重複值: %s)", ruleDetail.Message, dup.Value),
					Path:     path,
				})
			}
		}
	}

	return results
}

// executeNestedArrayItemRequiredFields 執行巢狀陣列項目必要欄位檢查
// 檢查父陣列中每個項目的子陣列項目是否有必要欄位
func (e *Executor) executeNestedArrayItemRequiredFields(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail NestedArrayItemRequiredFieldsRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	// 取得父陣列
	parentArr, exists := e.parser.GetArray(ruleDetail.ParentPath)
	if !exists {
		return nil // 父陣列不存在,不檢查
	}

	var results []*ValidationResult

	// 遍歷父陣列的每個項目
	for parentIdx := range parentArr {
		// 構建子陣列的完整路徑
		childArrayPath := fmt.Sprintf("%s[%d].%s", ruleDetail.ParentPath, parentIdx, ruleDetail.ChildPath)

		// 取得子陣列
		childArr, exists := e.parser.GetArray(childArrayPath)
		if !exists {
			continue // 子陣列不存在,繼續檢查下一個
		}

		// 檢查子陣列的每個項目
		for childIdx, item := range childArr {
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
						Path:     fmt.Sprintf("%s[%d].%s[%d].%s", ruleDetail.ParentPath, parentIdx, ruleDetail.ChildPath, childIdx, field),
					})
				}
			}
		}
	}

	return results
}

// executeNestedArrayItemField 執行巢狀陣列項目欄位驗證
// 驗證父陣列中每個項目的子陣列項目的欄位值
func (e *Executor) executeNestedArrayItemField(rule *ValidationRule, filePath string) []*ValidationResult {
	var ruleDetail NestedArrayItemFieldRule
	if err := unmarshalRule(rule.Rule.RawRule, &ruleDetail); err != nil {
		return makeErrorResult(rule, filePath, "", err.Error())
	}

	// 取得父陣列
	parentArr, exists := e.parser.GetArray(ruleDetail.ParentPath)
	if !exists {
		return nil // 父陣列不存在,不檢查
	}

	var results []*ValidationResult

	// 遍歷父陣列的每個項目
	for parentIdx := range parentArr {
		// 構建子陣列的完整路徑
		childArrayPath := fmt.Sprintf("%s[%d].%s", ruleDetail.ParentPath, parentIdx, ruleDetail.ChildPath)

		// 取得子陣列
		childArr, exists := e.parser.GetArray(childArrayPath)
		if !exists {
			continue // 子陣列不存在,繼續檢查下一個
		}

		// 檢查子陣列的每個項目
		for childIdx, item := range childArr {
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
						Path:     fmt.Sprintf("%s[%d].%s[%d].%s", ruleDetail.ParentPath, parentIdx, ruleDetail.ChildPath, childIdx, ruleDetail.Field),
					})
				}
			}
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
