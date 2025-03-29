package db

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Recipe struct {
	gorm.Model
	RecipeName string
}

type Category struct {
	gorm.Model
	Recipe
	CategoryName string
}

type Cuisine struct {
	gorm.Model
	Recipe
	CuisineName string
}

type Ingredients struct {
	gorm.Model
	Recipe
	Ingredient string
	Unit       string
}

type RecipeMetadata struct {
	gorm.Model
	Recipe
	Description string
	Author      string
	CookTime    time.Duration
	PrepTime    time.Duration
	URL         string
	Favourite   bool
	Rating      int8
}

type Instructions struct {
	gorm.Model
	Recipe
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

func (c *CookBook) RecipeByName(recipe_name string) (Recipe, error) {
	var recipe Recipe
	result := c.conn.First(&recipe, "RecipeName = ?", recipe_name)

	return recipe, result.Error
}

func (c *CookBook) CreateNewRecipe(recipe_name string) {
	c.conn.Create(&Recipe{RecipeName: recipe_name})
}

func (c *CookBook) AllRecipes() ([]Recipe, error) {
	var recipes []Recipe
	c.conn.Find(&recipes)

	return recipes, nil
}

// SaveScrapedRecipe saves a scraped recipe to the database
func (c *CookBook) SaveScrapedRecipe(name string, author string, cookTime time.Duration, ingredients []string, instructions []string) error {
	// Create the base recipe
	recipe := Recipe{
		RecipeName: name,
	}
	if err := c.conn.Create(&recipe).Error; err != nil {
		return fmt.Errorf("failed to create recipe: %w", err)
	}

	// Save metadata
	metadata := RecipeMetadata{
		Recipe:    recipe,
		Author:    author,
		CookTime:  cookTime,
		Favourite: false,
		Rating:    0,
	}
	if err := c.conn.Create(&metadata).Error; err != nil {
		return fmt.Errorf("failed to create recipe metadata: %w", err)
	}

	// Save ingredients
	for _, ingredient := range ingredients {
		ing := Ingredients{
			Recipe:     recipe,
			Ingredient: ingredient,
			Unit:       "", // You might want to parse this from the ingredient string
		}
		if err := c.conn.Create(&ing).Error; err != nil {
			return fmt.Errorf("failed to create ingredient: %w", err)
		}
	}

	// Save instructions
	for _, instruction := range instructions {
		inst := Instructions{
			Recipe:      recipe,
			Description: instruction,
		}
		if err := c.conn.Create(&inst).Error; err != nil {
			return fmt.Errorf("failed to create instruction: %w", err)
		}
	}

	return nil
}
