package middlewares

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/muesli/termenv"
)

// Login Middleware
func LoginMiddleware(next wish.Middleware) wish.Middleware {
	teaHandler := func(s ssh.Session) *tea.Program {
		_, _, active := s.Pty()
		if !active {
			wish.Fatalln(s, "no active terminal, skipping")
			return nil
		}
		m := LoginModel{}

		return tea.NewProgram(m, tea.WithInput(s), tea.WithOutput(s), tea.WithAltScreen())
	}
	return bubbletea.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}

type LoginModel struct {
	username     string
	password     string
	errMsg       string
	isPassword   bool // Flag to track if we're entering the password
	isEnterPwMsg bool // Flag to show the "Enter password:" message
}

func (m LoginModel) Init() tea.Cmd {
	return nil
}

func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch msg.Type {
		case tea.KeyEnter:
			if m.isPassword {
				// Check if the entered username and password are correct
				// make a database call to check the user details
				if m.username == "user" && m.password == "pass" {
					// Proceed to fetch articles
					fmt.Println("Authenticated! Fetching articles...")

				} else {
					m.errMsg = "Invalid username or password"
					m.password = "" // Clear the password field
					m.username = ""

					m.isPassword = false
					m.isEnterPwMsg = false
				}
			} else {
				m.isPassword = true
				m.isEnterPwMsg = true
			}
		case tea.KeyBackspace, tea.KeyDelete:
			if m.isPassword {
				if len(m.password) > 0 {
					m.password = m.password[:len(m.password)-1]
				}
			} else {
				if len(m.username) > 0 {
					m.username = m.username[:len(m.username)-1]
				}
			}
		case tea.KeyCtrlC, tea.KeyEsc, tea.KeyCtrlD:
			// Quit the program when the user presses "ctrl+c", "esc", or "ctrl+d"
			return m, tea.Quit
		default:
			if m.isPassword {
				m.password += msg.String()
			} else {
				m.username += msg.String()
			}
		}
	}
	return m, nil
}

func (m LoginModel) View() string {
	var sb strings.Builder

	sb.WriteString("Please enter your credentials:\n\n")

	if m.username == "" {
		sb.WriteString("Username: ")
	} else {
		sb.WriteString(fmt.Sprintf("Username: %s\n", m.username))
	}

	if m.isPassword {
		sb.WriteString("Password: ")
		for i := 0; i < len(m.password); i++ {
			sb.WriteString("*") // Mask the password with asterisks
		}
		sb.WriteString("\n")
	} else if m.isEnterPwMsg {
		sb.WriteString("Enter password: ")
	}
	if m.errMsg != "" {
		sb.WriteString(fmt.Sprintf("\n%s\n", m.errMsg))
	}

	return sb.String()
}
