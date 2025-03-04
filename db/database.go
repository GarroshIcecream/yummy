package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Recipe struct {
	gorm.Model
	RecipeName string
}

type RecipeCategory struct {
	gorm.Model
	CategoryName string
}

type RecipeMetadata struct {
	Recipe
	author string
}

type RecipeSteps struct {
	gorm.Model
	Recipe
	Description string
}

type CookBook struct {
	conn *gorm.DB
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
	err := c.conn.AutoMigrate(&Recipe{}, &RecipeCategory{}, &RecipeMetadata{}, &RecipeSteps{})
	return err
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
