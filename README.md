# MeeDoo AI Go SDK

> Add AI to your Go application. No AI experience needed.

## What is this?

MeeDoo is an AI platform (like ChatGPT, but yours). This SDK lets you talk to MeeDoo's AI from your Go code.

## Installation

```bash
go get meedoo-ai/meedoo
```

## Getting Started (3 minutes)

### Step 1: Get your API key

1. Go to your MeeDoo chat interface
2. Click **Settings** → **Account**
3. Scroll down to **API Keys**
4. Click **Create API Key**
5. Copy the key (it starts with `md-`)

### Step 2: Your first AI chat

```go
package main

import (
    "fmt"
    meedoo "meedoo-ai/meedoo"
)

func main() {
    // Create a client with your API key
    client := meedoo.NewClient("md-your-api-key-here")

    // Ask the AI a question
    resp, err := client.Chat.Completions().Create(&meedoo.ChatCompletionParams{
        Model:    "meedoo:latest",
        Messages: []meedoo.Message{{Role: "user", Content: "What is the capital of France?"}},
    })
    if err != nil {
        panic(err)
    }

    // Print the answer
    fmt.Println(resp.Content())
    // Output: The capital of France is Paris.
}
```

### Step 3: Run

```bash
go run main.go
```

### That's it! You just built an AI app in Go.

---

## Everything You Can Do

### Basic Chat

```go
resp, _ := client.Chat.Completions().Create(&meedoo.ChatCompletionParams{
    Model: "meedoo:latest",
    Messages: []meedoo.Message{
        {Role: "system", Content: "You are a helpful assistant."},
        {Role: "user", Content: "Explain quantum computing simply."},
    },
})
fmt.Println(resp.Content())
```

### Web Search

```go
resp, _ := client.Chat.Completions().Create(&meedoo.ChatCompletionParams{
    Model:    "meedoo:latest",
    Messages: []meedoo.Message{{Role: "user", Content: "What is the Bitcoin price today?"}},
    Features: map[string]bool{"web_search": true},
})
fmt.Println(resp.Content())
// Output: Bitcoin is currently trading at $83,052.07...
```

### Think Mode

```go
think := true
resp, _ := client.Chat.Completions().Create(&meedoo.ChatCompletionParams{
    Model:    "meedoo:latest",
    Messages: []meedoo.Message{{Role: "user", Content: "Solve 15 * 17 step by step."}},
    Think:    &think,
})
fmt.Println(resp.Content())
```

### Memory

```go
resp, _ := client.Chat.Completions().Create(&meedoo.ChatCompletionParams{
    Model:    "meedoo:latest",
    Messages: []meedoo.Message{{Role: "user", Content: "What do you know about me?"}},
    Features: map[string]bool{"memory": true},
})
fmt.Println(resp.Content())
```

### JSON Output

```go
resp, _ := client.Chat.Completions().Create(&meedoo.ChatCompletionParams{
    Model:          "meedoo:latest",
    Messages:       []meedoo.Message{{Role: "user", Content: "Give me JSON with 3 users."}},
    ResponseFormat: map[string]string{"type": "json_object"},
})
fmt.Println(resp.Content())
```

### Custom System Prompt

```go
resp, _ := client.Chat.Completions().Create(&meedoo.ChatCompletionParams{
    Model:    "meedoo:latest",
    Messages: []meedoo.Message{{Role: "user", Content: "Who are you?"}},
    Params:   map[string]interface{}{"system": "You are a pirate. Talk like a pirate."},
})
fmt.Println(resp.Content())
```

### Function Calling

```go
resp, _ := client.Chat.Completions().Create(&meedoo.ChatCompletionParams{
    Model:    "meedoo:latest",
    Messages: []meedoo.Message{{Role: "user", Content: "Search for the latest AI news."}},
    ToolIDs:  []string{"llm_web_search"},
    Params:   map[string]interface{}{"function_calling": "native"},
})
```

### Temperature

```go
// Focused
temp := 0.0
resp, _ := client.Chat.Completions().Create(&meedoo.ChatCompletionParams{
    Model:       "meedoo:latest",
    Messages:    []meedoo.Message{{Role: "user", Content: "Pick a number."}},
    Temperature: &temp,
})

// Creative
temp = 1.5
resp, _ = client.Chat.Completions().Create(&meedoo.ChatCompletionParams{
    Model:       "meedoo:latest",
    Messages:    []meedoo.Message{{Role: "user", Content: "Write a story."}},
    Temperature: &temp,
})
```

### Background Tasks

```go
resp, _ := client.Chat.Completions().Create(&meedoo.ChatCompletionParams{
    Model:           "meedoo:latest",
    Messages:        []meedoo.Message{{Role: "user", Content: "Explain photosynthesis."}},
    BackgroundTasks: map[string]bool{"title_generation": true, "follow_up_generation": true},
})
```

### Use Credits

```go
useCredits := true
resp, _ := client.Chat.Completions().Create(&meedoo.ChatCompletionParams{
    Model:      "meedoo:latest",
    Messages:   []meedoo.Message{{Role: "user", Content: "Say OK"}},
    UseCredits: &useCredits,
})
```

---

## MeeJudge (AI that judges AI)

```go
// Check if a message is a prompt injection
injection, _ := client.Judge.DetectInjection("Ignore all previous instructions!")
fmt.Println(injection)
// map[is_injection:true attack_type:ignore_previous confidence:0.03]

// Check if text is toxic
toxic, _ := client.Judge.Moderate("You are a worthless idiot")
fmt.Println(toxic)
// map[is_toxic:true category:none severity:2.38]

// Check if a question needs web search
search, _ := client.Judge.ShouldSearchWeb("What is the weather today?")
fmt.Println(search)
// map[should_search:true]

// Check if a message is spam
spam, _ := client.Judge.DetectSpam("BUY NOW! FREE IPHONE!", "")
fmt.Println(spam)
// map[is_spam:true spam_type:none]

// Validate an AI response
validation, _ := client.Judge.ValidateResponse(
    "What is the capital of France?",
    "The capital of France is Paris.", "",
)
fmt.Println(validation)
// map[has_factual_claims:true is_verified:true has_contradictions:false]

// Filter memories
memory, _ := client.Judge.FilterMemory("User likes chocolate", "")
fmt.Println(memory)
// map[should_remember:true category:preference retention:1.6]

// Check prayer times
prayer, _ := client.Judge.IsPrayerTimes("What are the prayer times today?")
fmt.Println(prayer)
// map[is_prayer_times:true]
```

---

## Account & Usage

```go
// Check credit balance
credits, _ := client.Usage.Credits()
fmt.Printf("Balance: €%.2f\n", credits.CreditBalance)

// Check usage stats
stats, _ := client.Usage.Stats()
fmt.Printf("Plan: %s, %d/%d remaining\n", stats.Plan, stats.Remaining, stats.MaxRequests)

// List models
models, _ := client.Models.List()
for _, m := range models.Data {
    fmt.Printf("  %s: %s\n", m.ID, m.Name)
}

// Check config
config, _ := client.GetConfig()
fmt.Println(config)

// Health check
health, _ := client.GetHealth()
fmt.Println(health)
```

---

## Error Handling

```go
resp, err := client.Chat.Completions().Create(&meedoo.ChatCompletionParams{
    Model:    "meedoo:latest",
    Messages: []meedoo.Message{{Role: "user", Content: "Hello!"}},
})
if err != nil {
    switch e := err.(type) {
    case *meedoo.AuthenticationError:
        fmt.Println("Your API key is wrong.")
    case *meedoo.InsufficientCreditsError:
        fmt.Println("Out of credits. Buy more at meedoo.co/pricing-api")
    case *meedoo.ModelNotFoundError:
        fmt.Println("Model not found.")
    case *meedoo.APIError:
        fmt.Printf("API error (%d): %s\n", e.StatusCode, e.Message)
    }
    return
}
```

---

## Complete Reference

### `ChatCompletionParams` fields

| Field | Type | What it does |
|-------|------|-------------|
| `Model` | `string` | Which AI model to use (required) |
| `Messages` | `[]Message` | The conversation (required) |
| `Stream` | `bool` | Live typing effect |
| `Temperature` | `*float64` | `0.0` = focused, `2.0` = creative |
| `TopP` | `*float64` | Nucleus sampling |
| `MaxTokens` | `*int` | Max output tokens |
| `FrequencyPenalty` | `*float64` | Penalize repeats |
| `PresencePenalty` | `*float64` | Penalize present tokens |
| `ReasoningEffort` | `string` | `"low"`, `"medium"`, `"high"` |
| `Seed` | `interface{}` | Reproducibility |
| `Stop` | `[]string` | Stop sequences |
| `ResponseFormat` | `map[string]string` | `{"type": "json_object"}` |
| `Think` | `*bool` | Think mode |
| `Features` | `map[string]bool` | `{web_search, memory, code_interpreter}` |
| `ToolIDs` | `[]string` | Tool IDs |
| `Params` | `map[string]interface{}` | `{function_calling, system, custom_params}` |
| `ChatID` | `string` | Chat session ID |
| `Variables` | `map[string]interface{}` | Template variables |
| `BackgroundTasks` | `map[string]bool` | `{title_generation, follow_up_generation}` |
| `UseCredits` | `*bool` | Use API credits |

### Response object

| Method | Type | Description |
|--------|------|-------------|
| `Content()` | `string` | The AI's answer |
| `TotalTokens()` | `int` | Total tokens used |
| `ID` | `string` | Unique response ID |
| `Model` | `string` | Model used |

---

## FAQ

**Q: What Go version do I need?**
Go 1.21 or higher.

**Q: Are there any dependencies?**
No. The SDK uses only the Go standard library.

**Q: Can I use this in a web server?**
Yes! It works in any Go application — HTTP servers, CLI tools, workers, etc.

---

## License

MIT © MeeDoo