package parser

import (
	"fmt"
	"os"
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
// 路徑格式: "database.pool.maxConnections"
func (p *YAMLParser) GetValue(path string) (interface{}, bool) {
	if path == "" {
		return p.data, true
	}

	parts := strings.Split(path, ".")
	current := interface{}(p.data)

	for _, part := range parts {
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

	return current, true
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
