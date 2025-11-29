package rule

// Severity 定義規則的嚴重程度
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// RuleType 定義規則類型
type RuleType string

const (
	RuleTypeRequiredField            RuleType = "required_field"
	RuleTypeRequiredFields           RuleType = "required_fields"
	RuleTypeFieldType                RuleType = "field_type"
	RuleTypeValueRange               RuleType = "value_range"
	RuleTypeArrayItemRequiredFields  RuleType = "array_item_required_fields"
	RuleTypeArrayItemField           RuleType = "array_item_field"
	RuleTypePatternMatch             RuleType = "pattern_match"
	RuleTypeArrayNoDuplicates        RuleType = "array_no_duplicates"
	RuleTypeArrayNoDuplicatesCombine RuleType = "array_no_duplicates_combine"
	RuleTypeHashedValueCheck         RuleType = "hashed_value_check"
	RuleTypeContainsKeywords         RuleType = "contains_keywords"
	RuleTypeNoTrailingWhitespace     RuleType = "no_trailing_whitespace"
)

// FieldType 定義欄位類型
type FieldType string

const (
	FieldTypeString  FieldType = "string"
	FieldTypeNumber  FieldType = "number"
	FieldTypeBoolean FieldType = "boolean"
	FieldTypeArray   FieldType = "array"
	FieldTypeObject  FieldType = "object"
)

// ValidationRule 定義驗證規則
type ValidationRule struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Enabled     bool     `yaml:"enabled"`
	Severity    Severity `yaml:"severity"`
	Description string   `yaml:"description,omitempty"`
	Targets     Targets  `yaml:"targets"`
	Rule        Rule     `yaml:"rule"`
}

// Targets 定義規則適用的目標檔案
type Targets struct {
	FilePatterns []string `yaml:"file_patterns"`
}

// Rule 定義具體的驗證邏輯
type Rule struct {
	Type     RuleType               `yaml:"type"`
	RawRule  map[string]interface{} `yaml:",inline"`
}

// RequiredFieldRule 必要欄位規則
type RequiredFieldRule struct {
	Path    string `yaml:"path"`
	Message string `yaml:"message"`
}

// RequiredFieldsRule 多個必要欄位規則
type RequiredFieldsRule struct {
	Path    string   `yaml:"path"`
	Fields  []string `yaml:"fields"`
	Message string   `yaml:"message"`
}

// FieldTypeRule 欄位類型規則
type FieldTypeRule struct {
	Path         string    `yaml:"path"`
	ExpectedType FieldType `yaml:"expected_type"`
	Message      string    `yaml:"message"`
}

// ValueRangeRule 數值範圍規則
type ValueRangeRule struct {
	Path    string  `yaml:"path"`
	Min     float64 `yaml:"min"`
	Max     float64 `yaml:"max"`
	Message string  `yaml:"message"`
}

// ArrayItemRequiredFieldsRule 陣列項目必要欄位規則
type ArrayItemRequiredFieldsRule struct {
	Path           string   `yaml:"path"`
	RequiredFields []string `yaml:"required_fields"`
	Message        string   `yaml:"message"`
}

// ArrayItemFieldRule 陣列項目欄位規則
type ArrayItemFieldRule struct {
	Path       string     `yaml:"path"`
	Field      string     `yaml:"field"`
	Validation Validation `yaml:"validation"`
	Message    string     `yaml:"message"`
}

// Validation 定義驗證類型
type Validation struct {
	Type          string   `yaml:"type"`
	AllowedValues []string `yaml:"allowed_values,omitempty"`
}

// PatternMatchRule 正則表達式規則
type PatternMatchRule struct {
	Path    string `yaml:"path"`
	Pattern string `yaml:"pattern"`
	Message string `yaml:"message"`
}

// ArrayNoDuplicatesRule 陣列欄位不可重複規則
type ArrayNoDuplicatesRule struct {
	Path    string `yaml:"path"`
	Field   string `yaml:"field"`
	Message string `yaml:"message"`
}

// ArrayNoDuplicatesCombineRule 陣列多欄位組合不可重複規則
type ArrayNoDuplicatesCombineRule struct {
	Path    string   `yaml:"path"`
	Fields  []string `yaml:"fields"`
	Message string   `yaml:"message"`
}

// HashedValueCheckRule SHA 雜湊值檢查規則
type HashedValueCheckRule struct {
	Path          string   `yaml:"path"`
	HashAlgorithm string   `yaml:"hash_algorithm"` // sha1, sha256, sha512, md5
	Mode          string   `yaml:"mode"`           // forbidden, allowed
	HashList      []string `yaml:"hash_list"`      // 雜湊值列表
	Message       string   `yaml:"message"`
}

// ContainsKeywordsRule 關鍵字檢查規則
type ContainsKeywordsRule struct {
	Path          string   `yaml:"path"`
	Mode          string   `yaml:"mode"`           // forbidden, required
	CaseSensitive bool     `yaml:"case_sensitive"` // 是否區分大小寫
	Keywords      []string `yaml:"keywords"`       // 關鍵字列表
	Message       string   `yaml:"message"`
}

// NoTrailingWhitespaceRule Trailing whitespace 檢查規則
// 自動掃描整個 YAML 檔案中所有字串欄位，檢查是否有 trailing/leading 空白
type NoTrailingWhitespaceRule struct {
	Message string `yaml:"message"`
}

// ValidationResult 驗證結果
type ValidationResult struct {
	File          string   `json:"file"`
	RuleID        string   `json:"rule_id"`
	RuleName      string   `json:"rule_name"`
	Severity      Severity `json:"severity"`
	Message       string   `json:"message"`
	Path          string   `json:"path"`
	ActualValue   string   `json:"actual_value,omitempty"`   // 實際值
	ExpectedValue string   `json:"expected_value,omitempty"` // 期望值
}
