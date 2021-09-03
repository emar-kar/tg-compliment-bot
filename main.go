package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Create a struct that mimics the webhook response body
// https://core.telegram.org/bots/api#update
type webhookReqBody struct {
	Message struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

func Handler(res http.ResponseWriter, req *http.Request) {
	body := &webhookReqBody{}
	if err := json.NewDecoder(req.Body).Decode(body); err != nil {
		log.Printf("could not decode request body: %s", err)
		return
	}
	if err := sendCompliment(body.Message.Chat.ID); err != nil {
		log.Printf("error in sending reply: %s", err)
		return
	}

	log.Println("reply sent")
}

type sendMessageReqBody struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

func sendCompliment(chatID int64) error {
	msg, err := getCompliment()
	if err != nil {
		return err
	}

	reqBody := &sendMessageReqBody{
		ChatID: chatID,
		Text:   msg.Text,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	client := http.Client{
		Timeout: 2 * time.Second,
	}
	defer client.CloseIdleConnections()

	apiKey := os.Getenv("BOT_KEY")
	if apiKey == "" {
		return errors.New("$BOT_KEY must be set")
	}

	res, err := client.Post("https://api.telegram.org/bot"+apiKey+"/sendMessage", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", res.Status)
	}

	return nil
}

type Message struct {
	Text string `json:"compliment"`
}

func getCompliment() (*Message, error) {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	defer client.CloseIdleConnections()

	resp, err := client.Get("https://complimentr.com/api")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	msg := &Message{}

	if err := json.NewDecoder(resp.Body).Decode(msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	if err := http.ListenAndServe(":"+port, http.HandlerFunc(Handler)); err != nil {
		log.Fatalln(err)
	}
}
