package ui

import (
	"fast-launcher/app"
	"fast-launcher/config"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)
var quitTextStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4)

type item struct {
	title   string
	desc    string
	command string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) Command() string     { return i.command }
func (i item) FilterValue() string { return i.title }

type model struct {
	list   list.Model
	choice string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" ||
			msg.String() == "esc" {
			return m, tea.Quit
		}
		if msg.String() == "enter" {
			i, ok := m.list.SelectedItem().(item)
			if ok {
				// m.choice = i.Title()

				go func() {
					app := app.App{}
					app.Run(i.Command())
				}()
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("%s? Sounds good to me.", m.choice))
	}
	return docStyle.Render(m.list.View())
}

func StartUi(configCommand []config.Config) {
	items := []list.Item{}

	for _, cc := range configCommand {
		items = append(items, item{
			title:   cc.Title,
			desc:    cc.Description,
			command: cc.Command,
		})
	}

	// listModel := list.NewDefaultDelegate()

	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "FastLauncher"
	m.list.SetShowHelp(false)
	m.list.SetShowTitle(true)
	m.list.SetShowStatusBar(false)
	keyMap := KeyMap{}
	m.list.KeyMap = keyMap.Get()

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

}
