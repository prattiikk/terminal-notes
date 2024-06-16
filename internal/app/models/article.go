type Article struct {
	ID          int
	Title       string
	Description string
	Content     string
	DateCreated time.Time
}

type ArticleList struct {
	articles []Article
	selected int
}