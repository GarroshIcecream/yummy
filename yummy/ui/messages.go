package ui

import recipes "github.com/GarroshIcecream/yummy/yummy/recipe"

type RecipeSelectedMsg struct {
	RecipeID uint
}

type SessionStateMsg struct {
	SessionState SessionState
}

type SaveMsg struct {
	Recipe *recipes.RecipeRaw
	Err    error
}

type LoadRecipeMsg struct {
	Recipe *recipes.RecipeRaw
	Err    error
}