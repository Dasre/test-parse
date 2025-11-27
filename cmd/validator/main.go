package main

import (
	"config-validator/internal/parser"
	"config-validator/internal/product"
	"config-validator/internal/reporter"
	"config-validator/internal/rule"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// 解析命令行參數
	jsonOutput := flag.Bool("json", false, "輸出 JSON 格式")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "用法: validator <path1> [path2] [path3] ... [--json]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "參數說明:")
		fmt.Fprintln(os.Stderr, "  <path>  配置檔或目錄路徑（可指定多個）")
		fmt.Fprintln(os.Stderr, "  --json  輸出 JSON 格式")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "範例:")
		fmt.Fprintln(os.Stderr, "  validator configs/")
		fmt.Fprintln(os.Stderr, "  validator configs/api.yaml configs/db.yaml")
		fmt.Fprintln(os.Stderr, "  validator testdata/valid testdata/invalid --json")
		os.Exit(1)
	}

	// 獲取所有路徑參數
	paths := flag.Args()

	// 決定產品配置檔路徑（Docker 環境使用 /products.yaml，本地使用 ./products.yaml）
	productsConfig := "/products.yaml"
	if _, err := os.Stat(productsConfig); os.IsNotExist(err) {
		productsConfig = "./products.yaml"
	}

	// 載入產品檢測器
	detector, err := product.NewDetector(productsConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "載入產品配置失敗: %v\n", err)
		os.Exit(1)
	}

	// 收集所有配置檔
	var allConfigFiles []string
	for _, path := range paths {
		// 檢查路徑是否存在
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "路徑不存在: %s\n", path)
			os.Exit(1)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "讀取路徑失敗 %s: %v\n", path, err)
			os.Exit(1)
		}

		// 如果是目錄，掃描其中的配置檔
		if info.IsDir() {
			configFiles, err := scanConfigFiles(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "掃描配置檔失敗 %s: %v\n", path, err)
				os.Exit(1)
			}
			allConfigFiles = append(allConfigFiles, configFiles...)
		} else {
			// 如果是檔案，直接添加（只處理 YAML 檔案）
			if strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml") {
				allConfigFiles = append(allConfigFiles, path)
			} else {
				fmt.Fprintf(os.Stderr, "⚠️  跳過非 YAML 檔案: %s\n", path)
			}
		}
	}

	if len(allConfigFiles) == 0 {
		fmt.Fprintf(os.Stderr, "在指定路徑中沒有找到配置檔\n")
		os.Exit(1)
	}

	// 建立輸出器
	rep := reporter.NewReporter()

	// 統計資訊
	productRulesCount := make(map[string]int)
	totalRulesCount := 0

	// 驗證每個配置檔
	for _, configFile := range allConfigFiles {
		// 檢測產品類型
		prod := detector.DetectProduct(configFile)
		if prod == nil {
			fmt.Fprintf(os.Stderr, "⚠️  無法識別配置檔 %s 的產品類型，跳過驗證\n", configFile)
			continue
		}

		// 如果是第一次處理此產品，載入規則
		if _, exists := productRulesCount[prod.Name]; !exists {
			// 決定規則目錄（Docker 環境使用絕對路徑，本地使用相對路徑）
			rulesDir := "/" + prod.RulesDir
			if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
				rulesDir = "./" + prod.RulesDir
			}

			// 載入該產品的規則
			loader := rule.NewLoader(rulesDir)
			rules, err := loader.LoadRules()
			if err != nil {
				fmt.Fprintf(os.Stderr, "載入產品 %s 的規則失敗: %v\n", prod.Name, err)
				os.Exit(1)
			}

			productRulesCount[prod.Name] = len(rules)
			totalRulesCount += len(rules)

			// 驗證配置檔
			if err := validateFile(configFile, rules, rep); err != nil {
				fmt.Fprintf(os.Stderr, "驗證檔案 %s 失敗: %v\n", configFile, err)
				os.Exit(1)
			}
		} else {
			// 已經載入過該產品的規則，直接驗證
			rulesDir := "/" + prod.RulesDir
			if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
				rulesDir = "./" + prod.RulesDir
			}

			loader := rule.NewLoader(rulesDir)
			rules, err := loader.LoadRules()
			if err != nil {
				fmt.Fprintf(os.Stderr, "載入產品 %s 的規則失敗: %v\n", prod.Name, err)
				os.Exit(1)
			}

			if err := validateFile(configFile, rules, rep); err != nil {
				fmt.Fprintf(os.Stderr, "驗證檔案 %s 失敗: %v\n", configFile, err)
				os.Exit(1)
			}
		}
	}

	// 輸出結果
	if *jsonOutput {
		if err := rep.PrintJSON(); err != nil {
			fmt.Fprintf(os.Stderr, "輸出結果失敗: %v\n", err)
			os.Exit(1)
		}
	} else {
		// 顯示載入的產品規則統計
		if len(productRulesCount) > 0 {
			fmt.Printf("📋 載入了 %d 個產品的規則：\n", len(productRulesCount))
			for prodName, count := range productRulesCount {
				fmt.Printf("   • %s: %d 條規則\n", prodName, count)
			}
			fmt.Println()
		}

		rep.PrintConsole(totalRulesCount)
	}

	// 設置退出碼
	if rep.HasErrors() {
		os.Exit(1)
	}
}

// scanConfigFiles 掃描配置檔目錄
func scanConfigFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// 只處理 YAML 檔案
		if strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// validateFile 驗證單個配置檔
func validateFile(filePath string, rules []*rule.ValidationRule, rep *reporter.Reporter) error {
	// 解析 YAML 檔案
	p := parser.NewYAMLParser()
	if err := p.ParseFile(filePath); err != nil {
		return fmt.Errorf("解析檔案失敗: %w", err)
	}

	// 匹配適用的規則
	matchedRules := rule.MatchRules(rules, filePath)

	// 執行每條規則
	executor := rule.NewExecutor(p)
	for _, r := range matchedRules {
		results := executor.Execute(r, filePath)
		rep.AddResults(results)
	}

	return nil
}
