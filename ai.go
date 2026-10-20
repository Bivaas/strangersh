package main 


import ( 

	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"

)


type message struct {

	Role string `json:"role"`
	Content string `json:"content"`
}

// ask the AI
func aiReply(p persona, chat []message, chatID int) tea.Cmd {

	return func() tea.Msg  {

		// persona and context as system prompt 
		system := message{Role: "system", Content: "You are " + p.Name + ". " + p.About + " You are texting as a stranger individual to someone. You have to reply like a real person texting with short, causal in completely human form while behaving like your character description. avoid excessive emoji in every message, only use where appropriate. Don't tell anything about being AI. If the user seem to be rude or boring, you can leave saying a short goodbye and end your message with [leave]"}

		reply, err := askAI(system, chat)
		if err != nil { 

			log.Error("AI failed", "error", err)
			return replyMsg{chatID, ""}
		}
		// human delay
		time.Sleep(time.Duration(min(len(reply)*30, 4000)) * time.Millisecond)

		return replyMsg{chatID, reply}
	}
}

// end summary 
func aiSummary(name string, chat []message, chatID int) tea.Cmd {

	return func() tea.Msg {

		system := message{Role: "system", Content: "write a short recap of this conversation between visitor(user) and " + name + "(assistant). Talk with user as third person narrator in 2-4 lines saying what you talked about, what you shared or learned and optionally very short reason for why chat ended. You should write in plain text, no list and NEVER use em dashes!"}

		text, err := askAI(system, chat)

		if err != nil {

			log.Error ("AI summary failed", "error", err)
			return summaryMsg{chatID, "no summaries :("}
		}
		return summaryMsg{chatID, text}
	}
}


// model details and config
func askAI(system message, chat []message) (string, error) {

	body, _ := json.Marshal(map[string]any{

		"model": aiModel,
		"messages": append([]message{system}, chat...),
		"max_tokens": 500,
		"chat_template_kwargs": map[string]any{"enable_thinking": false},
	})

	req, _ := http.NewRequest("POST", "https://integrate.api.nvidia.com/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")


	// for msg
	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)

	if err != nil {
		
		return "", err
	}
	defer resp.Body.Close()

	var result struct { 

		Choices []struct {
			Message message `json:"message"`
		} `json:"choices"`
	}

	json.NewDecoder(resp.Body).Decode(&result)
	if len(result.Choices) == 0 {

		return "", errors.New("Nvidia has said " + resp.Status)
	}
	return result.Choices[0].Message.Content, nil
	
}