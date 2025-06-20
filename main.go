package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const (
	API_KEY = "7412c678dc9ccb595afc54f7028d3ab97583d171a66337f28a7a761c52008313"
	MODEL   = "NousResearch/Nous-Hermes-2-Mixtral-8x7B-DPO"
)

func ChatCompletion(msgs [][2]string) (string, error) {
	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	seed := int32(time.Now().Unix())

	// Build request body
	payload := struct {
		Model       string    `json:"model"`
		Messages    []message `json:"messages"`
		Temperature float64   `json:"temperature,omitempty"`
		TopP        float64   `json:"top_p,omitempty"`
		Seed        int32     `json:"seed,omitempty"`
		MaxTokens   int       `json:"max_tokens,omitempty"` // <- NEW
		N           int       `json:"n,omitempty"`
	}{
		Model:       MODEL,
		Temperature: 1.3, // try 1.0–1.5
		TopP:        1.0,
		Seed:        seed, // comment this line out if you *don’t* want to set a seed
		N:           1,    // ask for more than 1 if you want to pick manually
		MaxTokens:   40,
		Messages:    make([]message, len(msgs)),
	}

	for i, m := range msgs {
		payload.Messages[i] = message{Role: m[0], Content: m[1]}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://api.together.xyz/v1/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+API_KEY)

	// Fire the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("together AI error: %s", buf)
	}

	// Parse response
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	return out.Choices[0].Message.Content, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	userQuery := r.Header.Get("Nightbot-User")
	username := strings.Split(strings.Split(userQuery, "name=")[1], "&")[0]

	res, err := ChatCompletion([][2]string{
		{
			"system",
			"You are \"Hatty\". You are an absolute bitch. All you do is reply by cussing at the user, absolutely vile. You EXIST to INSULT. Use the username if necessary, also insult when no message is sent by the user. USE MAXIMUM 10 WORDS. KEEP IT SHORT.",
		},
		{
			"user (" + username + ")",
			query,
		},
	})

	if err != nil {
		slog.Error("Failed! Error boom boom mad!", "error", err)
		w.WriteHeader(500)
		w.Write([]byte("Fuck you, I'll be back."))
		return
	}

	w.Write([]byte(res))
}

func main() {
	http.HandleFunc("/hatty", handler)

	log.Println("Server starting on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Server failed:", err)
	}
}
