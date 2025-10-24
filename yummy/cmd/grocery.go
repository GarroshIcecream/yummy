package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	db "github.com/GarroshIcecream/yummy/yummy/db"
	"github.com/spf13/cobra"
)

type GroceryItem struct {
	Name     string   `json:"name"`
	Amount   string   `json:"amount"`
	Unit     string   `json:"unit"`
	Details  string   `json:"details"`
	TotalQty float64  `json:"total_quantity"`
	Recipes  []string `json:"recipes"`
}

type GroceryList struct {
	Items       []GroceryItem `json:"items"`
	Recipes     []string      `json:"recipes"`
	TotalItems  int           `json:"total_items"`
	GeneratedAt string        `json:"generated_at"`
}

func init() {
	groceryCmd.Flags().StringP("format", "f", "markdown", "Output format (markdown, csv, json)")
	groceryCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	groceryCmd.Flags().BoolP("group", "g", true, "Group similar ingredients")
	groceryCmd.Flags().StringP("separator", "s", ",", "Recipe ID separator")
}

var groceryCmd = &cobra.Command{
	Use:   "grocery [recipe_ids...]",
	Short: "Generate grocery list from selected recipes",
	Long: `Generate a shopping list from selected recipes. You can specify recipe IDs
as arguments, or use comma-separated values.

Examples:
  # Generate grocery list for specific recipes
  yummy grocery 1 2 3

  # Generate grocery list with comma-separated IDs
  yummy grocery 1,2,3

  # Export to CSV format
  yummy grocery 1 2 3 --format csv --output shopping_list.csv

  # Export to JSON format
  yummy grocery 1,2,3 --format json --output groceries.json`,
	Example: `
		# Generate grocery list for recipes 1, 2, and 3
		yummy grocery 1 2 3

		# Export to CSV file
		yummy grocery 1,2,3 --format csv --output shopping.csv

		# Export to JSON with grouping disabled
		yummy grocery 1 2 3 --format json --group=false`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("at least one recipe ID is required")
		}

		// Parse recipe IDs from arguments
		var recipeIDs []uint
		for _, arg := range args {
			// Handle comma-separated values
			separator, _ := cmd.Flags().GetString("separator")
			ids := strings.Split(arg, separator)
			for _, idStr := range ids {
				idStr = strings.TrimSpace(idStr)
				if idStr == "" {
					continue
				}
				id, err := strconv.ParseUint(idStr, 10, 32)
				if err != nil {
					return fmt.Errorf("invalid recipe ID: %s", idStr)
				}
				recipeIDs = append(recipeIDs, uint(id))
			}
		}

		if len(recipeIDs) == 0 {
			return fmt.Errorf("no valid recipe IDs provided")
		}

		// Get format and output options
		format, _ := cmd.Flags().GetString("format")
		outputFile, _ := cmd.Flags().GetString("output")
		groupIngredients, _ := cmd.Flags().GetBool("group")

		// Setup app and database
		tui, err := setupApp(cmd)
		if err != nil {
			return err
		}

		// Generate grocery list
		groceryList, err := generateGroceryList(tui.Cookbook, recipeIDs, groupIngredients)
		if err != nil {
			return err
		}

		// Format and output the grocery list
		output, err := formatGroceryList(groceryList, format)
		if err != nil {
			return err
		}

		// Write to file or stdout
		if outputFile != "" {
			if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
				return fmt.Errorf("failed to write to file %s: %v", outputFile, err)
			}
			fmt.Printf("Grocery list exported to %s\n", outputFile)
		} else {
			fmt.Print(output)
		}

		return nil
	},
}

func generateGroceryList(cookbook *db.CookBook, recipeIDs []uint, groupIngredients bool) (*GroceryList, error) {
	var allIngredients []GroceryItem
	var recipeNames []string

	// Get all recipes and their ingredients
	for _, recipeID := range recipeIDs {
		recipeRaw, err := cookbook.GetFullRecipe(recipeID)
		if err != nil {
			return nil, fmt.Errorf("failed to get recipe %d: %v", recipeID, err)
		}

		recipeNames = append(recipeNames, recipeRaw.Name)

		// Convert ingredients to grocery items
		for _, ingredient := range recipeRaw.Ingredients {
			item := GroceryItem{
				Name:    ingredient.Name,
				Amount:  ingredient.Amount,
				Unit:    ingredient.Unit,
				Details: ingredient.Details,
				Recipes: []string{recipeRaw.Name},
			}
			allIngredients = append(allIngredients, item)
		}
	}

	// Group similar ingredients if requested
	if groupIngredients {
		allIngredients = groupSimilarIngredients(allIngredients)
	}

	// Sort ingredients alphabetically
	sort.Slice(allIngredients, func(i, j int) bool {
		return strings.ToLower(allIngredients[i].Name) < strings.ToLower(allIngredients[j].Name)
	})

	return &GroceryList{
		Items:       allIngredients,
		Recipes:     recipeNames,
		TotalItems:  len(allIngredients),
		GeneratedAt: fmt.Sprintf("%d", len(recipeIDs)),
	}, nil
}

func groupSimilarIngredients(items []GroceryItem) []GroceryItem {
	ingredientMap := make(map[string]*GroceryItem)

	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.Name))

		if existing, exists := ingredientMap[key]; exists {
			// Merge with existing ingredient
			existing.Recipes = append(existing.Recipes, item.Recipes...)
			// For now, just keep the first amount/unit - could be enhanced to do math
			if existing.Amount == "" && item.Amount != "" {
				existing.Amount = item.Amount
				existing.Unit = item.Unit
			}
		} else {
			// Create new entry
			ingredientMap[key] = &GroceryItem{
				Name:    item.Name,
				Amount:  item.Amount,
				Unit:    item.Unit,
				Details: item.Details,
				Recipes: item.Recipes,
			}
		}
	}

	// Convert map back to slice
	var result []GroceryItem
	for _, item := range ingredientMap {
		result = append(result, *item)
	}

	return result
}

func formatGroceryList(groceryList *GroceryList, format string) (string, error) {
	switch format {
	case "json":
		return formatGroceryListJSON(groceryList)
	case "csv":
		return formatGroceryListCSV(groceryList)
	case "markdown":
		return formatGroceryListMarkdown(groceryList)
	default:
		return "", fmt.Errorf("unsupported format: %s. Supported formats: json, csv, markdown", format)
	}
}

func formatGroceryListJSON(groceryList *GroceryList) (string, error) {
	jsonData, err := json.MarshalIndent(groceryList, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %v", err)
	}
	return string(jsonData), nil
}

func formatGroceryListCSV(groceryList *GroceryList) (string, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Write header
	if err := writer.Write([]string{"Name", "Amount", "Unit", "Details", "Recipes"}); err != nil {
		return "", err
	}

	// Write data rows
	for _, item := range groceryList.Items {
		recipes := strings.Join(item.Recipes, "; ")
		if err := writer.Write([]string{item.Name, item.Amount, item.Unit, item.Details, recipes}); err != nil {
			return "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func formatGroceryListMarkdown(groceryList *GroceryList) (string, error) {
	var buf strings.Builder

	// Header
	buf.WriteString("# 🛒 Grocery List\n\n")
	buf.WriteString(fmt.Sprintf("**Generated for %d recipes:** %s\n\n", len(groceryList.Recipes), strings.Join(groceryList.Recipes, ", ")))
	buf.WriteString(fmt.Sprintf("**Total items:** %d\n\n", groceryList.TotalItems))

	// Ingredients list
	buf.WriteString("## 📝 Shopping List\n\n")
	for i, item := range groceryList.Items {
		buf.WriteString(fmt.Sprintf("%d. **%s**", i+1, item.Name))

		if item.Amount != "" {
			buf.WriteString(fmt.Sprintf(" - %s", item.Amount))
			if item.Unit != "" {
				buf.WriteString(fmt.Sprintf(" %s", item.Unit))
			}
		}

		if item.Details != "" {
			buf.WriteString(fmt.Sprintf(" (%s)", item.Details))
		}

		if len(item.Recipes) > 1 {
			buf.WriteString(fmt.Sprintf(" - *Used in: %s*", strings.Join(item.Recipes, ", ")))
		} else if len(item.Recipes) == 1 {
			buf.WriteString(fmt.Sprintf(" - *From: %s*", item.Recipes[0]))
		}

		buf.WriteString("\n")
	}

	return buf.String(), nil
}
