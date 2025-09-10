package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	recipe "github.com/GarroshIcecream/yummy/yummy/recipe"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type CookBook struct {
	conn *gorm.DB
}

func NewCookBook(db_path string, gorm_opts ...gorm.Option) (*CookBook, error) {
	db_path = filepath.Join(db_path, "cookbook.db")
	_, err := os.Stat(db_path)
	if err != nil {
		log.Printf("Database does not exist at %s, creating new database...", db_path)
	}

	db_con, err := gorm.Open(sqlite.Open(db_path), gorm_opts...)
	if err != nil {
		return nil, err
	}

	if err := db_con.AutoMigrate(GetDBModels()...); err != nil {
		return nil, err
	}

	return &CookBook{conn: db_con}, nil
}

// RandomRecipe returns a random recipe from the database
func (c *CookBook) RandomRecipe() (Recipe, error) {
	var recipe Recipe
	result := c.conn.Order("RANDOM()").Take(&recipe)
	return recipe, result.Error
}

// RandomFullRecipe returns a random complete recipe with all related data
func (c *CookBook) RandomFullRecipe() (*recipe.RecipeRaw, error) {
	// First get a random recipe ID
	recipe, err := c.RandomRecipe()
	if err != nil {
		return nil, err
	}
	
	// Then get the full recipe using the existing method
	return c.GetFullRecipe(recipe.ID)
}

// HasRecipes checks if there are any recipes in the database
func (c *CookBook) HasRecipes() (bool, error) {
	result, err := c.RecipeCount()
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

// RecipeCount returns the total number of recipes in the database
func (c *CookBook) RecipeCount() (int64, error) {
	var count int64
	result := c.conn.Model(&Recipe{}).Count(&count)
	return count, result.Error
}

func (c *CookBook) DeleteRecipe(recipeID uint) error {
	log.Printf("Starting deletion of recipe with ID: %d", recipeID)

	tx := c.conn.Begin()
	if tx.Error != nil {
		log.Printf("Error starting transaction: %v", tx.Error)
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	log.Printf("Transaction started successfully")

	// Delete recipe metadata
	res := tx.Unscoped().Delete(&RecipeMetadata{}, "recipe_id = ?", recipeID)
	if res.Error != nil {
		log.Printf("Error deleting recipe metadata: %v", res.Error)
		tx.Rollback()
		return fmt.Errorf("failed to delete recipe metadata: %w", res.Error)
	}
	log.Printf("Recipe metadata deleted successfully, with rows affected: %d", res.RowsAffected)

	// Delete ingredients
	res = tx.Unscoped().Delete(&Ingredients{}, "recipe_id = ?", recipeID)
	if res.Error != nil {
		log.Printf("Error deleting ingredients: %v", res.Error)
		tx.Rollback()
		return fmt.Errorf("failed to delete ingredients: %w", res.Error)
	}
	log.Printf("Recipe ingredients deleted successfully, with rows affected: %d", res.RowsAffected)

	// Delete instructions
	res = tx.Unscoped().Delete(&Instructions{}, "recipe_id = ?", recipeID)
	if res.Error != nil {
		log.Printf("Error deleting instructions: %v", res.Error)
		tx.Rollback()
		return fmt.Errorf("failed to delete instructions: %w", res.Error)
	}
	log.Printf("Recipe instructions deleted successfully, with rows affected: %d", res.RowsAffected)

	// Delete categories
	res = tx.Unscoped().Delete(&Category{}, "recipe_id = ?", recipeID)
	if res.Error != nil {
		log.Printf("Error deleting categories: %v", res.Error)
		tx.Rollback()
		return fmt.Errorf("failed to delete categories: %w", res.Error)
	}
	log.Printf("Recipe categories deleted successfully, with rows affected: %d", res.RowsAffected)

	// Delete cuisines
	res = tx.Unscoped().Delete(&Cuisine{}, "recipe_id = ?", recipeID)
	if res.Error != nil {
		log.Printf("Error deleting cuisines: %v", res.Error)
		tx.Rollback()
		return fmt.Errorf("failed to delete cuisines: %w", res.Error)
	}
	log.Printf("Recipe cuisines deleted successfully, with rows affected: %d", res.RowsAffected)

	// Delete the main recipe
	res = tx.Unscoped().Delete(&Recipe{}, "id = ?", recipeID)
	if res.Error != nil {
		log.Printf("Error deleting main recipe: %v", res.Error)
		tx.Rollback()
		return fmt.Errorf("failed to delete recipe: %w", res.Error)
	}
	log.Printf("Main recipe deleted successfully, with rows affected: %d", res.RowsAffected)

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	log.Printf("Transaction committed successfully. Recipe %d completely deleted", recipeID)

	// Verify deletion
	var count int64
	if err := c.conn.Model(&Recipe{}).Where("id = ?", recipeID).Count(&count).Error; err != nil {
		log.Printf("Error verifying deletion: %v", err)
		return fmt.Errorf("failed to verify deletion: %w", err)
	}
	if count > 0 {
		log.Printf("Warning: Recipe %d still exists after deletion!", recipeID)
		return fmt.Errorf("recipe still exists after deletion")
	}
	log.Printf("Verified: Recipe %d no longer exists in database", recipeID)

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

func (c *CookBook) AllRecipes() ([]recipe.RecipeWithDescription, error) {
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

	formattedRecipes := make([]recipe.RecipeWithDescription, len(recipes))
	for i, rec := range recipes {
		formattedRecipes[i] = recipe.FormatRecipe(
			rec.ID,
			rec.RecipeName,
			rec.Author,
			rec.Description,
		)
	}

	return formattedRecipes, nil
}

// SaveScrapedRecipe saves a scraped recipe to the database
func (c *CookBook) SaveScrapedRecipe(recipeRaw *recipe.RecipeRaw) (uint, error) {
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
	// add step number to each instruction
	for i, instruction := range recipeRaw.Instructions {
		inst := Instructions{
			RecipeID:    recipe.ID,
			Step:        i + 1,
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
func (c *CookBook) GetFullRecipe(recipeID uint) (*recipe.RecipeRaw, error) {
	// Get the base recipe
	fmt.Printf("Fetching recipe with ID: %d\n", recipeID)
	log.Printf("Starting GetFullRecipe for ID: %d", recipeID)

	var recipe_raw Recipe
	if err := c.conn.First(&recipe_raw, recipeID).Error; err != nil {
		log.Printf("Error fetching base recipe: %v", err)
		return nil, fmt.Errorf("recipe not found: %w", err)
	}
	log.Printf("Base recipe loaded: %s", recipe_raw.RecipeName)

	// Get metadata
	var metadata RecipeMetadata
	if err := c.conn.Where("recipe_id = ?", recipe_raw.ID).First(&metadata).Error; err != nil {
		log.Printf("Error fetching metadata: %v", err)
		return nil, fmt.Errorf("metadata not found: %w", err)
	}
	log.Printf("Metadata loaded")

	// Get ingredients
	var ingredients []Ingredients
	if err := c.conn.Where("recipe_id = ?", recipe_raw.ID).Find(&ingredients).Error; err != nil {
		log.Printf("Error fetching ingredients: %v", err)
		return nil, fmt.Errorf("failed to fetch ingredients: %w", err)
	}
	log.Printf("Ingredients loaded: %d items", len(ingredients))

	// Get instructions
	var instructions []Instructions
	if err := c.conn.Where("recipe_id = ?", recipe_raw.ID).Find(&instructions).Error; err != nil {
		log.Printf("Error fetching instructions: %v", err)
		return nil, fmt.Errorf("failed to fetch instructions: %w", err)
	}
	log.Printf("Instructions loaded: %d items", len(instructions))

	// Get categories
	var categories []Category
	if err := c.conn.Where("recipe_id = ?", recipe_raw.ID).Find(&categories).Error; err != nil {
		log.Printf("Error fetching categories: %v", err)
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}
	log.Printf("Categories loaded: %d items", len(categories))

	// Convert to RecipeRaw
	recipeRaw := &recipe.RecipeRaw{
		ID:          recipe_raw.ID,
		Name:        recipe_raw.RecipeName,
		Description: metadata.Description,
		Author:      metadata.Author,
		CookTime:    metadata.CookTime,
		PrepTime:    metadata.PrepTime,
		TotalTime:   metadata.TotalTime,
		Quantity:    metadata.Quantity,
		URL:         metadata.URL,
	}

	// Convert ingredients
	recipeRaw.Ingredients = make([]recipe.Ingredient, len(ingredients))
	for i, ing := range ingredients {
		recipeRaw.Ingredients[i] = recipe.Ingredient{
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

	log.Printf("GetFullRecipe completed successfully")
	return recipeRaw, nil
}
