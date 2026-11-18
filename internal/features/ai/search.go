package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SearchClient memanggil Search Engine: GET {endpoint}?q=...
type SearchClient struct {
	Endpoint string
	HTTP     *http.Client
}

func NewSearchClient(endpoint string) *SearchClient {
	return &SearchClient{Endpoint: endpoint, HTTP: &http.Client{Timeout: 30 * time.Second}}
}

type rawResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
	Content string `json:"content"`
}

// Search menerima bentuk {"results":[...]}, {"data":[...]} atau {"data":{"results":[...]}}.
func (s *SearchClient) Search(query string) ([]Source, error) {
	res, err := s.HTTP.Get(s.Endpoint + "?q=" + url.QueryEscape(query))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("search engine mengembalikan status %d", res.StatusCode)
	}
	var body struct {
		Results []rawResult     `json:"results"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil, err
	}
	items := body.Results
	if len(items) == 0 && len(body.Data) > 0 {
		var arr []rawResult
		var nested struct {
			Results []rawResult `json:"results"`
		}
		if json.Unmarshal(body.Data, &arr) == nil {
			items = arr
		} else if json.Unmarshal(body.Data, &nested) == nil {
			items = nested.Results
		}
	}
	out := make([]Source, 0, len(items))
	for _, it := range items {
		link := it.URL
		if link == "" {
			link = it.Link
		}
		snippet := it.Snippet
		if snippet == "" {
			snippet = it.Content
		}
		if link != "" {
			out = append(out, Source{Title: strings.TrimSpace(it.Title), URL: link, Snippet: strings.TrimSpace(snippet)})
		}
	}
	return out, nil
}
