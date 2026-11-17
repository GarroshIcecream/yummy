package utils

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

// fakeLLM returns a canned reply (or error) for every request.
type fakeLLM struct {
	reply string
	err   error
}

func (f *fakeLLM) GenerateContent(_ context.Context, _ []llms.MessageContent, _ ...llms.CallOption) (*llms.ContentResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: f.reply}}}, nil
}

func (f *fakeLLM) Call(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, f, prompt, opts...)
}

func TestParseIngredientsWithLLMUsesModel(t *testing.T) {
	llm := &fakeLLM{reply: `{"amount":"2","unit":"cup","name":"flour","details":"","base_name":"flour"}`}
	got, err := ParseIngredientsWithLLM(context.Background(), []string{"2 cups flour"}, llm)
	if err != nil {
		t.Fatalf("ParseIngredientsWithLLM: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d ingredients, want 1", len(got))
	}
	if got[0].Amount != "2" || got[0].Unit != "cup" || got[0].Name != "flour" {
		t.Errorf("got %+v, want amount=2 unit=cup name=flour", got[0])
	}
}

func TestParseIngredientsWithLLMFallsBackOnModelError(t *testing.T) {
	llm := &fakeLLM{err: errors.New("boom")}
	got, err := ParseIngredientsWithLLM(context.Background(), []string{"2 cups flour"}, llm)
	if err != nil {
		t.Fatalf("ParseIngredientsWithLLM: %v", err)
	}
	want, err := ParseIngredient("2 cups flour")
	if err != nil {
		t.Fatalf("ParseIngredient: %v", err)
	}
	if len(got) != 1 || !reflect.DeepEqual(got[0], want) {
		t.Errorf("got %+v, want regex fallback %+v", got, want)
	}
}
