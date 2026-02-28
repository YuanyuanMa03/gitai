package gitai

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/mayuanyuan/gitai/pkg/token"
	"github.com/urfave/cli/v2"
)

// ModelPricing holds pricing information for AI models
type ModelPricing struct {
	Name           string  `json:"name"`
	Provider       string  `json:"provider"`
	InputPrice     float64 `json:"input_price"`     // Price per 1M input tokens
	OutputPrice    float64 `json:"output_price"`    // Price per 1M output tokens
	Currency       string  `json:"currency"`        // USD or CNY
	ExchangeRate   float64 `json:"exchange_rate"`   // For CNY calculation
}

// Default pricing table (as of 2026)
var defaultPricing = []ModelPricing{
	{
		Name:         "gpt-4o",
		Provider:     "openai",
		InputPrice:   2.50,  // $2.50 per 1M tokens
		OutputPrice:  10.00, // $10.00 per 1M tokens
		Currency:     "USD",
		ExchangeRate: 7.20,  // 1 USD = 7.2 CNY
	},
	{
		Name:         "gpt-4o-mini",
		Provider:     "openai",
		InputPrice:   0.15,  // $0.15 per 1M tokens
		OutputPrice:  0.60,  // $0.60 per 1M tokens
		Currency:     "USD",
		ExchangeRate: 7.20,
	},
	{
		Name:         "gpt-3.5-turbo",
		Provider:     "openai",
		InputPrice:   0.50,  // $0.50 per 1M tokens
		OutputPrice:  1.50,  // $1.50 per 1M tokens
		Currency:     "USD",
		ExchangeRate: 7.20,
	},
	{
		Name:         "qwen-plus",
		Provider:     "alibaba",
		InputPrice:   0.50,  // ¥0.50 per 1M tokens
		OutputPrice:  2.00,  // ¥2.00 per 1M tokens
		Currency:     "CNY",
		ExchangeRate: 1.00,
	},
	{
		Name:         "qwen-turbo",
		Provider:     "alibaba",
		InputPrice:   0.08,  // ¥0.08 per 1M tokens
		OutputPrice:  0.30,  // ¥0.30 per 1M tokens
		Currency:     "CNY",
		ExchangeRate: 1.00,
	},
	{
		Name:         "wenxin-4.0",
		Provider:     "baidu",
		InputPrice:   0.12,  // ¥0.12 per 1M tokens
		OutputPrice:  0.60,  // ¥0.60 per 1M tokens
		Currency:     "CNY",
		ExchangeRate: 1.00,
	},
}

// CostManager manages pricing and cost calculations
type CostManager struct {
	pricing      []ModelPricing
	defaultModel string
	exchangeRate float64
}

// NewCostManager creates a new cost manager
func NewCostManager() *CostManager {
	return &CostManager{
		pricing:      defaultPricing,
		defaultModel: "gpt-4o",
		exchangeRate: 7.20,
	}
}

// GetPricing returns pricing for a model
func (cm *CostManager) GetPricing(modelName string) (*ModelPricing, error) {
	for _, p := range cm.pricing {
		if p.Name == modelName || p.Name == cm.defaultModel {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("model not found: %s", modelName)
}

// CalculateCost calculates cost for given tokens
func (cm *CostManager) CalculateCost(tokens int, modelName string) (usd, cny float64, err error) {
	pricing, err := cm.GetPricing(modelName)
	if err != nil {
		return 0, 0, err
	}

	// Assume 50% input, 50% output for estimation
	inputTokens := float64(tokens) * 0.5
	outputTokens := float64(tokens) * 0.5

	inputCost := (inputTokens / 1000000) * pricing.InputPrice
	outputCost := (outputTokens / 1000000) * pricing.OutputPrice
	totalCost := inputCost + outputCost

	// Convert to USD and CNY
	if pricing.Currency == "USD" {
		usd = totalCost
		cny = totalCost * pricing.ExchangeRate
	} else {
		cny = totalCost
		usd = totalCost / pricing.ExchangeRate
	}

	return usd, cny, nil
}

// calculateCostWithPricing calculates cost using ModelPricing
func calculateCostWithPricing(tokens int, pricing *ModelPricing) (usd, cny float64) {
	// Assume 50% input, 50% output for estimation
	inputTokens := float64(tokens) * 0.5
	outputTokens := float64(tokens) * 0.5

	inputCost := (inputTokens / 1000000) * pricing.InputPrice
	outputCost := (outputTokens / 1000000) * pricing.OutputPrice
	totalCost := inputCost + outputCost

	// Convert to USD and CNY
	if pricing.Currency == "USD" {
		usd = totalCost
		cny = totalCost * pricing.ExchangeRate
	} else {
		cny = totalCost
		usd = totalCost / pricing.ExchangeRate
	}

	return usd, cny
}

// FormatCurrency formats a cost value for display
func FormatCurrency(amount float64, currency string) string {
	if currency == "USD" {
		return fmt.Sprintf("$%.4f", amount)
	}
	return fmt.Sprintf("¥%.2f", amount)
}

// CostCmd shows cost information for tracked files
func CostCmd(c *cli.Context) error {
	// Get git root
	repo, root := MustGetGitRepo()
	_ = repo

	// Create storage
	storage := NewStorage(root)

	// Check if .gitai/ is initialized
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	// Load configuration
	config := MustLoadConfig(root)

	// Parse options
	model := c.String("model")
	if model == "gpt-4o" && config.Model != "" {
		// Use default model from config if user didn't override
		model = config.Model
	}
	currency := c.String("currency")
	if currency == "USD" && config.Currency != "" {
		currency = config.Currency
	}
	aggregateBy := c.String("by") // "file", "section", "role"

	// Get tracked files
	tracked, err := storage.ListAllTracked()
	if err != nil {
		return fmt.Errorf("failed to get tracked files: %w", err)
	}

	if len(tracked) == 0 {
		fmt.Println("No tracked files.")
		return nil
	}

	// Calculate and display costs
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	switch aggregateBy {
	case "section":
		return showCostBySection(tracked, config, model, currency)
	case "role":
		return showCostByRole(tracked, config, model, currency)
	default:
		return showCostByFile(tracked, config, model, currency, w)
	}
}

// showCostByFile shows costs grouped by file
func showCostByFile(tracked []TrackedFile, config *Config, model, currency string, w *tabwriter.Writer) error {
	fmt.Printf("Cost breakdown by file (model: %s)\n\n", model)

	// Table header
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", "File", "Tokens", "USD", "CNY", "Role")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", "----", "------", "---", "---", "----")

	totalTokens := 0
	totalUSD := 0.0
	totalCNY := 0.0

	for _, f := range tracked {
		pricing, err := config.GetModelPricing(model)
		if err != nil {
			continue
		}

		usd, cny := calculateCostWithPricing(f.Tokens, pricing)

		totalTokens += f.Tokens
		totalUSD += usd
		totalCNY += cny

		shortRole := f.Role
		if len(shortRole) > 15 {
			shortRole = shortRole[:12] + ".."
		}

		fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\n",
			f.Path,
			f.Tokens,
			FormatCurrency(usd, "USD"),
			FormatCurrency(cny, "CNY"),
			shortRole,
		)
	}

	w.Flush()

	// Show totals
	fmt.Printf("\n%s\t%d\t%s\t%s\t\n",
		"TOTAL",
		totalTokens,
		FormatCurrency(totalUSD, "USD"),
		FormatCurrency(totalCNY, "CNY"),
	)

	// Show estimated API calls
	showAPICallEstimate(totalTokens, model)

	return nil
}

// showCostBySection shows costs grouped by section
func showCostBySection(tracked []TrackedFile, config *Config, model, currency string) error {
	fmt.Printf("Cost breakdown by section (model: %s)\n\n", model)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Aggregate tokens by section
	sectionTotals := make(map[token.SectionType]int)

	for _, f := range tracked {
		sectionTotals[token.SectionTask] += f.TaskTokens
		sectionTotals[token.SectionConstraints] += f.ConstraintsTokens
		sectionTotals[token.SectionExamples] += f.ExamplesTokens
		sectionTotals[token.SectionInput] += f.InputTokens
		sectionTotals[token.SectionOutput] += f.OutputTokens
		sectionTotals[token.SectionContext] += f.ContextTokens
		sectionTotals[token.SectionOther] += f.OtherTokens
	}

	// Table header
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "Section", "Tokens", "USD", "CNY")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "-------", "------", "---", "---")

	totalTokens := 0
	totalUSD := 0.0
	totalCNY := 0.0

	sections := []struct {
		name string
		typ  token.SectionType
	}{
		{"Task", token.SectionTask},
		{"Constraints", token.SectionConstraints},
		{"Examples", token.SectionExamples},
		{"Input", token.SectionInput},
		{"Output", token.SectionOutput},
		{"Context", token.SectionContext},
		{"Other", token.SectionOther},
	}

	// Get pricing once
	pricing, err := config.GetModelPricing(model)
	if err == nil {
		for _, s := range sections {
			tokens := sectionTotals[s.typ]
			if tokens == 0 {
				continue
			}

			usd, cny := calculateCostWithPricing(tokens, pricing)

			totalTokens += tokens
			totalUSD += usd
			totalCNY += cny

			fmt.Fprintf(w, "%s\t%d\t%s\t%s\n",
				s.name,
				tokens,
				FormatCurrency(usd, "USD"),
				FormatCurrency(cny, "CNY"),
			)
		}
	}

	w.Flush()

	// Show totals
	fmt.Printf("\n%s\t%d\t%s\t%s\n",
		"TOTAL",
		totalTokens,
		FormatCurrency(totalUSD, "USD"),
		FormatCurrency(totalCNY, "CNY"),
	)

	return nil
}

// showCostByRole shows costs grouped by role
func showCostByRole(tracked []TrackedFile, config *Config, model, currency string) error {
	fmt.Printf("Cost breakdown by role (model: %s)\n\n", model)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Aggregate by role
	roleTotals := make(map[string]int)
	roleFiles := make(map[string]int)

	for _, f := range tracked {
		role := f.Role
		if role == "" {
			role = "default"
		}
		roleTotals[role] += f.Tokens
		roleFiles[role]++
	}

	// Table header
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", "Role", "Files", "Tokens", "USD", "CNY")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", "----", "-----", "------", "---", "---")

	totalTokens := 0
	totalUSD := 0.0
	totalCNY := 0.0

	for role, tokens := range roleTotals {
		pricing, err := config.GetModelPricing(model)
		if err != nil {
			continue
		}

		usd, cny := calculateCostWithPricing(tokens, pricing)

		totalTokens += tokens
		totalUSD += usd
		totalCNY += cny

		fmt.Fprintf(w, "%s\t%d\t%d\t%s\t%s\n",
			role,
			roleFiles[role],
			tokens,
			FormatCurrency(usd, "USD"),
			FormatCurrency(cny, "CNY"),
		)
	}

	w.Flush()

	// Show totals
	fmt.Printf("\n%s\t%d\t%d\t%s\t%s\n",
		"TOTAL",
		len(tracked),
		totalTokens,
		FormatCurrency(totalUSD, "USD"),
		FormatCurrency(totalCNY, "CNY"),
	)

	return nil
}

// showAPICallEstimate shows estimated number of API calls
func showAPICallEstimate(totalTokens int, model string) {
	// Estimate calls based on typical prompt sizes
	fmt.Printf("\nEstimated API calls (%s):\n", model)

	// Assume context window of 128K tokens
	maxTokens := 128000
	fullCalls := totalTokens / maxTokens
	if totalTokens%maxTokens != 0 {
		fullCalls++
	}

	// Typical use case: 4K prompt tokens
	typicalPromptSize := 4000
	typicalCalls := totalTokens / typicalPromptSize
	if totalTokens%typicalPromptSize != 0 {
		typicalCalls++
	}

	fmt.Printf("  Full context (128K): ~%d calls\n", fullCalls)
	fmt.Printf("  Typical usage (4K): ~%d calls\n", typicalCalls)
}

// defaultPricingForModel gets pricing for a model
func defaultPricingForModel(model string) (*ModelPricing, error) {
	cm := NewCostManager()
	return cm.GetPricing(model)
}

// PriceCmd lists available model pricing
func PriceCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Load config
	config, err := LoadConfig(root)
	if err != nil {
		// Fall back to default config
		config = DefaultConfig()
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "Model", "Provider", "Input", "Output")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "-----", "--------", "-----", "------")

	for _, p := range config.Models {
		var inputPrice, outputPrice string
		if p.Currency == "USD" {
			inputPrice = fmt.Sprintf("$%.2f/M", p.InputPrice)
			outputPrice = fmt.Sprintf("$%.2f/M", p.OutputPrice)
		} else {
			inputPrice = fmt.Sprintf("¥%.2f/M", p.InputPrice)
			outputPrice = fmt.Sprintf("¥%.2f/M", p.OutputPrice)
		}

		marker := ""
		if p.Name == config.Model {
			marker = "*"
		}

		fmt.Fprintf(w, "%s%s\t%s\t%s\t%s\n",
			marker,
			p.Name,
			p.Provider,
			inputPrice,
			outputPrice,
		)
	}

	w.Flush()

	fmt.Printf("\n* = default model\n")
	fmt.Printf("Exchange rate: 1 USD = %.2f CNY\n", config.ExchangeRate)

	return nil
}
