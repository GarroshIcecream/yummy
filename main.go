package main

// These imports will be used later on the tutorial. If you save the file
// now, Go might complain they are unused, but that's fine.
// You may also need to run `go mod tidy` to download bubbletea and its
// dependencies.
import (
	"database/sql"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/glebarez/go-sqlite"
)

type model struct {
	choices  []string         // items on the to-do list
	cursor   int              // which to-do list item our cursor is pointing at
	selected map[int]struct{} // which to-do items are selected
}

type Recipe struct {
	ingredients []string
	steps       []string
	metadata    []string
}

func main() {
	// p := tea.NewProgram(initialModel())
	// if _, err := p.Run(); err != nil {
	// 	fmt.Printf("Alas, there's been an error: %v", err)
	// 	os.Exit(1)
	// }

	// rows := [][]string{
	// 	{"Chinese", "您好", "你好"},
	// 	{"Japanese", "こんにちは", "やあ"},
	// 	{"Arabic", "أهلين", "أهلا"},
	// 	{"Russian", "Здравствуйте", "Привет"},
	// 	{"Spanish", "Hola", "¿Qué tal?"},
	// }

	// t := table.New().
	// Border(lipgloss.NormalBorder()).
	// BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
	// StyleFunc(func(row, col int) lipgloss.Style {
	// 		switch {
	// 		case row == 0:
	// 			return HeaderStyle
	// 		case row%2 == 0:
	// 			return EvenRowStyle
	// 		default:
	// 			return OddRowStyle
	// 		}
	// 	}).
	// 	Headers("LANGUAGE", "FORMAL", "INFORMAL").
	// 	Rows(rows...)

	// // You can also add tables row-by-row
	// t.Row("English", "You look absolutely fabulous.", "How's it going?")
	// fmt.Println(t)

	// l := list.New(
	// 	"A", list.New("Artichoke"),
	// 	"B", list.New("Baking Flour", "Bananas", "Barley", "Bean Sprouts"),
	// 	"C", list.New("Cashew Apple", "Cashews", "Coconut Milk", "Curry Paste", "Currywurst"),
	// 	"D", list.New("Dill", "Dragonfruit", "Dried Shrimp"),
	// 	"E", list.New("Eggs"),
	// 	"F", list.New("Fish Cake", "Furikake"),
	// 	"J", list.New("Jicama"),
	// 	"K", list.New("Kohlrabi"),
	// 	"L", list.New("Leeks", "Lentils", "Licorice Root"),
	// )

	// fmt.Println(l)
	db, err := sql.Open("sqlite", "./my.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer db.Close()
	fmt.Println("Connected to the SQLite database successfully.")

	// Get the version of SQLite
	var sqliteVersion string
	db.Exec(query string, args ...any)
	err = db.QueryRow("select sqlite_version()").Scan(&sqliteVersion)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(sqliteVersion)
}

func initialModel() model {
	return model{
		// Our to-do list is a grocery list
		choices: []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},

		// A map which indicates which choices are selected. We're using
		// the  map like a mathematical set. The keys refer to the indexes
		// of the `choices` slice, above.
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	// The header
	s := "What should we buy at the market?\n\n"

	// Iterate over our choices
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}
