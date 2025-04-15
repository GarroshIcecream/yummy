package db

import (
	"fmt"
	"log"
	"time"

	recipes "github.com/GarroshIcecream/yummy/recipe"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Recipe struct {
	gorm.Model
	RecipeName string
}

type Category struct {
	gorm.Model
	RecipeID     uint
	CategoryName string
}

type Cuisine struct {
	gorm.Model
	RecipeID    uint
	CuisineName string
}

type Ingredients struct {
	gorm.Model
	RecipeID       uint
	IngredientName string
	Detail         string
	Amount         string
	Unit           string
}

type RecipeMetadata struct {
	gorm.Model
	RecipeID    uint
	Description string
	Author      string
	CookTime    time.Duration
	PrepTime    time.Duration
	TotalTime   time.Duration
	Quantity    string
	URL         string
	Favourite   bool
	Rating      int8
}

type Instructions struct {
	gorm.Model
	RecipeID    uint
	Description string
}

type CookBook struct {
	conn *gorm.DB
}

func (c *CookBook) GetModels() []any {
	return []any{
		&Recipe{},
		&Category{},
		&Cuisine{},
		&RecipeMetadata{},
		&Instructions{},
		&Ingredients{},
	}
}

func (c CookBook) New() (*CookBook, error) {
	db_con, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &CookBook{conn: db_con}, nil
}

func NewCookBook() (*CookBook, error) {
	cookbook := &CookBook{}
	return cookbook.New()
}

func (c *CookBook) Open() error {
	if err := c.conn.AutoMigrate(c.GetModels()...); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func (c *CookBook) RandomRecipe() Recipe {
	var recipe Recipe
	c.conn.Take(&recipe)
	return recipe
}

func (c *CookBook) DeleteRecipe(recipeID uint) error {
	if err := c.conn.Delete(&Recipe{}, recipeID).Error; err != nil {
		return fmt.Errorf("failed to delete recipe: %w", err)
	}
	return nil
}

func (c *CookBook) RecipeByName(recipe_name string) (Recipe, error) {
	var recipe Recipe
	result := c.conn.First(&recipe, "RecipeName = ?", recipe_name)

	return recipe, result.Error
}

func (c *CookBook) CreateNewRecipe(recipe_name string) {
	c.conn.Create(&Recipe{RecipeName: recipe_name})
}

type RecipeWithDescription struct {
	RecipeID             uint
	RecipeName           string
	FormattedDescription string
}

func (i RecipeWithDescription) Title() string       { return i.RecipeName }
func (i RecipeWithDescription) Description() string { return i.FormattedDescription }
func (i RecipeWithDescription) FilterValue() string {
	return fmt.Sprintf("%s - %s", i.RecipeName, i.FormattedDescription)
}

func FormatRecipe(
	id uint,
	name string,
	author string,
	description string) RecipeWithDescription {

	author_fin := "N/A"
	if author != "" {
		author_fin = author
	}

	desc := "N/A"
	if description != "" {
		desc = description
	}

	return RecipeWithDescription{
		RecipeID:             id,
		RecipeName:           name,
		FormattedDescription: fmt.Sprintf("%s - %s", author_fin, desc),
	}

}

func (c *CookBook) AllRecipes() ([]RecipeWithDescription, error) {
	var recipes []struct {
		ID          uint
		RecipeName  string
		Author      string
		Description string
	}

	err := c.conn.
		Table("recipes").
		Select("recipes.id, recipes.recipe_name, recipe_metadata.author, recipe_metadata.description").
		Joins("LEFT JOIN recipe_metadata ON recipes.id = recipe_metadata.recipe_id").
		Order("recipes.recipe_name").
		Find(&recipes).
		Error
	if err != nil {
		return nil, err
	}

	formattedRecipes := make([]RecipeWithDescription, len(recipes))
	for i, recipe := range recipes {
		formattedRecipes[i] = FormatRecipe(
			recipe.ID,
			recipe.RecipeName,
			recipe.Author,
			recipe.Description,
		)
	}

	return formattedRecipes, nil
}

// SaveScrapedRecipe saves a scraped recipe to the database
func (c *CookBook) SaveScrapedRecipe(recipeRaw *recipes.RecipeRaw) (uint, error) {
	// Create the base recipe
	recipe := Recipe{
		RecipeName: recipeRaw.Name,
	}
	if err := c.conn.Create(&recipe).Error; err != nil {
		log.Fatalf("Failed to create recipe: %s", err)
		return 0, fmt.Errorf("failed to create recipe: %w", err)
	}

	// Save metadata
	metadata := RecipeMetadata{
		RecipeID:    recipe.ID,
		Description: recipeRaw.Description,
		Author:      recipeRaw.Author,
		CookTime:    recipeRaw.CookTime,
		PrepTime:    recipeRaw.PrepTime,
		TotalTime:   recipeRaw.TotalTime,
		Quantity:    recipeRaw.Quantity,
		URL:         recipeRaw.URL,
		Favourite:   false,
		Rating:      0,
	}
	if err := c.conn.Create(&metadata).Error; err != nil {
		log.Fatalf("Failed to create recipe metadata: %s", err)
		return 0, fmt.Errorf("failed to create recipe metadata: %w", err)
	}

	// Save ingredients with parsed details
	for _, ingredient := range recipeRaw.Ingredients {
		ing := Ingredients{
			RecipeID:       recipe.ID,
			IngredientName: ingredient.Name,
			Detail:         ingredient.Details,
			Amount:         ingredient.Amount,
			Unit:           ingredient.Unit,
		}
		if err := c.conn.Create(&ing).Error; err != nil {
			log.Fatalf("Failed to create recipe ingredient: %s", err)
			return 0, fmt.Errorf("failed to create ingredient: %w", err)
		}
	}

	// Save instructions
	for _, instruction := range recipeRaw.Instructions {
		inst := Instructions{
			RecipeID:    recipe.ID,
			Description: instruction,
		}
		if err := c.conn.Create(&inst).Error; err != nil {
			return 0, fmt.Errorf("failed to create instruction: %w", err)
		}
	}

	// Save categories
	for _, categoryName := range recipeRaw.Categories {
		category := Category{
			RecipeID:     recipe.ID,
			CategoryName: categoryName,
		}
		if err := c.conn.Create(&category).Error; err != nil {
			return 0, fmt.Errorf("failed to create category: %w", err)
		}
	}

	return recipe.ID, nil
}

// GetFullRecipe retrieves a complete recipe with all its related data
func (c *CookBook) GetFullRecipe(recipeID uint) (*recipes.RecipeRaw, error) {
	// Get the base recipe
	fmt.Printf("Fetching recipe with ID: %d\n", recipeID)
	var recipe Recipe
	if err := c.conn.First(&recipe, recipeID).Error; err != nil {
		return nil, fmt.Errorf("recipe not found: %w", err)
	}

	// Get metadata
	var metadata RecipeMetadata
	if err := c.conn.Where("recipe_id = ?", recipe.ID).First(&metadata).Error; err != nil {
		return nil, fmt.Errorf("metadata not found: %w", err)
	}

	// Get ingredients
	var ingredients []Ingredients
	if err := c.conn.Where("recipe_id = ?", recipe.ID).Find(&ingredients).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch ingredients: %w", err)
	}

	// Get instructions
	var instructions []Instructions
	if err := c.conn.Where("recipe_id = ?", recipe.ID).Find(&instructions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch instructions: %w", err)
	}

	// Get categories
	var categories []Category
	if err := c.conn.Where("recipe_id = ?", recipe.ID).Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}

	// Convert to RecipeRaw
	recipeRaw := &recipes.RecipeRaw{
		Name:        recipe.RecipeName,
		Description: metadata.Description,
		Author:      metadata.Author,
		CookTime:    metadata.CookTime,
		PrepTime:    metadata.PrepTime,
		TotalTime:   metadata.TotalTime,
		Quantity:    metadata.Quantity,
		URL:         metadata.URL,
	}

	// Convert ingredients
	recipeRaw.Ingredients = make([]recipes.Ingredient, len(ingredients))
	for i, ing := range ingredients {
		recipeRaw.Ingredients[i] = recipes.Ingredient{
			Amount:  ing.Amount,
			Unit:    ing.Unit,
			Name:    ing.IngredientName,
			Details: ing.Detail,
		}
	}

	// Convert instructions
	recipeRaw.Instructions = make([]string, len(instructions))
	for i, inst := range instructions {
		recipeRaw.Instructions[i] = inst.Description
	}

	// Convert categories
	recipeRaw.Categories = make([]string, len(categories))
	for i, cat := range categories {
		recipeRaw.Categories[i] = cat.CategoryName
	}

	return recipeRaw, nil
}
