package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	recipe "github.com/GarroshIcecream/yummy/yummy/recipe"
	"github.com/tmc/langchaingo/llms"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type CookBook struct {
	conn *gorm.DB
}

// GetDB returns the underlying database connection
func (c *CookBook) GetDB() *gorm.DB {
	return c.conn
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

func (c *CookBook) AllRecipes(favourite bool) ([]recipe.RecipeWithDescription, error) {
	var recipes []struct {
		ID          uint
		RecipeName  string
		Author      string
		Description string
		Favourite   bool
	}

	query := c.conn.
		Table("recipes").
		Select("recipes.id, recipes.recipe_name, recipe_metadata.author, recipe_metadata.description, recipe_metadata.favourite").
		Joins("LEFT JOIN recipe_metadata ON recipes.id = recipe_metadata.recipe_id")

	if favourite {
		query = query.Where("recipe_metadata.favourite = ?", favourite)
	}

	err := query.Order("recipes.recipe_name").Find(&recipes).Error
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
			rec.Favourite,
		)
	}

	return formattedRecipes, nil
}

func (c *CookBook) SetFavourite(recipe_id uint) error {
	var metadata RecipeMetadata
	err := c.conn.Where("recipe_id = ?", recipe_id).First(&metadata).Error
	if err != nil {
		return err
	}

	newFavourite := !metadata.Favourite
	err = c.conn.Model(&RecipeMetadata{}).Where("recipe_id = ?", recipe_id).Update("favourite", newFavourite).Error
	if err != nil {
		return err
	}

	return nil
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

// CreateSession creates a new chat session and returns the session ID
func (c *CookBook) CreateSession() (uint, error) {
	session := SessionHistory{}
	if err := c.conn.Create(&session).Error; err != nil {
		return 0, fmt.Errorf("failed to create session: %w", err)
	}
	return session.ID, nil
}

// SaveSessionMessage saves a message to the database
func (c *CookBook) SaveSessionMessage(sessionID uint, message string, role llms.ChatMessageType, modelName, content string, inputTokens, outputTokens, totalTokens int) error {
	// Convert ChatMessageType to string for database storage
	roleStr := string(role)

	sessionMessage := SessionMessage{
		SessionID:    sessionID,
		Message:      message,
		Role:         roleStr,
		ModelName:    modelName,
		Content:      content,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  totalTokens,
	}

	if err := c.conn.Create(&sessionMessage).Error; err != nil {
		return fmt.Errorf("failed to save session message: %w", err)
	}

	return nil
}

// GetSessionMessages retrieves all messages for a given session
func (c *CookBook) GetSessionMessages(sessionID uint) ([]SessionMessage, error) {
	var messages []SessionMessage
	if err := c.conn.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to get session messages: %w", err)
	}
	return messages, nil
}

// GetSessionStats returns statistics for a given session
func (c *CookBook) GetSessionStats(sessionID uint) (messageCount int64, totalInputTokens, totalOutputTokens int64, err error) {
	var count int64
	var inputTokens, outputTokens int64

	// Count messages
	if err := c.conn.Model(&SessionMessage{}).Where("session_id = ?", sessionID).Count(&count).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("failed to count session messages: %w", err)
	}

	// Sum input tokens
	if err := c.conn.Model(&SessionMessage{}).Where("session_id = ?", sessionID).Select("COALESCE(SUM(input_tokens), 0)").Scan(&inputTokens).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("failed to sum input tokens: %w", err)
	}

	// Sum output tokens
	if err := c.conn.Model(&SessionMessage{}).Where("session_id = ?", sessionID).Select("COALESCE(SUM(output_tokens), 0)").Scan(&outputTokens).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("failed to sum output tokens: %w", err)
	}

	return count, inputTokens, outputTokens, nil
}
