// Package and imports
package middlewares

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/muesli/termenv"
)

// Styles
var docStyle = lipgloss.NewStyle().Margin(4, 10, 0)

// var (
// 	keywordStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Background(lipgloss.Color("235"))
// 	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Bold(true).Height(6)
// )

type model struct {
	listView     listViewModel     // struct for list.model
	textareaView textareaViewModel // struct for textarea.model
	listItemView listItemViewModel // struct for selected-item
	currentView  int               // current page to render
	quitting     bool              // not needed really
	altscreen    bool              // if we want to enter the fullscreen or not
}

// list
type listViewModel struct {
	list         list.Model
	showSelected bool
}

// textarea
type textareaViewModel struct {
	textarea     textarea.Model
	showTextArea bool
}

// item
type listItemViewModel struct {
	title           string
	desc            string
	content         string
	showItemContent bool
}

// Method to return the FilterValue for an item (used for filtering)
func (i listItemViewModel) FilterValue() string { return i.title }
func (i listItemViewModel) Title() string       { return i.title }
func (i listItemViewModel) Description() string { return i.desc }

// init method
func (m model) Init() tea.Cmd {
	return fetchItems
}

// main_view.go
func (m listViewModel) View() string {
	return docStyle.Render(m.list.View())
}

// text_area_view.go
func (m textareaViewModel) View() string {
	return m.textarea.View()
}

// content_view.go
func (m listItemViewModel) View() string {

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Margin(4, 10, 0).Height(16).Width(100).Border(lipgloss.NormalBorder(), false, false, true, false).Render(m.content),
		lipgloss.NewStyle().Height(2).MarginLeft(10).MarginTop(2).Render("ctrl+a: exit alt screen"),
	)
}

func (m model) View() string {
	if m.quitting {
		return "exiting"
	}

	switch m.currentView {
	case 1:
		return m.listView.View()
	case 2:
		return m.textareaView.View()
	case 3:
		return m.listItemView.View()
	default:
		return "Invalid view"
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.listView.list.SetSize(msg.Width-20, msg.Height-10)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "a":
			m.textareaView.showTextArea = !m.textareaView.showTextArea
			m.currentView = 2 // Switch to textarea view
			return m, nil
		case "enter":
			if m.textareaView.showTextArea {
				// Handle textarea submission
				newItem := listItemViewModel{
					title: m.textareaView.textarea.Value(),
					desc:  "Description for " + m.textareaView.textarea.Value(),
				}
				m.listView.list.InsertItem(len(m.listView.list.Items()), newItem)
				m.textareaView.showTextArea = false
				m.currentView = 1 // Switch back to main view
				return m, nil
			} else {
				// Handle item selection
				// i := m.listView.list.SelectedItem()
				//fmt.Print(i)
				// m.textareaView.showTextArea = !m.textareaView.showTextArea
				i, ok := m.listView.list.SelectedItem().(listItemViewModel)
				if ok {
					m.listItemView = i
					m.currentView = 3 // Switch to content view
				}

				return m, nil
			}
		case "ctrl+a":
			m.currentView = 1
			return m, nil
		case "p":
			// Handle alt screen
			if m.altscreen {
				return m, tea.ExitAltScreen
			} else {
				return m, tea.EnterAltScreen
			}
		}
	case itemsMsg:
		var items []list.Item
		for _, i := range msg.items {
			items = append(items, i)
		}
		m.listView.list.SetItems(items)
		m.textareaView.textarea.Reset()
		m.currentView = 1 // Switch to main view
		return m, nil
	}

	switch m.currentView {
	case 1:
		var cmd tea.Cmd
		m.listView.list, cmd = m.listView.list.Update(msg)
		return m, cmd
	case 2:
		var cmd tea.Cmd
		m.textareaView.textarea, cmd = m.textareaView.textarea.Update(msg)
		return m, cmd
	case 3:
		var cmd tea.Cmd
		m.textareaView.textarea, cmd = m.textareaView.textarea.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

// Fetch dummy items for the list
func fetchItems() tea.Msg {
	dummyItems := []listItemViewModel{
		{
			title:   "HTML",
			desc:    "HTML (HyperText Markup Language) is the standard markup language used to create and structure web pages.",
			content: "HTML Content goes here...",
		},
		{
			title:   "CSS",
			desc:    "CSS (Cascading Style Sheets) is a style sheet language used for describing the presentation of a document written in a markup language like HTML.",
			content: "CSS Content goes here...",
		},
		{
			title:   "JavaScript",
			desc:    "JavaScript is a programming language used to add interactivity and dynamic behavior to web pages.",
			content: "JavaScript Content goes here...",
		},
		{
			title:   "React",
			desc:    "React is a JavaScript library for building user interfaces. It is maintained by Facebook and a community of individual developers and companies.",
			content: "React Content goes here...",
		},
		{
			title:   "Vue.js",
			desc:    "Vue.js is a progressive JavaScript framework for building user interfaces. It is designed to be incrementally adoptable, and focuses on the view layer.",
			content: "Vue.js Content goes here...",
		},
	}

	return itemsMsg{items: dummyItems}
}

// Struct to hold a slice of items
type itemsMsg struct {
	items []listItemViewModel
}

func ListMiddleware() wish.Middleware {
	teaHandler := func(s ssh.Session) *tea.Program {
		_, _, active := s.Pty()
		if !active {
			wish.Fatalln(s, "no active terminal, skipping")
			return nil
		}
		l := list.New([]list.Item{}, list.NewDefaultDelegate(), 6, 24)
		l.Title = "your notes -> "
		t := textarea.New()
		t.Placeholder = "Enter some text..."
		t.Focus()
		t.CharLimit = 200
		t.ShowLineNumbers = true

		m := model{
			listView:     listViewModel{list: l},
			textareaView: textareaViewModel{textarea: t},
		}

		return tea.NewProgram(m, tea.WithInput(s), tea.WithOutput(s), tea.WithAltScreen())
	}
	return bubbletea.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}
