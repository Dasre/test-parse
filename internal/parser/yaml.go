package parser

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// YAMLParser 處理 YAML 檔案解析
type YAMLParser struct {
	data map[string]interface{}
}

// NewYAMLParser 建立新的 YAML 解析器
func NewYAMLParser() *YAMLParser {
	return &YAMLParser{
		data: make(map[string]interface{}),
	}
}

// ParseFile 解析 YAML 檔案
func (p *YAMLParser) ParseFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("讀取檔案失敗: %w", err)
	}

	if err := yaml.Unmarshal(content, &p.data); err != nil {
		return fmt.Errorf("解析 YAML 失敗: %w", err)
	}

	return nil
}

// GetValue 根據路徑獲取值
// 路徑格式: "database.pool.maxConnections" 或 "routes[0].path" 或 "routes[*].middlewares[0]"
// 支援萬用字元 [*] 表示所有陣列項目,當使用 [*] 時會返回多個結果的陣列
func (p *YAMLParser) GetValue(path string) (interface{}, bool) {
	if path == "" {
		return p.data, true
	}

	// 檢查路徑中是否包含萬用字元 [*]
	if strings.Contains(path, "[*]") {
		return p.getValuesWithWildcard(path)
	}

	// 原有的單一路徑邏輯
	parts := strings.Split(path, ".")
	current := interface{}(p.data)

	for _, part := range parts {
		// 檢查是否包含陣列索引 (如: routes[0])
		if strings.Contains(part, "[") {
			fieldName, index, err := parseArrayIndex(part)
			if err != nil {
				return nil, false
			}

			// 先取得陣列
			switch v := current.(type) {
			case map[string]interface{}:
				arr, exists := v[fieldName]
				if !exists {
					return nil, false
				}
				current = arr
			case map[interface{}]interface{}:
				arr, exists := v[fieldName]
				if !exists {
					return nil, false
				}
				current = arr
			default:
				return nil, false
			}

			// 再取得陣列中的元素
			arrValue, ok := current.([]interface{})
			if !ok {
				return nil, false
			}
			if index < 0 || index >= len(arrValue) {
				return nil, false
			}
			current = arrValue[index]
		} else {
			// 一般的欄位存取
			switch v := current.(type) {
			case map[string]interface{}:
				value, exists := v[part]
				if !exists {
					return nil, false
				}
				current = value
			case map[interface{}]interface{}:
				// YAML 有時會將 key 解析為 interface{}
				value, exists := v[part]
				if !exists {
					return nil, false
				}
				current = value
			default:
				return nil, false
			}
		}
	}

	return current, true
}

// getValuesWithWildcard 處理包含萬用字元 [*] 的路徑
// 返回所有匹配的值的陣列
func (p *YAMLParser) getValuesWithWildcard(path string) (interface{}, bool) {
	parts := strings.Split(path, ".")

	// 從根開始,逐步處理每個部分
	var results []interface{}
	results = append(results, p.data)

	for _, part := range parts {
		var newResults []interface{}

		for _, current := range results {
			if strings.Contains(part, "[*]") {
				// 處理萬用字元 [*]
				fieldName := strings.TrimSuffix(part, "[*]")

				// 取得陣列
				var arr []interface{}
				switch v := current.(type) {
				case map[string]interface{}:
					if arrVal, exists := v[fieldName]; exists {
						if a, ok := arrVal.([]interface{}); ok {
							arr = a
						}
					}
				case map[interface{}]interface{}:
					if arrVal, exists := v[fieldName]; exists {
						if a, ok := arrVal.([]interface{}); ok {
							arr = a
						}
					}
				}

				// 將陣列中的所有項目加入結果
				newResults = append(newResults, arr...)

			} else if strings.Contains(part, "[") {
				// 處理特定索引 [0], [1] 等
				fieldName, index, err := parseArrayIndex(part)
				if err != nil {
					continue
				}

				var arr []interface{}
				switch v := current.(type) {
				case map[string]interface{}:
					if arrVal, exists := v[fieldName]; exists {
						if a, ok := arrVal.([]interface{}); ok {
							arr = a
						}
					}
				case map[interface{}]interface{}:
					if arrVal, exists := v[fieldName]; exists {
						if a, ok := arrVal.([]interface{}); ok {
							arr = a
						}
					}
				}

				if index >= 0 && index < len(arr) {
					newResults = append(newResults, arr[index])
				}

			} else {
				// 一般欄位存取
				switch v := current.(type) {
				case map[string]interface{}:
					if value, exists := v[part]; exists {
						newResults = append(newResults, value)
					}
				case map[interface{}]interface{}:
					if value, exists := v[part]; exists {
						newResults = append(newResults, value)
					}
				}
			}
		}

		results = newResults
		if len(results) == 0 {
			return nil, false
		}
	}

	// 如果只有一個結果,直接返回該值
	if len(results) == 1 {
		return results[0], true
	}

	// 多個結果時返回陣列
	return results, true
}

// parseArrayIndex 解析陣列索引，如: "routes[0]" -> ("routes", 0, nil)
func parseArrayIndex(s string) (string, int, error) {
	start := strings.Index(s, "[")
	end := strings.Index(s, "]")

	if start == -1 || end == -1 || start >= end {
		return "", 0, fmt.Errorf("無效的陣列索引格式: %s", s)
	}

	fieldName := s[:start]
	indexStr := s[start+1 : end]

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return "", 0, fmt.Errorf("無效的索引數字: %s", indexStr)
	}

	return fieldName, index, nil
}

// HasField 檢查路徑是否存在
func (p *YAMLParser) HasField(path string) bool {
	_, exists := p.GetValue(path)
	return exists
}

// GetArray 獲取陣列
func (p *YAMLParser) GetArray(path string) ([]interface{}, bool) {
	value, exists := p.GetValue(path)
	if !exists {
		return nil, false
	}

	arr, ok := value.([]interface{})
	return arr, ok
}

// GetString 獲取字串值
func (p *YAMLParser) GetString(path string) (string, bool) {
	value, exists := p.GetValue(path)
	if !exists {
		return "", false
	}

	str, ok := value.(string)
	return str, ok
}

// GetNumber 獲取數值
func (p *YAMLParser) GetNumber(path string) (float64, bool) {
	value, exists := p.GetValue(path)
	if !exists {
		return 0, false
	}

	switch v := value.(type) {
	case int:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

// GetBool 獲取布林值
func (p *YAMLParser) GetBool(path string) (bool, bool) {
	value, exists := p.GetValue(path)
	if !exists {
		return false, false
	}

	b, ok := value.(bool)
	return b, ok
}

// GetType 獲取值的類型
func (p *YAMLParser) GetType(path string) string {
	value, exists := p.GetValue(path)
	if !exists {
		return "unknown"
	}

	switch value.(type) {
	case string:
		return "string"
	case int, float64:
		return "number"
	case bool:
		return "boolean"
	case []interface{}:
		return "array"
	case map[string]interface{}, map[interface{}]interface{}:
		return "object"
	default:
		return "unknown"
	}
}

// GetMap 獲取 map 物件
func (p *YAMLParser) GetMap(path string) (map[string]interface{}, bool) {
	value, exists := p.GetValue(path)
	if !exists {
		return nil, false
	}

	switch v := value.(type) {
	case map[string]interface{}:
		return v, true
	case map[interface{}]interface{}:
		// 轉換 map[interface{}]interface{} 為 map[string]interface{}
		result := make(map[string]interface{})
		for key, val := range v {
			if strKey, ok := key.(string); ok {
				result[strKey] = val
			}
		}
		return result, true
	default:
		return nil, false
	}
}

// DuplicateInfo 重複欄位資訊
type DuplicateInfo struct {
	Field   string // 重複的欄位名稱
	Value   string // 重複的值
	Indices []int  // 重複出現的陣列索引
}

// PathInfo 路徑資訊，用於追蹤萬用字元展開後的實際路徑
type PathInfo struct {
	Path  string      // 實際路徑 (如: "routes[0].middlewares[1]")
	Value interface{} // 該路徑的值
}

// ExpandWildcardPath 展開包含萬用字元 [*] 的路徑
// 返回所有匹配的實際路徑及其值
func (p *YAMLParser) ExpandWildcardPath(path string) []*PathInfo {
	if !strings.Contains(path, "[*]") {
		// 沒有萬用字元,直接返回
		if value, exists := p.GetValue(path); exists {
			return []*PathInfo{{Path: path, Value: value}}
		}
		return nil
	}

	parts := strings.Split(path, ".")
	var paths []*PathInfo
	paths = append(paths, &PathInfo{Path: "", Value: p.data})

	for partIdx, part := range parts {
		var newPaths []*PathInfo

		for _, pathInfo := range paths {
			currentPath := pathInfo.Path
			current := pathInfo.Value

			if strings.Contains(part, "[*]") {
				// 處理萬用字元 [*]
				fieldName := strings.TrimSuffix(part, "[*]")

				// 取得陣列
				var arr []interface{}
				switch v := current.(type) {
				case map[string]interface{}:
					if arrVal, exists := v[fieldName]; exists {
						if a, ok := arrVal.([]interface{}); ok {
							arr = a
						}
					}
				case map[interface{}]interface{}:
					if arrVal, exists := v[fieldName]; exists {
						if a, ok := arrVal.([]interface{}); ok {
							arr = a
						}
					}
				}

				// 為陣列中的每個項目建立新路徑
				for idx, item := range arr {
					newPath := currentPath
					if newPath != "" {
						newPath += "."
					}
					newPath += fmt.Sprintf("%s[%d]", fieldName, idx)

					// 如果這是最後一個部分,直接加入結果
					// 否則繼續處理後續部分
					if partIdx == len(parts)-1 {
						newPaths = append(newPaths, &PathInfo{Path: newPath, Value: item})
					} else {
						newPaths = append(newPaths, &PathInfo{Path: newPath, Value: item})
					}
				}

			} else if strings.Contains(part, "[") {
				// 處理特定索引
				fieldName, index, err := parseArrayIndex(part)
				if err != nil {
					continue
				}

				var arr []interface{}
				switch v := current.(type) {
				case map[string]interface{}:
					if arrVal, exists := v[fieldName]; exists {
						if a, ok := arrVal.([]interface{}); ok {
							arr = a
						}
					}
				case map[interface{}]interface{}:
					if arrVal, exists := v[fieldName]; exists {
						if a, ok := arrVal.([]interface{}); ok {
							arr = a
						}
					}
				}

				if index >= 0 && index < len(arr) {
					newPath := currentPath
					if newPath != "" {
						newPath += "."
					}
					newPath += fmt.Sprintf("%s[%d]", fieldName, index)
					newPaths = append(newPaths, &PathInfo{Path: newPath, Value: arr[index]})
				}

			} else {
				// 一般欄位存取
				var value interface{}
				var exists bool

				switch v := current.(type) {
				case map[string]interface{}:
					value, exists = v[part]
				case map[interface{}]interface{}:
					value, exists = v[part]
				}

				if exists {
					newPath := currentPath
					if newPath != "" {
						newPath += "."
					}
					newPath += part
					newPaths = append(newPaths, &PathInfo{Path: newPath, Value: value})
				}
			}
		}

		paths = newPaths
		if len(paths) == 0 {
			return nil
		}
	}

	return paths
}

// CheckArrayDuplicates 檢查陣列中某個欄位是否有重複值
// 返回所有重複的欄位資訊
func (p *YAMLParser) CheckArrayDuplicates(arrayPath string, field string) ([]*DuplicateInfo, error) {
	arr, exists := p.GetArray(arrayPath)
	if !exists {
		return nil, fmt.Errorf("陣列路徑不存在: %s", arrayPath)
	}

	// 用於追蹤每個值出現的索引
	valueIndices := make(map[string][]int)

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

		// 取得欄位值
		fieldValue, exists := itemMap[field]
		if !exists {
			continue
		}

		// 將值轉為字串以便比較
		strValue := fmt.Sprintf("%v", fieldValue)
		valueIndices[strValue] = append(valueIndices[strValue], i)
	}

	// 找出重複的值
	var duplicates []*DuplicateInfo
	for value, indices := range valueIndices {
		if len(indices) > 1 {
			duplicates = append(duplicates, &DuplicateInfo{
				Field:   field,
				Value:   value,
				Indices: indices,
			})
		}
	}

	return duplicates, nil
}

// CheckArrayMultiFieldDuplicates 檢查陣列中多個欄位組合是否有重複
// fields: 要檢查的欄位列表，例如 ["name", "version"] 會檢查 name+version 的組合是否重複
func (p *YAMLParser) CheckArrayMultiFieldDuplicates(arrayPath string, fields []string) ([]*DuplicateInfo, error) {
	arr, exists := p.GetArray(arrayPath)
	if !exists {
		return nil, fmt.Errorf("陣列路徑不存在: %s", arrayPath)
	}

	// 用於追蹤每個組合值出現的索引
	valueIndices := make(map[string][]int)

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

		// 組合所有欄位的值
		var combinedValue strings.Builder
		allFieldsExist := true
		for j, field := range fields {
			fieldValue, exists := itemMap[field]
			if !exists {
				allFieldsExist = false
				break
			}
			if j > 0 {
				combinedValue.WriteString("|")
			}
			combinedValue.WriteString(fmt.Sprintf("%v", fieldValue))
		}

		if !allFieldsExist {
			continue
		}

		key := combinedValue.String()
		valueIndices[key] = append(valueIndices[key], i)
	}

	// 找出重複的值
	var duplicates []*DuplicateInfo
	for value, indices := range valueIndices {
		if len(indices) > 1 {
			duplicates = append(duplicates, &DuplicateInfo{
				Field:   strings.Join(fields, "+"),
				Value:   value,
				Indices: indices,
			})
		}
	}

	return duplicates, nil
}
