package main 


import (

	tea "github.com/charmbracelet/bubbletea"
	
	"os"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	
	"github.com/charmbracelet/bubbles/textinput"
)

const aiModel = "nvidia/nemotron-3.5-lightning-30b-a3b"
var apiKey string 

func main() { 

	// API import and setup
	apiKey = os.Getenv("NV_API_KEY")
	if apiKey == "" {

		log.Error("apiKey not setup correctly !")
		os.Exit(1)
	}
	log.Info("key loaded", "model", aiModel)


	srv, err := wish.NewServer(

		// config and server key 
		wish.WithAddress(":2222"),
		
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(bubbletea.Middleware(teaHandler)),
	)

	if err != nil {

		log.Error("server cannot be started !", "error", err)
		os.Exit(1)
	
	}

	log.Info("Server started !")

	if err := srv.ListenAndServe(); err != nil {

		log.Error("Server has stopped !!", "error", err)
		os.Exit(1)
	} 
}

// new function for each unique user
func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {

	input := textinput.New()
	input.Placeholder = "say hi to me :) "
	input.Width = 50
	input.Focus()

	return model{input: input, who: randomPersona()}, nil
}


func (m model) Init() tea.Cmd {
	return nil
}


// to run in new events like keypresses
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// esc quit
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
		return m, tea.Quit
	}

	// enter send and fake msg
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" && m.input.Value() != "" {

		m.messages = append(m.messages, message{Role: "user", Content: m.input.Value()})
		m.input.Reset()
		m.typing = true
		return m, aiReply(m.who, m.messages )
	}

	// got reply now with name and context
	if reply, ok := msg.(replyMsg); ok {

		m.messages = append(m.messages, message{Role: "assistant", Content: string(reply)})
		m.typing = false
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd

}

// user box and response with name
func (m model) View() string {

	chat := ""
	for _, line := range m.messages {

		name := "you"
		if line.Role == "assistant" {
			name = m.who.Name
		}
		chat += name + ": " + line.Content + "\n"

		// typing indicator
		if m.typing { 
			
			chat += m.who.Name + " is typing....\n"
		}
	}

	return "welcome to my new world ..... \n\n you are chatting with " + m.who.Name + "\n\n" + chat + "\n" + m.input.View() + "\n\n press esc to leave \n" 
}


// user input and message storage array
type  model struct {

	input textinput.Model
	messages []message
	typing bool
	who persona
}

// bot ans back to update
type replyMsg string
