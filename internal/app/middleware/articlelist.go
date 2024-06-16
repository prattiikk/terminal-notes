package middlewares

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/muesli/termenv"
)

type articleListModel struct {
	articles []Article
	cursor   int
	selected map[int]struct{}
}

func ArticleListMiddleware() wish.Middleware {
	teaHandler := func(s ssh.Session) *tea.Program {
		_, _, active := s.Pty()
		if !active {
			wish.Fatalln(s, "no active terminal, skipping")
			return nil
		}
		// m := articleListModel{articles: fetchArticles()}
		return tea.NewProgram(initialModel(), tea.WithInput(s), tea.WithOutput(s), tea.WithAltScreen())
	}
	return bubbletea.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}

func initialModel() articleListModel {
	return articleListModel{
		// Our to-do list is a grocery list
		articles: []Article{{ID: 1, Title: "First Article", Description: "This is the first article.", Content: "Content of the first article", DateCreated: time.Now()},
			{ID: 2, Title: "Second Article", Description: "This is the second article.", Content: "Content of the second article", DateCreated: time.Now()},
		},

		// A map which indicates which choices are selected. We're using
		// the map like a mathematical set. The keys refer to the indexes
		// of the `choices` slice, above.
		selected: make(map[int]struct{}),
	}
}

func (m articleListModel) Init() tea.Cmd {
	return nil
}

func (m articleListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "j" :
			if m.cursor < len(m.articles)-1 {
				m.cursor++
			}
		case "down", "k" :
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter","space":
			// Proceed to view selected article
			// Here you could potentially switch to an article view model
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}

		}
	}
	return m, nil
}

func (m articleListModel) View() string {
	s := "Articles:\n"
	for i, article := range m.articles {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		checked := ""
		if _,ok := m.selected[i];ok{
			checked = "X"
		}
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, article.Title)
	}
	return s
}
