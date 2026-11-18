package ai

type GenerateRequest struct {
	Topic            string `json:"topic" binding:"required"`
	Category         string `json:"category"`
	IncludeWebSearch bool   `json:"include_web_search"`
}

type Source struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet,omitempty"`
}

type Draft struct {
	Title    string   `json:"title"`
	Category string   `json:"category"`
	Excerpt  string   `json:"excerpt"`
	Content  string   `json:"content"`
	Sources  []Source `json:"sources"`
}
