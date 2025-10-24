package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	recipe "github.com/GarroshIcecream/yummy/yummy/recipe"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Creates new instance of CookBook struct
func NewCookBook(dbPath string, opts ...gorm.Option) (*CookBook, error) {
	dbPath = filepath.Join(dbPath, "cookbook.db")
	_, err := os.Stat(dbPath)
	if err != nil {
		log.Printf("Database does not exist at %s, creating new database...", dbPath)
	}

	dbCon, err := gorm.Open(sqlite.Open(dbPath), opts...)
	if err != nil {
		return nil, err
	}

	if err := dbCon.AutoMigrate(GetCookbookModels()...); err != nil {
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
	if err := c.conn.Model(&Category{}).Distinct("category_name").Where("category_name != ''").Pluck("category_name", &categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

// GetAllAuthors returns list of all authors from the database
func (c *CookBook) GetAllAuthors() ([]string, error) {
	var authors []string
	if err := c.conn.Model(&RecipeMetadata{}).Distinct("author").Where("author != ''").Pluck("author", &authors).Error; err != nil {
		return nil, err
	}
	return authors, nil
}

// RecipeByName gets first matching recipe by name from the database
func (c *CookBook) RecipeByName(recipeName string) (Recipe, error) {
	var recipe Recipe
	if err := c.conn.First(&recipe, "RecipeName = ?", recipeName).Error; err != nil {
		return Recipe{}, err
	}
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

// DeleteRecipe deletes a recipe from the database by ID
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

// CreateNewRecipe creates a new recipe in the database
func (c *CookBook) CreateNewRecipe(recipeName string) (uint, error) {
	newRecipe := Recipe{RecipeName: recipeName}
	if err := c.conn.Create(&newRecipe).Error; err != nil {
		return 0, fmt.Errorf("failed to create recipe: %w", err)
	}
	return newRecipe.ID, nil
}

// AllRecipes returns all recipes with their metadata
func (c *CookBook) AllRecipes(favourite bool) ([]recipe.RecipeWithDescription, error) {
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

	// Apply favourite filter at database level if requested
	if favourite {
		query = query.Where("recipe_metadata.favourite = ?", true)
	}

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
		return nil, err
	}

	// If no recipes found, return empty slice
	if len(recipesWithMetadata) == 0 {
		return []recipe.RecipeWithDescription{}, nil
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
	if err := c.conn.Model(&Category{}).
		Select("recipe_id, category_name").
		Where("recipe_id IN ?", recipeIDs).
		Find(&categories).Error; err != nil {
		return nil, err
	}

	// Group categories by recipe ID
	categoriesByRecipe := make(map[uint][]string)
	for _, cat := range categories {
		categoriesByRecipe[cat.RecipeID] = append(categoriesByRecipe[cat.RecipeID], cat.CategoryName)
	}

	// Build the final result
	result := make([]recipe.RecipeWithDescription, 0, len(recipesWithMetadata))
	for _, r := range recipesWithMetadata {
		recipeCategories, exists := categoriesByRecipe[r.ID]
		if !exists {
			recipeCategories = []string{}
		}

		recipeWithDesc := recipe.RecipeWithDescription{
			RecipeID:          r.ID,
			RecipeName:        r.RecipeName,
			AuthorName:        r.Author,
			RecipeDescription: r.Description,
			IsFavourite:       r.Favourite,
			Metadata: recipe.RecipeMetadata{
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

		result = append(result, recipeWithDesc)
	}

	return result, nil
}

// SetFavourite sets the favourite status of a recipe
func (c *CookBook) SetFavourite(recipeID uint) (bool, error) {
	var metadata RecipeMetadata
	err := c.conn.Where("recipe_id = ?", recipeID).First(&metadata).Error
	if err != nil {
		return false, err
	}

	newFavourite := !metadata.Favourite
	err = c.conn.Model(&RecipeMetadata{}).Where("recipe_id = ?", recipeID).Update("favourite", newFavourite).Error
	if err != nil {
		return false, err
	}

	return newFavourite, nil
}

// SaveScrapedRecipe saves a scraped recipe to the database and returns ID
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

// UpdateRecipe updates an existing recipe in the database
func (c *CookBook) UpdateRecipe(recipeRaw *recipe.RecipeRaw) error {
	if recipeRaw.ID == 0 {
		return fmt.Errorf("recipe ID is required for update")
	}

	tx := c.conn.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	// Update the base recipe
	if err := tx.Model(&Recipe{}).Where("id = ?", recipeRaw.ID).Update("recipe_name", recipeRaw.Name).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update recipe: %w", err)
	}

	// Update metadata
	metadata := RecipeMetadata{
		Description: recipeRaw.Description,
		Author:      recipeRaw.Author,
		CookTime:    recipeRaw.CookTime,
		PrepTime:    recipeRaw.PrepTime,
		TotalTime:   recipeRaw.TotalTime,
		Quantity:    recipeRaw.Quantity,
		URL:         recipeRaw.URL,
	}
	if err := tx.Model(&RecipeMetadata{}).Where("recipe_id = ?", recipeRaw.ID).Updates(metadata).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update recipe metadata: %w", err)
	}

	// Delete existing ingredients
	if err := tx.Unscoped().Delete(&Ingredients{}, "recipe_id = ?", recipeRaw.ID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete existing ingredients: %w", err)
	}

	// Add new ingredients
	for _, ingredient := range recipeRaw.Ingredients {
		ing := Ingredients{
			RecipeID:       recipeRaw.ID,
			IngredientName: ingredient.Name,
			Detail:         ingredient.Details,
			Amount:         ingredient.Amount,
			Unit:           ingredient.Unit,
		}
		if err := tx.Create(&ing).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create ingredient: %w", err)
		}
	}

	// Delete existing instructions
	if err := tx.Unscoped().Delete(&Instructions{}, "recipe_id = ?", recipeRaw.ID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete existing instructions: %w", err)
	}

	// Add new instructions
	for i, instruction := range recipeRaw.Instructions {
		inst := Instructions{
			RecipeID:    recipeRaw.ID,
			Step:        i + 1,
			Description: instruction,
		}
		if err := tx.Create(&inst).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create instruction: %w", err)
		}
	}

	// Delete existing categories
	if err := tx.Unscoped().Delete(&Category{}, "recipe_id = ?", recipeRaw.ID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete existing categories: %w", err)
	}

	// Add new categories
	for _, categoryName := range recipeRaw.Categories {
		category := Category{
			RecipeID:     recipeRaw.ID,
			CategoryName: categoryName,
		}
		if err := tx.Create(&category).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create category: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
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
