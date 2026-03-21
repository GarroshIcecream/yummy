package edit

import (
	"reflect"
	"testing"
	"time"
)

func TestParseDurationInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{name: "empty", input: "", want: 0},
		{name: "compact", input: "1h30m", want: time.Hour + 30*time.Minute},
		{name: "spaced", input: "1h 30m", want: time.Hour + 30*time.Minute},
		{name: "minutes words", input: "45 minutes", want: 45 * time.Minute},
		{name: "invalid", input: "soon", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDurationInput(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseIngredientsText(t *testing.T) {
	raw := "For the sauce:\n2 tbsp olive oil\n4 cloves garlic (minced)\n\n1 pound spaghetti"

	got, err := parseIngredientsText(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("got %d ingredients, want 3", len(got))
	}

	if got[0].Group != "For the sauce" {
		t.Fatalf("got group %q, want %q", got[0].Group, "For the sauce")
	}

	if got[0].Amount != "2" || got[0].Unit != "tbl" || got[0].Name != "olive oil" {
		t.Fatalf("unexpected first ingredient: %#v", got[0])
	}

	if got[1].Details != "minced" {
		t.Fatalf("got details %q, want %q", got[1].Details, "minced")
	}

	if got[2].Group != "For the sauce" {
		t.Fatalf("expected group to persist, got %q", got[2].Group)
	}
}

func TestParseInstructionsText(t *testing.T) {
	raw := "1. Boil water\n- Salt the water\n3) Cook the pasta"
	got, err := parseInstructionsText(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"Boil water", "Salt the water", "Cook the pasta"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseCategories(t *testing.T) {
	got := parseCategories("Dinner, Pasta\nVegetarian, dinner")
	want := []string{"Dinner", "Pasta", "Vegetarian"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}
