package main 


import (

	tea "github.com/charmbracelet/bubbletea"
	"os"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
)



func main() { 

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

	return model{}, nil
}

type model struct{}

func (m model) Init() tea.Cmd {
	return nil
}


// to run in new events like keypresses
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "q" {
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {

	return "Welcome....... \n\n Press q to leave \n"
}