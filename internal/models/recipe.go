package models

import "gorm.io/gorm"

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
