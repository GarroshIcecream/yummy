package main

import (
	"fmt"

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

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to your own database for some reason, today you cook with passion only...")
	}

	// Migrates the schema
	db.AutoMigrate(&Recipe{})

	// Create
	db.Create(&Recipe{RecipeName: "Guloash"})

	// Read
	var product Recipe
	db.First(&product, "RecipeName = ?", "Guloash")

	fmt.Println("Recipe is: ", product)
	// Update - update product's price to 200
	// db.Model(&product).Update("Price", 200)
	// Update - update multiple fields
	// db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
	// db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// Delete - delete product
	// db.Delete(&product, 1)
}
