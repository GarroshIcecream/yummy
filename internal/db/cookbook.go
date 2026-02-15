package db

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/GarroshIcecream/yummy/internal/config"
	"github.com/GarroshIcecream/yummy/internal/log"
	utils "github.com/GarroshIcecream/yummy/internal/utils"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Creates new instance of CookBook struct
func NewCookBook(dbPath string, config *config.DatabaseConfig, opts ...gorm.Option) (*CookBook, error) {
	dbDir := filepath.Join(dbPath, "db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		slog.Error("Failed to create database directory", "dir", dbDir, "error", err)
		return nil, err
	}

	dbPath = filepath.Join(dbDir, config.RecipeDBName)
	_, err := os.Stat(dbPath)
	if err != nil {
		slog.Info("Database does not exist at %s, creating new database...", "dbPath", dbPath, "error", err)
	}

	dbCon, err := gorm.Open(sqlite.Open(dbPath), opts...)
	if err != nil {
		slog.Error("Error opening database", "dbPath", dbPath, "error", err)
		return nil, err
	}

	// Configure GORM to use slog logger (logs to file via slog setup, not stdout)
	dbCon.Logger = log.NewGormLogger(200*time.Millisecond, true, gormlogger.Info)

	if err := dbCon.AutoMigrate(GetCookbookModels()...); err != nil {
		slog.Error("Error migrating cookbook models", "error", err)
		return nil, err
	}

	return &CookBook{conn: dbCon}, nil
}

// GetDB returns the underlying database connection
func (c *CookBook) GetDB() *gorm.DB {
	return c.conn
}

// GetAllCategories returns list of all categories in the database
func (c *CookBook) GetAllCategories() ([]string, error) {
	var categories []string
	result := c.conn.
		Model(&Category{}).
		Distinct("category_name").
		Where("category_name != ''").
		Pluck("category_name", &categories)

	if result.Error != nil {
		slog.Error("Error fetching categories", "error", result.Error)
		return nil, result.Error
	}

	slog.Debug("Categories fetched", "categories", categories)
	return categories, nil
}

// GetAllAuthors returns list of all authors from the database
func (c *CookBook) GetAllAuthors() ([]string, error) {
	var authors []string
	if err := c.conn.Model(&RecipeMetadata{}).Distinct("author").Where("author != ''").Pluck("author", &authors).Error; err != nil {
		slog.Error("Error fetching authors", "error", err)
		return nil, err
	}

	slog.Debug("Authors fetched", "authors", authors)
	return authors, nil
}

// RecipeByName gets first matching recipe by name from the database
func (c *CookBook) RecipeByName(recipeName string) (Recipe, error) {
	var recipe Recipe
	if err := c.conn.First(&recipe, "RecipeName = ?", recipeName).Error; err != nil {
		slog.Error("Error fetching recipe by name", "recipe_name", recipeName, "error", err)
		return Recipe{}, err
	}

	slog.Debug("Recipe fetched by name", "recipe_name", recipeName, "id", recipe.ID)
	return recipe, nil
}

// RandomRecipe returns a random recipe from the database
func (c *CookBook) RandomRecipe() (Recipe, error) {
	var recipe Recipe
	result := c.conn.Order("RANDOM()").Take(&recipe)
	return recipe, result.Error
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

// RecipeExistsByURL returns the recipe ID for a given source URL, or 0 if not found.
func (c *CookBook) RecipeExistsByURL(url string) (uint, error) {
	var meta RecipeMetadata
	err := c.conn.Where("url = ?", url).First(&meta).Error
	if err != nil {
		if err.Error() == "record not found" {
			return 0, nil
		}
		return 0, err
	}
	return meta.RecipeID, nil
}

// DeleteRecipe deletes a recipe from the database by ID
func (c *CookBook) DeleteRecipe(recipeID uint) error {
	slog.Debug("Starting deletion of recipe with ID", "id", recipeID)

	tx := c.conn.Begin()
	if tx.Error != nil {
		slog.Error("Error starting transaction", "error", tx.Error)
		return tx.Error
	}
	slog.Debug("Transaction started successfully")

	// Delete recipe metadata
	res := tx.Unscoped().Delete(&RecipeMetadata{}, "recipe_id = ?", recipeID)
	if res.Error != nil {
		slog.Error("Error deleting recipe metadata", "error", res.Error)
		tx.Rollback()
		return res.Error
	}

	// Delete ingredients
	res = tx.Unscoped().Delete(&Ingredients{}, "recipe_id = ?", recipeID)
	if res.Error != nil {
		slog.Error("Error deleting ingredients", "error", res.Error)
		tx.Rollback()
		return res.Error
	}

	// Delete instructions
	res = tx.Unscoped().Delete(&Instructions{}, "recipe_id = ?", recipeID)
	if res.Error != nil {
		slog.Error("Error deleting instructions", "error", res.Error)
		tx.Rollback()
		return res.Error
	}

	// Delete categories
	res = tx.Unscoped().Delete(&Category{}, "recipe_id = ?", recipeID)
	if res.Error != nil {
		slog.Error("Error deleting categories", "error", res.Error)
		tx.Rollback()
		return res.Error
	}

	// Delete cuisines
	res = tx.Unscoped().Delete(&Cuisine{}, "recipe_id = ?", recipeID)
	if res.Error != nil {
		slog.Error("Error deleting cuisines", "error", res.Error)
		tx.Rollback()
		return res.Error
	}

	// Delete the main recipe
	res = tx.Unscoped().Delete(&Recipe{}, "id = ?", recipeID)
	if res.Error != nil {
		slog.Error("Error deleting main recipe", "error", res.Error)
		tx.Rollback()
		return res.Error
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		slog.Error("Error committing transaction", "error", err)
		return err
	}
	slog.Debug("Transaction committed successfully", "id", recipeID)

	return nil
}

// CreateNewRecipe creates a new recipe in the database
func (c *CookBook) CreateNewRecipe(recipeName string) (uint, error) {
	newRecipe := Recipe{RecipeName: recipeName}
	if err := c.conn.Create(&newRecipe).Error; err != nil {
		slog.Error("Error creating recipe", "error", err)
		return 0, err
	}
	return newRecipe.ID, nil
}

// AllFavouriteRecipes returns all favourite recipes with their metadata
func (c *CookBook) AllFavouriteRecipes() ([]utils.RecipeRaw, error) {
	allRecipes, err := c.AllRecipes()
	if err != nil {
		return nil, err
	}

	var favouriteRecipes []utils.RecipeRaw
	for _, recipe := range allRecipes {
		if recipe.IsFavourite {
			favouriteRecipes = append(favouriteRecipes, recipe)
		}
	}

	return favouriteRecipes, nil
}

// AllRecipes returns all recipes with their metadata
func (c *CookBook) AllRecipes() ([]utils.RecipeRaw, error) {
	// Build the base query with JOINs to get all data in one query
	query := c.conn.Table("recipes").
		Select(`
			recipes.id,
			recipes.recipe_name,
			COALESCE(recipe_metadata.description, '') as description,
			COALESCE(recipe_metadata.author, '') as author,
			COALESCE(recipe_metadata.cook_time, 0) as cook_time,
			COALESCE(recipe_metadata.prep_time, 0) as prep_time,
			COALESCE(recipe_metadata.total_time, 0) as total_time,
			COALESCE(recipe_metadata.quantity, '') as quantity,
			COALESCE(recipe_metadata.url, '') as url,
			COALESCE(recipe_metadata.favourite, 0) as favourite
		`).
		Joins("LEFT JOIN recipe_metadata ON recipes.id = recipe_metadata.recipe_id").
		Order("recipes.recipe_name")

	// Execute the query to get all recipes with metadata
	type RecipeWithMetadata struct {
		ID          uint
		RecipeName  string
		Description string
		Author      string
		CookTime    time.Duration
		PrepTime    time.Duration
		TotalTime   time.Duration
		Quantity    string
		URL         string
		Favourite   bool
	}

	var recipesWithMetadata []RecipeWithMetadata
	if err := query.Scan(&recipesWithMetadata).Error; err != nil {
		slog.Error("Error fetching recipes with metadata", "error", err)
		return nil, err
	}

	// If no recipes found, return empty slice
	if len(recipesWithMetadata) == 0 {
		slog.Debug("No recipes found")
		return []utils.RecipeRaw{}, nil
	}

	// Collect all recipe IDs for batch category query
	recipeIDs := make([]uint, len(recipesWithMetadata))
	for i, r := range recipesWithMetadata {
		recipeIDs[i] = r.ID
	}

	// Get all categories for all recipes in one query
	var categories []struct {
		RecipeID     uint   `gorm:"column:recipe_id"`
		CategoryName string `gorm:"column:category_name"`
	}

	result := c.conn.
		Model(&Category{}).
		Select("recipe_id, category_name").
		Where("recipe_id IN ?", recipeIDs).
		Find(&categories)

	if result.Error != nil {
		slog.Error("Error fetching categories", "error", result.Error)
		return nil, result.Error
	}

	// Group categories by recipe ID
	categoriesByRecipe := make(map[uint][]string)
	for _, cat := range categories {
		categoriesByRecipe[cat.RecipeID] = append(categoriesByRecipe[cat.RecipeID], cat.CategoryName)
	}

	// Build the final result
	resultWithDescriptions := make([]utils.RecipeRaw, 0, len(recipesWithMetadata))
	for _, r := range recipesWithMetadata {
		recipeCategories, exists := categoriesByRecipe[r.ID]
		if !exists {
			recipeCategories = []string{}
		}

		recipeWithDesc := utils.RecipeRaw{
			RecipeID:          r.ID,
			RecipeName:        r.RecipeName,
			RecipeDescription: r.Description,
			IsFavourite:       r.Favourite,
			Metadata: utils.RecipeMetadata{
				Categories: recipeCategories,
				Author:     r.Author,
				CookTime:   r.CookTime,
				PrepTime:   r.PrepTime,
				TotalTime:  r.TotalTime,
				Quantity:   r.Quantity,
				URL:        r.URL,
				Favourite:  r.Favourite,
			},
		}

		resultWithDescriptions = append(resultWithDescriptions, recipeWithDesc)
	}

	slog.Debug("Recipes with descriptions fetched", "count", len(resultWithDescriptions))
	return resultWithDescriptions, nil
}

// SetFavourite sets the favourite status of a recipe
func (c *CookBook) SetFavourite(recipeID uint) (bool, error) {
	var metadata RecipeMetadata
	err := c.conn.Where("recipe_id = ?", recipeID).First(&metadata).Error
	if err != nil {
		slog.Error("Error getting metadata", "error", err)
		return false, err
	}

	newFavourite := !metadata.Favourite
	err = c.conn.Model(&RecipeMetadata{}).Where("recipe_id = ?", recipeID).Update("favourite", newFavourite).Error
	if err != nil {
		slog.Error("Error setting favourite", "error", err)
		return false, err
	}

	slog.Debug("SetFavourite completed successfully", "id", recipeID, "favourite", newFavourite)
	return newFavourite, nil
}

// SetRating sets the rating of a recipe (0-5, 0 clears the rating)
func (c *CookBook) SetRating(recipeID uint, rating int8) error {
	if rating < 0 || rating > 5 {
		return fmt.Errorf("rating must be between 0 and 5")
	}
	err := c.conn.Model(&RecipeMetadata{}).Where("recipe_id = ?", recipeID).Update("rating", rating).Error
	if err != nil {
		slog.Error("Error setting rating", "error", err)
		return err
	}
	slog.Debug("SetRating completed", "id", recipeID, "rating", rating)
	return nil
}

// SaveScrapedRecipe saves a scraped recipe to the database and returns ID
func (c *CookBook) SaveScrapedRecipe(recipeRaw *utils.RecipeRaw) (uint, error) {
	// Create the base recipe
	recipe := Recipe{
		RecipeName: recipeRaw.RecipeName,
	}
	if err := c.conn.Create(&recipe).Error; err != nil {
		slog.Error("Error creating base recipe", "error", err)
		return 0, err
	}

	// Save metadata
	metadata := RecipeMetadata{
		RecipeID:    recipe.ID,
		Description: recipeRaw.RecipeDescription,
		Author:      recipeRaw.Metadata.Author,
		CookTime:    recipeRaw.Metadata.CookTime,
		PrepTime:    recipeRaw.Metadata.PrepTime,
		TotalTime:   recipeRaw.Metadata.TotalTime,
		Quantity:    recipeRaw.Metadata.Quantity,
		URL:         recipeRaw.Metadata.URL,
		Favourite:   false,
		Rating:      0,
	}
	if err := c.conn.Create(&metadata).Error; err != nil {
		slog.Error("Error creating metadata", "error", err)
		return 0, err
	}

	// Save ingredients with parsed details
	for _, ingredient := range recipeRaw.Metadata.Ingredients {
		ing := Ingredients{
			RecipeID:       recipe.ID,
			IngredientName: ingredient.Name,
			Detail:         ingredient.Details,
			Amount:         ingredient.Amount,
			Unit:           ingredient.Unit,
			BaseName:       ingredient.BaseName,
		}
		if err := c.conn.Create(&ing).Error; err != nil {
			slog.Error("Error creating ingredient", "error", err)
			return 0, err
		}
	}

	// Save instructions
	// add step number to each instruction
	for i, instruction := range recipeRaw.Metadata.Instructions {
		inst := Instructions{
			RecipeID:    recipe.ID,
			Step:        i + 1,
			Description: instruction,
		}
		if err := c.conn.Create(&inst).Error; err != nil {
			slog.Error("Error creating instruction", "error", err)
			return 0, err
		}
	}

	// Save categories
	for _, categoryName := range recipeRaw.Metadata.Categories {
		category := Category{
			RecipeID:     recipe.ID,
			CategoryName: categoryName,
		}
		if err := c.conn.Create(&category).Error; err != nil {
			slog.Error("Error creating category", "error", err)
			return 0, err
		}
	}

	slog.Debug("Saved scraped recipe", "id", recipe.ID)
	return recipe.ID, nil
}

// UpdateRecipe updates an existing recipe in the database
func (c *CookBook) UpdateRecipe(recipeRaw *utils.RecipeRaw) error {
	tx := c.conn.Begin()
	if tx.Error != nil {
		slog.Error("Error starting transaction", "error", tx.Error)
		return tx.Error
	}
	slog.Debug("Transaction started successfully")

	// Update the base recipe
	if err := tx.Model(&Recipe{}).Where("id = ?", recipeRaw.RecipeID).Update("recipe_name", recipeRaw.RecipeName).Error; err != nil {
		tx.Rollback()
		slog.Error("Error updating recipe", "error", err)
		return err
	}

	// Update metadata
	metadata := RecipeMetadata{
		Description: recipeRaw.RecipeDescription,
		CookTime:    recipeRaw.Metadata.CookTime,
		PrepTime:    recipeRaw.Metadata.PrepTime,
		TotalTime:   recipeRaw.Metadata.TotalTime,
		Quantity:    recipeRaw.Metadata.Quantity,
		URL:         recipeRaw.Metadata.URL,
	}
	if err := tx.Model(&RecipeMetadata{}).Where("recipe_id = ?", recipeRaw.RecipeID).Updates(metadata).Error; err != nil {
		tx.Rollback()
		slog.Error("Error updating recipe metadata", "error", err)
		return err
	}

	// Delete existing ingredients
	if err := tx.Unscoped().Delete(&Ingredients{}, "recipe_id = ?", recipeRaw.RecipeID).Error; err != nil {
		tx.Rollback()
		slog.Error("Error deleting existing ingredients", "error", err)
		return err
	}

	// Add new ingredients
	for _, ingredient := range recipeRaw.Metadata.Ingredients {
		ing := Ingredients{
			RecipeID:       recipeRaw.RecipeID,
			IngredientName: ingredient.Name,
			Detail:         ingredient.Details,
			Amount:         ingredient.Amount,
			Unit:           ingredient.Unit,
			BaseName:       ingredient.BaseName,
		}
		if err := tx.Create(&ing).Error; err != nil {
			tx.Rollback()
			slog.Error("Error creating ingredient", "error", err)
			return err
		}
	}

	// Delete existing instructions
	if err := tx.Unscoped().Delete(&Instructions{}, "recipe_id = ?", recipeRaw.RecipeID).Error; err != nil {
		tx.Rollback()
		slog.Error("Error deleting existing instructions", "error", err)
		return err
	}

	// Add new instructions
	for i, instruction := range recipeRaw.Metadata.Instructions {
		inst := Instructions{
			RecipeID:    recipeRaw.RecipeID,
			Step:        i + 1,
			Description: instruction,
		}
		if err := tx.Create(&inst).Error; err != nil {
			tx.Rollback()
			slog.Error("Error creating instruction", "error", err)
			return err
		}
	}

	// Delete existing categories
	if err := tx.Unscoped().Delete(&Category{}, "recipe_id = ?", recipeRaw.RecipeID).Error; err != nil {
		tx.Rollback()
		slog.Error("Error deleting existing categories", "error", err)
		return err
	}

	// Add new categories
	for _, categoryName := range recipeRaw.Metadata.Categories {
		category := Category{
			RecipeID:     recipeRaw.RecipeID,
			CategoryName: categoryName,
		}
		if err := tx.Create(&category).Error; err != nil {
			tx.Rollback()
			slog.Error("Error creating category", "error", err)
			return err
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		slog.Error("Error committing transaction", "error", err)
		return err
	}

	slog.Debug("Transaction committed successfully")
	return nil
}

// GetFullRecipe retrieves a complete recipe with all its related data
func (c *CookBook) GetFullRecipe(recipeID uint) (*utils.RecipeRaw, error) {
	// Get the base recipe
	slog.Debug("Starting GetFullRecipe for ID", "id", recipeID)

	var recipe_raw Recipe
	if err := c.conn.First(&recipe_raw, recipeID).Error; err != nil {
		slog.Error("Error fetching base recipe", "error", err)
		return nil, err
	}

	// Get metadata
	var metadata RecipeMetadata
	if err := c.conn.Where("recipe_id = ?", recipe_raw.ID).First(&metadata).Error; err != nil {
		slog.Error("Error fetching metadata", "error", err)
		return nil, err
	}

	// Get ingredients
	var ingredients []Ingredients
	if err := c.conn.Where("recipe_id = ?", recipe_raw.ID).Find(&ingredients).Error; err != nil {
		slog.Error("Error fetching ingredients", "error", err)
		return nil, err
	}

	// Get instructions
	var instructions []Instructions
	if err := c.conn.Where("recipe_id = ?", recipe_raw.ID).Find(&instructions).Error; err != nil {
		slog.Error("Error fetching instructions", "error", err)
		return nil, err
	}

	// Get categories
	var categories []Category
	if err := c.conn.Where("recipe_id = ?", recipe_raw.ID).Find(&categories).Error; err != nil {
		slog.Error("Error fetching categories", "error", err)
		return nil, err
	}

	// Convert instructions
	instructionDescriptions := make([]string, len(instructions))
	for i, inst := range instructions {
		instructionDescriptions[i] = inst.Description
	}

	// Convert categories
	categoryNames := make([]string, len(categories))
	for i, cat := range categories {
		categoryNames[i] = cat.CategoryName
	}

	// Convert ingredients
	parsedIngredients := make([]utils.Ingredient, len(ingredients))
	for i, ing := range ingredients {
		parsedIngredients[i] = utils.Ingredient{
			Name:     ing.IngredientName,
			Details:  ing.Detail,
			Amount:   ing.Amount,
			Unit:     ing.Unit,
			BaseName: ing.BaseName,
		}
	}

	// Convert to RecipeRaw
	recipeRaw := &utils.RecipeRaw{
		RecipeID:          recipe_raw.ID,
		RecipeName:        recipe_raw.RecipeName,
		RecipeDescription: metadata.Description,
		Metadata: utils.RecipeMetadata{
			Author:       metadata.Author,
			CookTime:     metadata.CookTime,
			PrepTime:     metadata.PrepTime,
			TotalTime:    metadata.TotalTime,
			Quantity:     metadata.Quantity,
			URL:          metadata.URL,
			Favourite:    metadata.Favourite,
			Rating:       metadata.Rating,
			CreatedAt:    metadata.CreatedAt,
			UpdatedAt:    metadata.UpdatedAt,
			Categories:   categoryNames,
			Instructions: instructionDescriptions,
			Ingredients:  parsedIngredients,
		},
	}

	slog.Debug("GetFullRecipe completed successfully", "id", recipeID)
	return recipeRaw, nil
}
