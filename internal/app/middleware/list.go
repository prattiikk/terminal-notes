package middlewares

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh"
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
	formModel    *formModel
	listView     listViewModel
	textareaView textareaViewModel
	viewportView viewportViewModel
	listItemView listItemViewModel
	currentView  int
	quitting     bool
	loggedIn     bool
}

// Define the form model struct
type formModel struct {
	form  *huh.Form
	style lipgloss.Style
	state huh.FormState
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

	if m.formModel != nil && m.formModel.form != nil {
		// If the form model is not nil, initialize the form
		fmt.Println("about to start the form init")
		return m.formModel.form.Init()
	}
	return func() tea.Msg {
		// Fetch the user's list items based on the username stored in the form
		username := m.formModel.form.GetString("username")
		// password := m.formModel.form.GetString("password")
		fmt.Println(username)
		return fetchItems(username)
	}

}

/* VIEW METHODS */
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

// // Renders the login form view
// func (m model) View() string {
// 	if m.quitting {
// 		return "exiting the ssh session"
// 	}

// 	// if m.formModel == nil {
// 	// 	return "Starting..."
// 	// }

// 	if m.formModel.state == huh.StateCompleted {
// 		return m.formModel.style.Render("Welcome, " + m.formModel.form.GetString("username") + "!")
// 	}
// 	switch m.currentView {
// 	case 1:
// 		return m.listView.View()
// 	case 2:
// 		return lipgloss.JoinHorizontal(lipgloss.Top, m.textareaView.View(), m.viewportView.View())
// 	case 3:
// 		centeredViewportStyle := lipgloss.NewStyle().
// 			MarginLeft(40).
// 			Render(m.viewportView.View())
// 		return centeredViewportStyle
// 		// return m.viewportView.View()
// 	default:
// 		return m.formModel.form.View()

// 	}

// }

// Renders the login form view
func (m model) View() string {
	if m.quitting {
		return "exiting the ssh session"
	}

	// Prioritize the current view after form submission
	if m.loggedIn {
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
		default:
			return m.listView.View() // Default to list view if logged in
		}
	}

	// If the form is still active (not submitted), render the form view
	return m.formModel.form.View()
}

/* UPDATE METHODS */

// Update method to handle key presses and window resizing
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Update the form if it's not nil
	if m.formModel != nil {
		f, cmd := m.formModel.form.Update(msg)
		m.formModel.form = f.(*huh.Form)
		m.formModel.state = m.formModel.form.State
		cmds = append(cmds, cmd)
	}

	// Handle the form state and user login status
	if m.formModel != nil {
		switch m.formModel.state {
		case huh.StateAborted:
			return m, tea.Quit

		case huh.StateCompleted:
			if !m.loggedIn {
				// Successfully logged in; redirect to the list view
				username := m.formModel.form.GetString("username")
				fmt.Println("form is in update state with values:", username)
				m.currentView = 1
				m.loggedIn = true

				// Command to fetch the user's list items
				cmd := func() tea.Msg {
					return fetchItems(username)
				}
				cmds = append(cmds, cmd)

				// Return the updated model and combined commands
				return m, tea.Batch(cmds...)
			}
		}
	}

	// Handle messages for resizing and input events
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Adjust the sizes of the views based on window size
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
				if i, ok := m.listView.list.SelectedItem().(listItemViewModel); ok {
					m.listItemView = i
					m.currentView = 3
					m.viewportView.viewport.Style.MarginLeft(20)
				}
				return m, nil
			}
			m.currentView = 1
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

	// Update the current view based on the view state
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
		m.viewportView.viewport, cmd = m.viewportView.viewport.Update(msg)
		return m, cmd

	default:
		return m, tea.Batch(cmds...)
	}
}

/* ----------------------------------------------------------------------------------------------------------------------- */

// Fetch dummy items for the list (later this will be fetched from the database)
func fetchItems(username string) tea.Msg {
	// Make an API call or database query to fetch the user's list items
	userItems := []listItemViewModel{
		{title: "User Item 1", desc: "Description for User Item 1", content: "User Item 1 content"},
		{title: "User Item 2", desc: "Description for User Item 2", content: "User Item 2 content"},
		{title: "User Item 3", desc: "Description for User Item 3", content: "User Item 3 content"},
	}

	return itemsMsg{items: userItems}
}

// ListMiddleware returns a Wish middleware that sets up the Bubble Tea program
func ListMiddleware() wish.Middleware {
	teaHandler := func(s ssh.Session) *tea.Program {
		_, _, active := s.Pty()
		if !active {
			wish.Fatalln(s, "no active terminal, skipping")
			return nil
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("Username").Key("username"),
				huh.NewInput().Title("Password").EchoMode(huh.EchoModePassword),
			),
		)

		style := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(1, 2).
			BorderForeground(lipgloss.Color("#444444")).
			Foreground(lipgloss.Color("#7571F9"))

		l := list.New([]list.Item{}, list.NewDefaultDelegate(), 6, 24)
		l.Title = "your notes -> "
		t := textarea.New()
		t.Placeholder = "Enter some text…"
		t.Focus()
		t.ShowLineNumbers = true
		t.Cursor.Blink = true
		t.CharLimit = 10000
		v := viewport.New(100, 40)
		v.SetContent("Viewport content goes here…")
		m := model{
			formModel: &formModel{
				form:  form,
				style: style,
			},
			listView:     listViewModel{list: l},
			textareaView: textareaViewModel{textarea: t},
			viewportView: viewportViewModel{viewport: v},
		}

		return tea.NewProgram(m, tea.WithInput(s), tea.WithOutput(s), tea.WithAltScreen(), tea.WithMouseCellMotion())
	}
	return bubbletea.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}
