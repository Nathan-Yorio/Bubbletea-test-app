// Something is wrong with the way I've merged the logic of choice and cursor
// that's why my view selection is not updating, I'll have to figure out what's up with that
package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	// term "golang.org/x/term"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/indent"
	"github.com/spf13/viper" //for config file
)

// 📜📜📜📜📜📜~~WHAT CONTENT IS UPDATING~~📜📜📜📜📜📜

// Functions that return messages to other functions
type (
	frameMsg struct{}
)

func frame() tea.Cmd {
	return tea.Tick(time.Second/60, func(time.Time) tea.Msg {
		return frameMsg{}
	})
}

type model struct {
	programsPath   string
	programs       []string // list of programs in the directory
	programOptions []string //things you can do with the program you chose
	selected       map[int]struct{}
	Choice         int //stores value of cursor position for first selection
	secondChoice   int //stores value of cursor position for second selection
	Chosen         bool
	secondChosen   bool
	Frames         int
	Quitting       bool
	optionOne      string
	optionTwo      string
	renderFlag     bool
}

func pathSelectModel() model {
	return model{
		programs: wut_files(configRead()), //hand the list of programs to the model from the viper read function
		programOptions: []string{
			"Run Program",
			"Return",
			"Exit",
		},
		selected:     make(map[int]struct{}), //mathematical set mapping for choice selection
		Choice:       0,
		secondChoice: 0,
		Chosen:       false,
		secondChosen: false,
		Frames:       0,
		Quitting:     false,
		optionOne:    "",
		optionTwo:    "",
		renderFlag:   false, //used to wait and render checkmark before moving on
	}
}

// 🏁🏁🏁🏁🏁🏁~~~~~Initialize Commands~~~~~🏁🏁🏁🏁🏁🏁
func (m model) Init() tea.Cmd {
	return nil //don't need to have an initially running function
}

// 🤔🤔🤔🤔🤔🤔~~~~~~~~Wut IS MY LOGIC~~~~~🤔🤔🤔🤔🤔🤔

// Main update function.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Make sure these keys always quit
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			m.Quitting = true
			return m, tea.Quit
		}
	}

	// Hand off the message and model to the appropriate update function for the
	// appropriate view based on the current state.
	if !m.Chosen {
		return updateProgChoice(msg, m)
	}
	return updateOptionChoice(msg, m)
}

// View Update 1 ~~~ Choosing a Program
func updateProgChoice(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			m.Choice++
			if m.Choice > len(m.programs)-1 { //don't allow to exceed array bounds
				m.Choice = len(m.programs) - 1
			}
			if m.Choice > (m.Frames+1)*10-1 && m.Frames < len(m.programs)/10 {
				m.Frames++
			}
		case "k", "up":
			m.Choice--
			if m.Choice < 0 { //don't allow to exceed array bounds
				m.Choice = 0
			}
			if m.Choice < m.Frames*10 && m.Frames > 0 {
				m.Frames--
			}
		case "enter":
			// Store that we've chosen the first option, save the choice
			m.Chosen = true
			m.optionOne = m.programs[m.Choice]
			return m, frame()
		}
	}

	return m, nil
}

// View Update 2 ~~~ What to do with the program
func updateOptionChoice(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			m.secondChoice++
			if m.secondChoice > len(m.programs)-1 { //don't allow cursor to exceed bounds
				m.secondChoice = len(m.programs) - 1
			}
		case "k", "up":
			m.secondChoice--
			if m.secondChoice < 0 { // don't allow cursor to exceed bounds
				m.secondChoice = 0
			}
		case "enter":
			m.secondChosen = true
			m.optionTwo = m.programOptions[m.secondChoice] //store the user's second choice
			switch m.optionTwo {
			case "Run Program":

				//run the program
				cmd := exec.Command(m.optionOne)
				cmd.Dir = configRead()

				// Set the output to os.Stdout and os.Stderr
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Stdin = os.Stdin

				err := cmd.Run()
				if err != nil {
					fmt.Println("Error executing the command:", err)
					os.Exit(1)
				}

			case "Return":
				//Send me back to menu one
				m.Chosen = false //nothing has been chosen
				m.secondChosen = false
				m.optionOne = "" //reset first choice
				m.optionTwo = "" //reset return choice

			case "Exit":
				//quit the program
				m.Quitting = true
				return m, tea.Quit
			}
			return m, frame()
		}
	}

	return m, nil
}

// 👀👀👀👀👀~~~~~~~Wut MY USER SEES~~~~~~👀👀👀👀👀👀👀
// The main view, which just calls the appropriate sub-view
func (m model) View() string {
	var s string
	if m.Quitting {
		return "\n  さようなら!\n\n"
	}
	if !m.Chosen { //have we made our first choice? && !m.renderFlag, was a test
		s = listPrograms(m)

	} else if !m.secondChosen { //have we made our second choice?
		//time.Sleep(300 * time.Millisecond) //debug wait line
		s = programQuestions(m)
	} else {
		s = chosenProgram(m) //now we're executing the program... or something
	}

	return indent.String("\n"+s+"\n\n", 5)
}

// Subview 1 ~~~ List Programs
func listPrograms(m model) string {
	// The header
	//s := "Which file will you select?\n\n"

	// The header
	//s := dynamicStyles(m).Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#333333")).Render("Which file will you select?\n\n")

	header := headerStyle.
		Render("Which file will you select?")

	// Number of programs to display per "terminal" page
	programsPerPage := 10

	// Calculate the starting and ending index of programs to display
	startIndex := m.Frames * programsPerPage
	endIndex := (m.Frames + 1) * programsPerPage

	// If the ending index is greater than the total number of programs, set it to the total number of programs
	if endIndex > len(m.programs) {
		endIndex = len(m.programs)
	}

	// Iterate over the programs in the current page
	programList := ""
	for i := startIndex; i < endIndex; i++ {
		// Is the cursor pointing at this choice?
		//cursor := " " // no cursor
		cursor := unselectedStyle
		if m.Choice == i {
			cursor = selectedStyle
		}

		//s += fmt.Sprintf("%s ゝ %s\n", cursor, m.programs[i]) //render the choice selected

		// Render the program name with the appropriate styles
		//program := "ゝ" + cursor + m.programs[i]
		program := cursor.Render("🔥 ゝ " + m.programs[i])
		// Append the program to the programList string
		programList += listStyle.Render(program)
	}

	// The footer
	// s += "\nPress q, esc, or ctrl-c to quit.\n"

	// The footer
	footer := footerStyle.
		Render("Press j or down arrow to scroll down, k or up arrow to scroll up.\n" + "Press enter to select a program.\n" + "Press q, esc, or ctrl-c to quit.\n")

	if m.Quitting {
		farewell := "\n  さようなら!\n\n"
		return farewell
	} else {
		return header + programList + footer
	}
}

func programQuestions(m model) string {
	// The header
	s := "Which file will you select?\n\n"
	s += "your first choice was:"
	s += fmt.Sprintf("%s\n", m.optionOne)

	// Iterate over our choices
	for i, choice := range m.programOptions {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.secondChoice == i {
			cursor = "🔥" // cursor!
		}

		s += fmt.Sprintf("%s ゝ %s\n", cursor, choice) //render the choice selected
	}

	// The footer
	s += "\nPress q, esc, or ctrl-c to quit.\n"

	if m.Quitting {
		s += "\n  さようなら!\n\n"
	}

	// Send the UI for rendering

	return s
}

// Subview 2 ~~~ Chosen Program
func chosenProgram(m model) string {
	s := "This is the selected option view\n\n"
	s += "Ideally there will be some status or something here\n\n"
	s += "You've chosen to: "
	s += fmt.Sprintf("%s\n", m.optionTwo)
	if m.Quitting {
		s += "\n  さようなら!\n\n"
	}
	return s
}

// 📑 what files are in a given directory
func wut_files(dirPath string) (the_files []string) {
	files, _ := os.ReadDir(dirPath)
	for _, file := range files {
		the_files = append(the_files, file.Name()) // make a list of the files in the chosen directory
	}
	return the_files
}

// 🐍🐍~~~~~~~~~~~~~~~YML CONFIG~~~~~~~~~~~~~~~~~~~~🐍🐍
// Config file name, where, type, error handling
func configRead() string {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
		tea.Quit()
	}

	viper.SetDefault("PROGRAMS_PATH", ".") //Default value for path if unset
	return viper.GetString("PROGRAMS_PATH")
}

// 🔥🔥🔥🔥🔥~~~~~~Make the Magic Happen~~~~~~🔥🔥🔥🔥🔥
func main() {
	//clear the terminal window before starting MVC
	clear := exec.Command("clear")
	clear.Stdout = os.Stdout
	clear.Run()

	p := tea.NewProgram(pathSelectModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
	}
}

// 💄💄💄💄💄💄💄💄~~~~~~~STYLE~~~~~~💄💄💄💄💄💄💄💄💄
// Lip Gloss styles
var (
	// Define some common styles
	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("200")).
			Background(lipgloss.Color("252")).
			Width(70).
			Margin(0, 0, 0, 0).
			Padding(0, 0, 0, 0)

	selectedStyle = lipgloss.NewStyle().
			Width(70).
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#ff00bf"))
	unselectedStyle   = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("232"))

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E0E2E4")).
			Background(lipgloss.Color("#333333")).
			Width(70).
			Height(0).
			PaddingBottom(0).
			Align(lipgloss.Center)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E0E2E4")).
			Background(lipgloss.Color("#333333")).
			Width(70).
			Padding(0, 0, 0, 0).
			Margin(1, 0, 0, 0).
			Align(lipgloss.Center)

	listStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("232")).
			Width(70).
			Margin(1, 0, 0, 0).
			Align(lipgloss.Left)

	centerStyle = lipgloss.NewStyle().
			Align(lipgloss.Center)
)
