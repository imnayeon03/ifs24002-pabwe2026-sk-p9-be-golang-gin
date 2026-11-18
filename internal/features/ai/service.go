package ai

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var ErrBadAnswer = errors.New("jawaban AI tidak dapat dibaca")

type Chatter interface {
	Chat(system, user string) (string, error)
}

type Searcher interface {
	Search(query string) ([]Source, error)
}

type Service struct {
	chat   Chatter
	search Searcher
}

func NewService(c Chatter, s Searcher) *Service { return &Service{chat: c, search: s} }

const systemPrompt = `Kamu penulis blog profesional berbahasa Indonesia. Tulis artikel lengkap dalam Markdown.
Bila ada konteks hasil pencarian web, gunakan sebagai rujukan dan sebut sumbernya.
Balas HANYA dengan JSON: {"title":"...","category":"...","excerpt":"1-2 kalimat","content":"markdown"}`

// Generate: (opsional) cari fakta terkini → prompt ke LLM → draf + daftar sumber.
// Kegagalan search engine tidak menggagalkan pembuatan; artikel tetap dibuat tanpa sumber.
func (s *Service) Generate(req GenerateRequest) (*Draft, error) {
	sources := []Source{}
	if req.IncludeWebSearch {
		if found, err := s.search.Search(req.Topic); err == nil {
			sources = found
		}
	}
	var prompt strings.Builder
	fmt.Fprintf(&prompt, "Topik: %s\nKategori: %s\n", req.Topic, req.Category)
	if len(sources) > 0 {
		prompt.WriteString("\nKonteks hasil pencarian web:\n")
		for i, src := range sources {
			fmt.Fprintf(&prompt, "[%d] %s (%s): %s\n", i+1, src.Title, src.URL, src.Snippet)
		}
	}
	text, err := s.chat.Chat(systemPrompt, prompt.String())
	if err != nil {
		return nil, err
	}
	var d Draft
	if err := json.Unmarshal([]byte(ExtractJSON(text)), &d); err != nil || d.Title == "" || d.Content == "" {
		return nil, ErrBadAnswer
	}
	if d.Category == "" {
		d.Category = req.Category
	}
	d.Sources = sources
	return &d, nil
}
