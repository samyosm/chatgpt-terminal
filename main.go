package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/theckman/yacspin"
	"gopkg.in/yaml.v2"
)

const CHAT_GPT_URL = "https://api.openai.com/v1/chat/completions"
const OPENAI_MODEL = "gpt-3.5-turbo"

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type PostBody struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type ResponseBody struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Choices []struct {
		Index        int     `json:"index"`
		Message      message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type Config struct {
	ApiKey string `yaml:"apikey"`
}

func GetApiKey() string {
	home, _ := os.UserHomeDir()
	f, err := os.Open(home + "/.config/chatgpt-terminal/config.yml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		panic(err)
	}

	return cfg.ApiKey

}

func main() {

	apiKey := GetApiKey()

	var messages []message

	messages = append(messages, message{
		Role:    "system",
		Content: "You are talking inside a shell terminal",
	})

	for {
		fmt.Print("You: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()

		name := scanner.Text()

		messages = append(messages, message{
			Role:    "user",
			Content: name,
		})

		cfg := yacspin.Config{
			Frequency:       100 * time.Millisecond,
			CharSet:         yacspin.CharSets[11],
			Suffix:          " ChatGPT",
			SuffixAutoColon: true,
			Message:         "waiting for answer",
			StopCharacter:   "✓",
			ColorAll:        true,
			StopColors:      []string{"fgGreen"},
		}

		spinner, err := yacspin.New(cfg)
		if err != nil {
			panic(err)
		}

		spinner.Start()

		postbody := &PostBody{
			Model:    OPENAI_MODEL,
			Messages: messages,
		}

		// JSON body
		body, err := json.Marshal(postbody)
		if err != nil {
			panic(err)
		}
		// Create a HTTP post request
		r, err := http.NewRequest("POST", CHAT_GPT_URL, bytes.NewBuffer(body))
		if err != nil {
			panic(err)
		}
		r.Header.Add("Content-Type", "application/json")
		r.Header.Add("Authorization", "Bearer "+apiKey)

		client := &http.Client{}
		res, err := client.Do(r)
		if err != nil {
			panic(err)
		}

		responseBody := &ResponseBody{}
		derr := json.NewDecoder(res.Body).Decode(responseBody)
		if derr != nil {
			panic(derr)
		}
		res.Body.Close()

		generated := responseBody.Choices[0].Message
		messages = append(messages, generated)

		spinner.StopMessage(generated.Content)
		spinner.Stop()
	}
}
