package main 


import (

	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"os"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"time"
	"fmt"
)

const aiModel = "nvidia/nemotron-3-super-120b-a12b"
var apiKey string 

func main() { 

	lipgloss.SetColorProfile(termenv.ANSI256)

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
	input.Placeholder = "Your nickname ?"
	input.Width = 50
	input.Focus()

	return model{input: input, who: randomPersona(), started: time.Now()}, nil
}


func (m model) Init() tea.Cmd {
	return nil
}


// to run in new events like keypresses
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = size.Width
		m.height = size.Height
	}

	// esc quit and bye message
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {

		m.bye = true
		return m, tea.Quit
	}

	// new stranger
	if key, ok :=  msg.(tea.KeyMsg); ok && key.String() == "ctrl+n" {

		old := m.who.Name
		for m.who.Name == old {

			m.who = randomPersona()
		}

		m.messages = nil
		m.typing = false
		m.notice = ""
		m.left = false
		m.input.Reset()

		m.chatID++
		m.started = time.Now()
		m.searching = true
		
		return m, tea.Tick(1500*time.Millisecond, func(time.Time) tea.Msg { return searchDoneMsg{} })
	}
	
	// gets nickname
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" && m.nick == "" && m.input.Value() != "" {
		
		m.nick = m.input.Value()
		m.input.Reset()
		m.input.Placeholder = "say hi to me :) "
		m.started = time.Now()
		return m, nil
	}
	
	// enter send and fake msg
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" && m.input.Value() != "" && !m.typing && !m.left {

		m.messages = append(m.messages, message{Role: "user", Content: m.input.Value()})
		m.input.Reset()
		m.typing = true
		return m, aiReply (m.who, m.messages, m.chatID)
	}

	// search screen
	if _, ok := msg.(searchDoneMsg); ok {

		m.searching = false
		return m, nil
	}

	// got reply now with name and context
	if reply, ok := msg.(replyMsg); ok {

		// old reply skip
		if reply.chatID != m.chatID {

			return m, nil
		}

		m.typing = false
		if reply.text == "" {

			m.notice = "connection failed, send again !"
			return m, nil
		}

		m.notice = ""
		if strings.Contains(reply.text, "[leave]") {

			m.left = true
		}

		text := strings.ReplaceAll(reply.text, "[leave]", "")
		m.messages = append(m.messages, message{Role: "assistant", Content: text})
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd

}

var (
	youStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	strangerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	boxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("205")).Padding(0, 2)

)

// user box and response with name
func (m model) View() string {

	// message summary length and how much talked
	summary := "you talked with " + m.who.Name + " for " + time.Since(m.started).Round(time.Second).String() + " with " + fmt.Sprint(len(m.messages)) + " messages"
	if m.bye {
		return boxStyle.Render("goodbye! " +summary) + "\n"
	}


	if m.searching { 

		return boxStyle.Render("searching for a stranger...")
	}

	if m.nick == "" {
		return boxStyle.Render("* welcome to my new world * \n\nbefore you go in.. what should we call you ?") + "\n\n" + m.input.View()
	}

	// starting and end greetings
	header := boxStyle.Render("* welcome to my new world * \n\n Meet "+m.who.Name+"and greet him !") + "\n\n"
	footer := "\n" + m.input.View() + "\n\n" + dimStyle.Render("ctrl+n: New Chat \nesc: leave")

	chat := ""
	for _, line := range m.messages {

		name := youStyle.Render(m.nick)
		if line.Role == "assistant" {

			name = strangerStyle.Foreground(lipgloss.Color(m.who.Color)).Render(m.who.Name)
		}

		text:= lipgloss.NewStyle().PaddingLeft(2).Width(m.width).Render(line.Content)
		chat += name + "\n" + text + "\n\n"
	}


	// typing indicator
	if m.typing {

		chat += dimStyle.Render(m.who.Name+" is typing....") + "\n"
	}
	if m.left {

		chat += dimStyle.Render(m.who.Name + " has left the chat \n" +summary + "\n" )
	}
	chat += dimStyle.Render(m.notice)


	// lines to fit window
	lines := strings.Split(chat, "\n")
	room := max(m.height-lipgloss.Height(header)-lipgloss.Height(footer), 1)

	if len(lines) > room { 
		lines = lines[len(lines)-room:]
	}
	chat = strings.Join(lines, "\n")

	return header + chat + footer
}


// user input and message storage array
type model struct {

	input textinput.Model
	messages []message
	typing bool
	who persona
	notice string

	width int
	height int
	chatID int 
	left bool
	searching bool
	started time.Time
	bye bool
	nick string
}

type searchDoneMsg struct{}

// bot ans back to update
type replyMsg struct {

	chatID int 
	text string
}
