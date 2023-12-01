package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ChatBot struct {
	URL     string
	APIKey  string
	Headers http.Header
	Payload Payload
}

type Payload struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	// 根据你的需求定义 Message 结构体
}

func NewChatBot() *ChatBot {
	return &ChatBot{
		URL:    "https://openai-proxy-api.pages.dev/api/v1/chat/completions",
		APIKey: "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",
		Headers: http.Header{
			"Authorization": []string{"Bearer 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"},
			"Content-Type":  []string{"application/json"},
		},
		Payload: Payload{
			Model: "gpt-4",
			// 初始化 Messages 切片
			Messages: []Message{},
		},
	}
}

func (bot *ChatBot) SendRequest() (*http.Response, error) {
	payloadBytes, err := json.Marshal(bot.Payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", bot.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}

	req.Header = bot.Headers

	client := &http.Client{}
	return client.Do(req)
}

func main() {
	bot := NewChatBot()
	resp, err := bot.SendRequest()
	if err != nil {
		// 处理错误
		fmt.Println(err)
	}
	fmt.Println(resp)
	// 处理响应
}
