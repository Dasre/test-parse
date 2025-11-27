package product

// ProductConfig 產品配置
type ProductConfig struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	RulesDir     string   `yaml:"rules_dir"`
	PathPatterns []string `yaml:"path_patterns"`
}

// ProductsConfig 產品配置集合
type ProductsConfig struct {
	Products []ProductConfig `yaml:"products"`
}
