package db

import (
	"time"

	"gorm.io/gorm"
)

func GetDBModels() []any {
	return []any{
		&Recipe{},
		&Category{},
		&Cuisine{},
		&RecipeMetadata{},
		&Instructions{},
		&Ingredients{},
	}
}

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
	Step        int
	Description string
}
