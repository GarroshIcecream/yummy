package styles

import (
	list "github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

func GetDelegateStyles() list.DefaultItemStyles {
	normalTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 2)

	normalDesc := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0")).
		Padding(0, 2)

	selectedColor := lipgloss.Color("#FF6B6B")
	selectedTitle := lipgloss.NewStyle().
		Foreground(selectedColor).
		BorderLeftForeground(selectedColor).
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		Padding(0, 2)

	selectedDesc := selectedTitle.
		Foreground(lipgloss.Color("#FFB6B6"))

	dimmedTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0")).
		Padding(0, 2)

	dimmedDesc := dimmedTitle.
		Foreground(lipgloss.Color("#A0A0A0"))

	filterMatch := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Underline(true)

	return list.DefaultItemStyles{
		NormalTitle:   normalTitle,
		NormalDesc:    normalDesc,
		SelectedTitle: selectedTitle,
		SelectedDesc:  selectedDesc,
		DimmedTitle:   dimmedTitle,
		DimmedDesc:    dimmedDesc,
		FilterMatch:   filterMatch,
	}
}

func GetListStyles() list.Styles {

	// Make the title pop with a subtle glow effect
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1)

	// Style pagination with dots
	paginationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6")).
		PaddingLeft(2)

	// Style the help text to be more subtle
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0")).
		Padding(1, 0, 0, 2)

	// Make the filter prompt stand out
	filterPrompt := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(true)

	// Style the filter cursor
	filterCursor := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#FF6B6B"))

	// Style the "no items" message
	noItems := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0")).
		Italic(true)

	// Style the pagination dots
	activePaginationDot := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		SetString("●")

	inactivePaginationDot := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6")).
		SetString("○")

	dividerDot := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6")).
		SetString(" • ")

	// Style the status bar
	statusBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0")).
		Padding(0, 0, 1, 2)

	// Style the status bar when filtering
	statusBarActiveFilter := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(true)

	// Style the filter count
	statusBarFilterCount := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6"))

	spinnerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6"))

	defaultFilterCharacterMatch := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6"))

	statusEmpty := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6"))

	arabicPagination := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6"))

	return list.Styles{
		TitleBar:                    titleStyle,
		Title:                       titleStyle,
		Spinner:                     spinnerStyle,
		FilterPrompt:                filterPrompt,
		FilterCursor:                filterCursor,
		DefaultFilterCharacterMatch: defaultFilterCharacterMatch,
		StatusBar:                   statusBar,
		StatusEmpty:                 statusEmpty,
		StatusBarActiveFilter:       statusBarActiveFilter,
		StatusBarFilterCount:        statusBarFilterCount,
		NoItems:                     noItems,
		PaginationStyle:             paginationStyle,
		HelpStyle:                   helpStyle,
		ActivePaginationDot:         activePaginationDot,
		InactivePaginationDot:       inactivePaginationDot,
		ArabicPagination:            arabicPagination,
		DividerDot:                  dividerDot,
	}
}
