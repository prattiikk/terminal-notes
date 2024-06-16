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

type Article struct {
	ID          int
	Title       string
	Description string
	Content     string
	DateCreated time.Time
}

type articleViewModel struct {
	article Article
}

func ArticleViewMiddleware() wish.Middleware {
	teaHandler := func(s ssh.Session) *tea.Program {
		_, _, active := s.Pty()
		if !active {
			wish.Fatalln(s, "no active terminal, skipping")
			return nil
		}
		m := articleViewModel{article: fetchArticleByID(1)}
		return tea.NewProgram(m, tea.WithInput(s), tea.WithOutput(s), tea.WithAltScreen())
	}
	return bubbletea.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}

func (m articleViewModel) Init() tea.Cmd {
	return nil
}

func (m articleViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m articleViewModel) View() string {
	return fmt.Sprintf("Title: %s\n\n%s\n\nPress 'q' to quit\n", m.article.Title, m.article.Content)
}

func fetchArticleByID(id int) Article {
	articles := fetchArticles()
	for _, article := range articles {
		if article.ID == id {
			return article
		}
	}
	return Article{}
}

func fetchArticles() []Article {
	return []Article{
		{ID: 1, Title: "First Article", Description: "This is the first article.", Content: "Content of the first article", DateCreated: time.Now()},
		{ID: 2, Title: "Second Article", Description: "This is the second article.", Content: "Content of the second article", DateCreated: time.Now()},
	}
}
