// Package orchestrator provides cost estimation for LLM API calls.
package orchestrator

import (
	"fmt"
	"strings"

	"github.com/2mcode/2mcode/internal/team"
)

// Pricing per 1K tokens (input and output) in USD.
// Sources: provider pricing pages as of 2026. These are estimates — actual
// costs depend on the specific model version, caching, and throughput discounts.
var pricingTable = map[string]struct{ Input, Output float64 }{
	// Anthropic
	"claude-opus":        {0.015, 0.075},
	"claude-sonnet":      {0.003, 0.015},
	"claude-haiku":       {0.00025, 0.00125},

	// Google
	"gemini-1.5-pro":     {0.00125, 0.005},
	"gemini-1.5-flash":   {0.000075, 0.0003},
	"gemini-2.0-flash":   {0.0001, 0.0004},

	// OpenAI
	"gpt-4o":             {0.0025, 0.01},
	"gpt-4o-mini":        {0.00015, 0.0006},
	"o1":                 {0.015, 0.06},

	// Mistral
	"mistral-large":      {0.002, 0.006},
	"mistral-medium":     {0.0027, 0.0081},
	"codestral":          {0.001, 0.003},

	// Groq
	"llama3":             {0.0005, 0.0008},
	"mixtral":            {0.0003, 0.0006},
	"gemma2":             {0.0002, 0.0004},

	// Default for unknown models
	"default":            {0.002, 0.008},
}

// EstimateCost returns the estimated USD cost for a given token usage and model ID.
func EstimateCost(model string, inputTokens, outputTokens int) float64 {
	price := lookupPricing(model)
	inputCost := price.Input * float64(inputTokens) / 1000
	outputCost := price.Output * float64(outputTokens) / 1000
	return inputCost + outputCost
}

// lookupPricing finds the best matching pricing row for a model string.
func lookupPricing(model string) struct{ Input, Output float64 } {
	modelLower := strings.ToLower(model)
	for pattern, price := range pricingTable {
		if strings.Contains(modelLower, pattern) {
			return price
		}
	}
	return pricingTable["default"]
}

// FormatCost formats a USD cost for display.
func FormatCost(cost float64) string {
	switch {
	case cost >= 1.0:
		return fmt.Sprintf("$%.2f", cost)
	case cost >= 0.01:
		return fmt.Sprintf("$%.3f", cost)
	case cost >= 0.0001:
		return fmt.Sprintf("$%.4f", cost)
	default:
		return "$0.0000"
	}
}

// TotalCost computes estimated cost across all agents in a team.
func TotalCost(t *team.Team, tokenUsage map[string]struct{ Input, Output int }) float64 {
	total := 0.0
	for _, agent := range t.Agents {
		usage := tokenUsage[agent.Name]
		total += EstimateCost(agent.Model, usage.Input, usage.Output)
	}
	return total
}
