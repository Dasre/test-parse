package reporter

import (
	"config-validator/internal/rule"
	"encoding/json"
	"fmt"
	"sort"
)

// Reporter 結果輸出器
type Reporter struct {
	results []*rule.ValidationResult
}

// NewReporter 建立新的輸出器
func NewReporter() *Reporter {
	return &Reporter{
		results: make([]*rule.ValidationResult, 0),
	}
}

// AddResults 添加驗證結果
func (r *Reporter) AddResults(results []*rule.ValidationResult) {
	r.results = append(r.results, results...)
}

// HasErrors 是否有錯誤
func (r *Reporter) HasErrors() bool {
	for _, result := range r.results {
		if result.Severity == rule.SeverityError {
			return true
		}
	}
	return false
}

// PrintConsole 輸出到終端（友好格式）
func (r *Reporter) PrintConsole(ruleCount int) {
	fmt.Printf("📋 載入了 %d 條規則\n\n", ruleCount)

	if len(r.results) == 0 {
		fmt.Println("✅ 所有驗證通過")
		return
	}

	// 按檔案分組
	fileGroups := r.groupByFile()

	// 輸出每個檔案的驗證結果
	for _, file := range r.getSortedFiles(fileGroups) {
		fmt.Printf("📄 %s\n", file)

		for _, result := range fileGroups[file] {
			icon := r.getSeverityIcon(result.Severity)
			fmt.Printf("  %s [%s] %s\n", icon, result.RuleID, result.RuleName)
			fmt.Printf("     %s\n", result.Message)
			if result.Path != "" {
				fmt.Printf("     路徑: %s\n", result.Path)
			}
			if result.ActualValue != "" {
				fmt.Printf("     實際值: %s\n", result.ActualValue)
			}
			if result.ExpectedValue != "" {
				fmt.Printf("     期望值: %s\n", result.ExpectedValue)
			}
		}
		fmt.Println()
	}

	// 統計
	fmt.Println("==================================================")
	errorCount, warningCount := r.countBySeverity()
	if errorCount > 0 {
		fmt.Printf("❌ %d 個錯誤\n", errorCount)
	}
	if warningCount > 0 {
		fmt.Printf("⚠️  %d 個警告\n", warningCount)
	}
}

// PrintJSON 輸出 JSON 格式
func (r *Reporter) PrintJSON() error {
	output := map[string]interface{}{
		"total":   len(r.results),
		"results": r.results,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("生成 JSON 失敗: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// groupByFile 按檔案分組結果
func (r *Reporter) groupByFile() map[string][]*rule.ValidationResult {
	groups := make(map[string][]*rule.ValidationResult)

	for _, result := range r.results {
		groups[result.File] = append(groups[result.File], result)
	}

	return groups
}

// getSortedFiles 獲取排序後的檔案列表
func (r *Reporter) getSortedFiles(groups map[string][]*rule.ValidationResult) []string {
	files := make([]string, 0, len(groups))
	for file := range groups {
		files = append(files, file)
	}
	sort.Strings(files)
	return files
}

// getSeverityIcon 獲取嚴重程度圖示
func (r *Reporter) getSeverityIcon(severity rule.Severity) string {
	switch severity {
	case rule.SeverityError:
		return "❌"
	case rule.SeverityWarning:
		return "⚠️ "
	case rule.SeverityInfo:
		return "ℹ️ "
	default:
		return "  "
	}
}

// countBySeverity 統計各嚴重程度的數量
func (r *Reporter) countBySeverity() (errors, warnings int) {
	for _, result := range r.results {
		switch result.Severity {
		case rule.SeverityError:
			errors++
		case rule.SeverityWarning:
			warnings++
		}
	}
	return
}
