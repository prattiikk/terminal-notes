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

// list interface implementing the list.model from bubble tea
type listViewModel struct {
	list         list.Model
	showSelected bool
}

// textarea interface implementing the textarea.model from the bubble tea
type textareaViewModel struct {
	textarea     textarea.Model
	showTextArea bool
}

// custom item interface for storing the title, desc, content of the list items
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

// Method to return the FilterValue for an item (used for filtering)
func (i listItemViewModel) FilterValue() string { return i.title }
func (i listItemViewModel) Title() string       { return i.title }
func (i listItemViewModel) Description() string { return i.desc }

// init method
func (m model) Init() tea.Cmd {
	return fetchItems
}

/*VIEWS METHODS IMPLEMENTED FORM TEH BUBBLE TEA ARCHITECTURE ------------------------------------------------------------------------------------------------- */

// styles to render the list module of the bubble tea goes here
func (m listViewModel) View() string {
	return docStyle.Render(m.list.View())
}

// styles to render the textarea of bubble tea goes here
func (m textareaViewModel) View() string {
	return m.textarea.View()
}

// styles to render the individual item of list goes here
func (m listItemViewModel) View() string {

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Margin(4, 10, 0).Height(16).Width(100).Border(lipgloss.NormalBorder(), false, false, true, false).Render(m.content),
		lipgloss.NewStyle().Height(2).MarginLeft(10).MarginTop(2).Render("ctrl+a: exit alt screen"),
	)
}

// our main view (how to when to render which component based on 'currentView' field of module interface )
func (m model) View() string {
	// if quitting is true simply return text exiting the terminal
	if m.quitting {
		return "exiting the ssh session"
	}

	switch m.currentView {

	// if currentView == 1 then render the list component
	case 1:
		return m.listView.View()

	// if currentView == 2 then render the textarea component
	case 2:
		return m.textareaView.View()

	// if currentView == 3 then render the selected list item
	case 3:
		return m.listItemView.View()

	// default
	default:
		return "Invalid view"
	}
}

/*UPDATE METHODS IMPLEMENTED FROM BUBBLE TEA ARCHITECTURE ------------------------------------------------------------------------------------------------------------------- */

// when to update the current module interface which contains the current state of the application
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// based on the keypressed on the client side we update the module (current state of application)
	switch msg := msg.(type) {

	// when the terminal's width and height is changed then update the height and weight styles of the components
	case tea.WindowSizeMsg:
		m.listView.list.SetSize(msg.Width-20, msg.Height-10)
		return m, nil

	// based on the pressed keys on the keyboard update the module (current state) logic :
	case tea.KeyMsg:
		switch msg.String() {

		// if ctrl+c pressed quit the terminal session
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit //current state, quit command

		// if ctrl+a presses toggle between list and textarea
		case "ctrl+a":
			m.textareaView.showTextArea = !m.textareaView.showTextArea
			if m.textareaView.showTextArea {
				m.currentView = 2 // render textarea
			} else {
				m.currentView = 1 // render list
			}
			return m, nil

		// if enter is presses :
		case "enter":
			// if enter is presses and textarea is open then update the list with textarea content
			if m.textareaView.showTextArea {
				// create a new list item with textarea submission
				newItem := listItemViewModel{
					title: m.textareaView.textarea.Value(),
					desc:  "Description for " + m.textareaView.textarea.Value(),
				}
				// insert created item into the list at the end
				m.listView.list.InsertItem(len(m.listView.list.Items()), newItem)
				// go back to list view
				m.textareaView.showTextArea = false
				m.currentView = 1 // Switch back to main view
				return m, nil

			} else if m.listItemView.showItemContent {
				// get the selected item from the list
				i, ok := m.listView.list.SelectedItem().(listItemViewModel)
				// render the selected item's content into another component we created for desplaying the notes content
				if ok {
					// saves the selected items title, desc, content into m.listItemView from i
					m.listItemView = i
					m.currentView = 3 // Switch to content view
				}
				return m, nil
			}

		case "p":
			// toggle fullscreen on/off
			if m.altscreen {
				return m, tea.ExitAltScreen
			} else {
				return m, tea.EnterAltScreen
			}
		}

	// when the init method fetches the list from api it returns a tea.msg which is of type itemsMsg struct which triggers this case
	case itemsMsg:
		var items []list.Item
		// we update the list.model using the fetched list
		for _, i := range msg.items {
			items = append(items, i)
		}
		m.listView.list.SetItems(items)
		m.textareaView.textarea.Reset()
		m.currentView = 1 // Switch to main view
		return m, nil
	}

	// based on the m.currentView update the view
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

/* ----------------------------------------------------------------------------------------------------------------------- */

// Fetch dummy items for the list / later we will fetch these from the database from api
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

// // Main
// func main() {
// 	// Initialize components
// 	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 6, 24)
// 	l.Title = "your notes -> "
// 	t := textarea.New()
// 	t.Placeholder = "write text here ......!"
// 	t.Focus()
// 	t.CharLimit = 200
// 	t.ShowLineNumbers = true

// 	m := model{
// 		listView:     listViewModel{list: l},
// 		textareaView: textareaViewModel{textarea: t},
// 	}

// 	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
// 		fmt.Println("Error running program:", err)
// 		os.Exit(1)
// 	}
// }

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
