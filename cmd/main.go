package main

import (
	"fmt"
	"os"
	meedoo "github.com/MeeDoo-AI/SDK_Go/meedoo"
)

func main() {
	apiKey := os.Getenv("MEEDEO_API_KEY")
	baseURL := os.Getenv("MEEDEO_BASE_URL")
	if baseURL == "" { baseURL = "https://api.meedoo.co/api" }
	if apiKey == "" { fmt.Println("MEEDEO_API_KEY not set"); os.Exit(1) }

	c := meedoo.NewClientWithBaseURL(apiKey, baseURL)

	fmt.Println("=== Test 1: Basic Chat ===")
	r, err := c.Chat.Completions().Create(&meedoo.ChatCompletionParams{
		Model: "meedoo:latest", Messages: []meedoo.Message{{Role: "user", Content: "Say hello in one word."}},
	})
	if err != nil { panic(err) }
	fmt.Println("AI says:", r.Content())

	fmt.Println("\n=== Test 2: Web Search ===")
	r2, err := c.Chat.Completions().Create(&meedoo.ChatCompletionParams{
		Model: "meedoo:latest", Messages: []meedoo.Message{{Role: "user", Content: "Bitcoin price today?"}},
		Features: map[string]bool{"web_search": true},
	})
	if err != nil { panic(err) }
	c2 := r2.Content()
	if len(c2) > 100 { c2 = c2[:100] }
	fmt.Println("AI says:", c2, "...")

	fmt.Println("\n=== Test 3: Think Mode ===")
	think := true
	r3, err := c.Chat.Completions().Create(&meedoo.ChatCompletionParams{
		Model: "meedoo:latest", Messages: []meedoo.Message{{Role: "user", Content: "What is 7*8?"}}, Think: &think,
	})
	if err != nil { panic(err) }
	c3 := r3.Content()
	if len(c3) > 100 { c3 = c3[:100] }
	fmt.Println("AI says:", c3)

	fmt.Println("\n=== Test 4: Credits ===")
	credits, _ := c.Usage.Credits()
	fmt.Printf("Balance: EUR%.2f\n", credits.CreditBalance)

	fmt.Println("\n=== ALL TESTS PASSED ===")
}
