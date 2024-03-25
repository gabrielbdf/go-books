package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	table "github.com/charmbracelet/bubbles/table"
	textinput "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table     table.Model
	textInput textinput.Model
}

type person struct {
	Name string `json:"name"`
	Age  string `json:"age"`
}

var people []person

func initialModel() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Nome, Idade"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return ti
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
				m.textInput.Focus()
			} else {
				m.table.Focus()
				m.textInput.Blur()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "backspace":
			if m.table.Focused() {
				row := m.table.SelectedRow()
				if len(row) > 0 {
					name := row[0]
					indexToDelete := -1
					for i := 0; i < len(people); i++ {
						if name == people[i].Name {
							indexToDelete = i
						}
					}
					if indexToDelete > -1 {
						people = append(people[:indexToDelete], people[indexToDelete+1:]...)
						m.table.SetRows(getTableRows())
						m.table.Update(msg)
						saveDatabase()
					}
				}
			}
		case
			"enter":
			if m.textInput.Focused() {
				values := strings.Split(m.textInput.Value(), ",")
				age := " "
				if len(values) >= 2 {
					age = values[1]
				}
				people = append(people, person{Name: values[0], Age: age})
				m.table.SetRows(getTableRows())
				m.table.Update(msg)
				m.textInput.SetValue("")
				saveDatabase()
			}
		}
	}

	var tableUpdate tea.Cmd
	var textUpdate tea.Cmd
	m.table, tableUpdate = m.table.Update(msg)
	m.textInput, textUpdate = m.textInput.Update(msg)
	cmd = tea.Batch(tableUpdate, textUpdate)
	return m, cmd
}

func (m model) View() string {
	return fmt.Sprintf(
		"\n%s\n\n%s\n\n%s",
		baseStyle.Render(m.table.View())+"\n",
		m.textInput.View(),
		"(esc to alternate, directional to navigate, enter to create)",
	) + "\n"
}

func getTableRows() []table.Row {
	rows := []table.Row{}
	for _, p := range people {
		rows = append(rows, table.Row{p.Name, p.Age})
	}
	return rows
}

func saveDatabase() {
	file, err := os.Create("./data/data.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	jsonWriter := json.NewEncoder(file)
	err = jsonWriter.Encode(people)
	if err != nil {
		panic(err)
	}
}

func main() {

	database, err := os.ReadFile("./data/data.json")
	if err != nil {
		panic(err)
	}
	buffer := bytes.NewBuffer(database)
	json.NewDecoder(buffer).Decode(&people)

	columns := []table.Column{
		{Title: "Name", Width: 15},
		{Title: "Age", Width: 15},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(getTableRows()),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := model{t, initialModel()}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

}
