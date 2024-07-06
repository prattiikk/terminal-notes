// Package and imports
package middlewares

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/muesli/termenv"
)

// Styles
var docStyle = lipgloss.NewStyle().Margin(4, 10, 0)

// Define the main model struct
type model struct {
	listView     listViewModel     // Struct for list.model
	textareaView textareaViewModel // Struct for textarea.model
	viewportView viewportViewModel // Struct for viewport.model
	listItemView listItemViewModel // Struct for selected-item
	currentView  int               // Current page to render
	quitting     bool              // Flag to exit the program
}

// Define the list view model struct
type listViewModel struct {
	list         list.Model
	showSelected bool
}

// Define the textarea view model struct
type textareaViewModel struct {
	textarea     textarea.Model
	showTextArea bool
}

// Define the viewport view model struct
type viewportViewModel struct {
	viewport viewport.Model
	content  string
}

// Define the list item view model struct
type listItemViewModel struct {
	title           string
	desc            string
	content         string
	showItemContent bool
}

// Struct to hold a slice of items
type itemsMsg struct {
	items []listItemViewModel
}

// Methods to fulfill the list.Item interface
func (i listItemViewModel) FilterValue() string { return i.title }
func (i listItemViewModel) Title() string       { return i.title }
func (i listItemViewModel) Description() string { return i.desc }

// Init method
func (m model) Init() tea.Cmd {
	return fetchItems
}

/* VIEW METHODS */

// Renders the list view
func (m listViewModel) View() string {
	return docStyle.Render(m.list.View())
}

// Renders the textarea view
func (m textareaViewModel) View() string {
	textareaStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true, false, false).
		// Padding(1, 2).
		MarginTop(2).
		BorderForeground(lipgloss.Color("#d534eb")).
		// Background(lipgloss.Color("#020d14")).
		Foreground(lipgloss.Color("#eb9e34"))

	return textareaStyle.Render(m.textarea.View())
}

// Renders the viewport view
func (m viewportViewModel) View() string {
	viewportStyle := lipgloss.NewStyle().
		// Border(lipgloss.ThickBorder(), false, false, true, false).
		//Padding(1, 2).
		MarginTop(2).
		// BorderForeground(lipgloss.Color("63")).
		// Background(lipgloss.Color("#020d14")).
		Foreground(lipgloss.Color("#eb9e34"))

	return viewportStyle.Render(m.viewport.View())
}

// Renders the individual item view
func (m listItemViewModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Margin(4, 10, 0).Height(16).Width(100).Border(lipgloss.NormalBorder(), false, false, true, false).Render(m.content),
		lipgloss.NewStyle().Height(2).MarginLeft(10).MarginTop(2).Render("ctrl+a: exit alt screen"),
	)
}

// Main view function
func (m model) View() string {
	if m.quitting {
		return "exiting the ssh session"
	}

	switch m.currentView {
	case 1:
		return m.listView.View()
	case 2:
		return lipgloss.JoinHorizontal(lipgloss.Top, m.textareaView.View(), m.viewportView.View())
	case 3:
		centeredViewportStyle := lipgloss.NewStyle().
			MarginLeft(40).
			Render(m.viewportView.View())
		return centeredViewportStyle
		// return m.viewportView.View()
	default:
		return "Invalid view"
	}
}

/* UPDATE METHODS */

// Update method to handle key presses and window resizing
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.listView.list.SetSize(msg.Width-20, msg.Height-10)
		m.viewportView.viewport.Width = msg.Width / 2
		m.viewportView.viewport.Height = msg.Height - 4
		m.textareaView.textarea.SetWidth(msg.Width / 2)
		m.textareaView.textarea.SetHeight(msg.Height - 4)

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "ctrl+a":
			m.textareaView.showTextArea = !m.textareaView.showTextArea
			if m.textareaView.showTextArea {
				m.currentView = 2
			} else {
				m.currentView = 1
			}
			return m, nil

		case "ctrl+e":
			if m.textareaView.showTextArea {
				newItem := listItemViewModel{
					title: m.textareaView.textarea.Value(),
					desc:  "Description for " + m.textareaView.textarea.Value(),
				}
				m.listView.list.InsertItem(len(m.listView.list.Items()), newItem)
				m.textareaView.showTextArea = false
				m.currentView = 1
				return m, nil
			}

		case "ctrl+z":
			if m.currentView == 1 {
				i, ok := m.listView.list.SelectedItem().(listItemViewModel)
				if ok {
					m.listItemView = i
					m.currentView = 3
					m.viewportView.viewport.Style.MarginLeft(20)
				}
				return m, nil
			} else {
				m.currentView = 1
			}
		}

	case tea.MouseMsg:
		if m.currentView == 2 {
			var cmd tea.Cmd
			m.viewportView.viewport, cmd = m.viewportView.viewport.Update(msg)
			return m, cmd
		}

	case itemsMsg:
		var items []list.Item
		for _, i := range msg.items {
			items = append(items, i)
		}
		m.listView.list.SetItems(items)
		m.textareaView.textarea.Reset()
		m.currentView = 1
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
		out, _ := glamour.Render(m.textareaView.textarea.Value(), "dark")
		m.viewportView.viewport.SetContent(out)
		return m, cmd
	case 3:

		var cmd tea.Cmd
		// var width = m.viewportView.viewport.Style.GetWidth()
		// m.viewportView.viewport.Width = width * 1
		m.viewportView.viewport, cmd = m.viewportView.viewport.Update(msg)
		return m, cmd

	default:
		return m, nil
	}
}

/* ----------------------------------------------------------------------------------------------------------------------- */

// Fetch dummy items for the list (later this will be fetched from the database)
func fetchItems() tea.Msg {
	dummyItems := []listItemViewModel{
		{title: "HTML", desc: "HTML (HyperText Markup Language) is the standard markup language used to create and structure web pages.", content: "HTML Content goes here..."},
		{title: "CSS", desc: "CSS (Cascading Style Sheets) is a style sheet language used for describing the presentation of a document written in a markup language like HTML.", content: "CSS Content goes here..."},
		{title: "JavaScript", desc: "JavaScript is a programming language used to add interactivity and dynamic behavior to web pages.", content: "JavaScript Content goes here..."},
		{title: "React", desc: "React is a JavaScript library for building user interfaces. It is maintained by Facebook and a community of individual developers and companies.", content: "React Content goes here..."},
		{title: "Vue.js", desc: "Vue.js is a progressive JavaScript framework for building user interfaces. It is designed to be incrementally adoptable, and focuses on the view layer.", content: "Vue.js Content goes here..."},
	}

	return itemsMsg{items: dummyItems}
}

// ListMiddleware returns a Wish middleware that sets up the Bubble Tea program
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
		// t.SetWidth(100)
		// t.SetHeight(40)
		t.ShowLineNumbers = true
		t.Cursor.Blink = true
		t.CharLimit = 10000
		v := viewport.New(100, 40)
		v.SetContent("Viewport content goes here...")

		m := model{
			listView:     listViewModel{list: l},
			textareaView: textareaViewModel{textarea: t},
			viewportView: viewportViewModel{viewport: v},
		}

		return tea.NewProgram(m, tea.WithInput(s), tea.WithOutput(s), tea.WithAltScreen(), tea.WithMouseCellMotion())
	}
	return bubbletea.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}
