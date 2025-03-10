package main

import (
	"fmt"

	recipe "github.com/kkyr/go-recipe/pkg/recipe"
)

func main() {

	// _, err := db.NewCookBook()
	// if err != nil {
	// 	fmt.Println("Error creating CookBook:", err)
	// }

	// p := tea.NewProgram(models.NewModel())

	// if _, err := p.Run(); err != nil {
	// 	fmt.Printf("Alas, there's been an error: %v", err)
	// 	os.Exit(1)
	// }

	url := "https://blog.fatfreevegan.com/2013/06/kale-and-quinoa-salad-with-black-beans.html"

	recipe, err := recipe.ScrapeURL(url)
	if err != nil {
		// handle err
	}

	ingredients, _ := recipe.Ingredients()
	instructions, _ := recipe.Instructions()
	cook_time, _ := recipe.CookTime()
	author, _ := recipe.Author()

	fmt.Println("Ingredients are : ", ingredients)
	fmt.Println("Instructions are : ", instructions)
	fmt.Println("Author is : ", author)
	fmt.Println("Cook time is : ", cook_time)
}
