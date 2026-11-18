package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client mengirim prompt ke OpenAI Chat Completions (kompatibel).
type Client struct {
	Endpoint, Key, Model string
	HTTP                 *http.Client
}

func NewClient(endpoint, key, model string) *Client {
	return &Client{Endpoint: strings.TrimRight(endpoint, "/"), Key: key, Model: model,
		HTTP: &http.Client{Timeout: 60 * time.Second}}
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Chat mengembalikan isi teks jawaban pertama.
func (c *Client) Chat(system, user string) (string, error) {
	body, err := json.Marshal(map[string]interface{}{
		"model":    c.Model,
		"messages": []message{{"system", system}, {"user", user}},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, c.Endpoint+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Key)
	res, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("layanan AI mengembalikan status %d", res.StatusCode)
	}
	var out struct {
		Choices []struct {
			Message message `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", errors.New("layanan AI tidak memberi jawaban")
	}
	return out.Choices[0].Message.Content, nil
}

// ExtractJSON mengambil blok JSON pertama dari teks (menoleransi pagar ```json).
func ExtractJSON(text string) string {
	text = strings.TrimSpace(text)
	start := strings.IndexAny(text, "[{")
	end := strings.LastIndexAny(text, "]}")
	if start < 0 || end < start {
		return text
	}
	return text[start : end+1]
}
