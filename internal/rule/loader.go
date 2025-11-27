package rule

import (
	"fmt"
	"os"
	"path/filepath"
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
