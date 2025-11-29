package product

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Detector 產品檢測器
type Detector struct {
	products []ProductConfig
}

// NewDetector 建立新的產品檢測器
func NewDetector(configPath string) (*Detector, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("讀取產品配置失敗: %w", err)
	}

	var config ProductsConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("解析產品配置失敗: %w", err)
	}

	return &Detector{
		products: config.Products,
	}, nil
}

// DetectProduct 根據檔案路徑檢測產品類型
func (d *Detector) DetectProduct(filePath string) *ProductConfig {
	// 標準化路徑分隔符
	normalizedPath := filepath.ToSlash(filePath)
	fileName := filepath.Base(filePath)

	// 遍歷所有產品配置
	for _, product := range d.products {
		for _, pattern := range product.PathPatterns {
			// 檢查路徑是否匹配
			if matchPattern(normalizedPath, pattern) || matchPattern(fileName, pattern) {
				return &product
			}
		}
	}

	return nil
}

// matchPattern 簡單的模式匹配
// 支援 **/ 和 * 通配符
func matchPattern(path, pattern string) bool {
	// 處理 **/ 通配符（匹配任意層級目錄）
	if len(pattern) > 3 && pattern[:3] == "**/" {
		// 移除 **/ 前綴，檢查後續部分
		suffix := pattern[3:]
		matched, _ := filepath.Match(suffix, filepath.Base(path))
		if matched {
			return true
		}
		// 也檢查完整路徑
		if containsPattern(path, suffix) {
			return true
		}
	}

	// 標準 filepath.Match
	matched, err := filepath.Match(pattern, filepath.Base(path))
	if err == nil && matched {
		return true
	}

	// 檢查完整路徑
	matched, err = filepath.Match(pattern, path)
	return err == nil && matched
}

// containsPattern 檢查路徑中是否包含匹配模式的部分
func containsPattern(path, pattern string) bool {
	// 移除模式中的通配符進行簡單匹配
	// 例如：pattern = "api*.yaml" -> 檢查路徑中是否有包含 "api" 且結尾是 ".yaml"
	if len(pattern) > 0 && pattern[0] == '*' {
		pattern = pattern[1:]
	}
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		pattern = pattern[:len(pattern)-1]
	}

	// 檢查檔案名稱
	fileName := filepath.Base(path)
	matched, _ := filepath.Match(pattern, fileName)
	return matched
}
