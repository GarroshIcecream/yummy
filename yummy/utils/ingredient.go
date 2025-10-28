package utils

import (
	"fmt"
	"regexp"
	"strings"
)

type Ingredient struct {
	Amount  string
	Unit    string
	Name    string
	Details string
}

func ParseIngredient(input string) Ingredient {
	unitsPattern := strings.Join(CorpusMeasures, "|")

	re := regexp.MustCompile(fmt.Sprintf(`^(?:(\d+(?:/\d+)?(?:\s*-\s*\d+(?:/\d+)?)?\s*))?((?:%s)\s+)?([^(]+?)(?:\s*\((.*?)\))?$`, unitsPattern))

	matches := re.FindStringSubmatch(strings.ToLower(input))
	if len(matches) < 5 {
		return Ingredient{Name: strings.TrimSpace(input)}
	}

	unit := strings.TrimSpace(matches[2])
	if unit != "" {
		if normalizedUnit, exists := CorpusMeasuresMap[unit]; exists {
			unit = normalizedUnit
		}
	}

	return Ingredient{
		Amount:  strings.TrimSpace(matches[1]),
		Unit:    unit,
		Name:    strings.TrimSpace(matches[3]),
		Details: strings.TrimSpace(matches[4]),
	}
}
