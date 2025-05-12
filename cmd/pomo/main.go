package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	args := os.Args
	var minutes int

	// let the user pass in a number of minutes for the session. If they don't, just use a default of 25minutes
	if len(args) > 1 {
		mins, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("Argument must be an integer representing the number of minutes to run the pomodoro")
			os.Exit(1)
		}
		minutes = mins
	} else {
		minutes = 25
	}

	starttime = time.Now()

	m := model{
		sessionLength: time.Duration(minutes) * time.Minute,
		progress:      progress.New(progress.WithDefaultGradient()),
	}

	secs := strconv.Itoa(int(m.sessionLength.Seconds()))
	fmt.Println("SessionLength (sec): " + secs)

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Printf("Unable to run the timer: %s\n", err)
		os.Exit(1)
	}
}

var (
	starttime time.Time

	containerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("86")).
			Render

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render
)

const (
	padding  = 2
	maxWidth = 80

	tickTimeSeconds = 2 * time.Second
)

type state int

const (
	inProgress state = iota
	complete
)

type model struct {
	state           state
	percentComplete float64
	elapsed         time.Duration
	progress        progress.Model
	sessionLength   time.Duration
	containerHeight int
	containerWidth  int
}

type tick time.Time

func (m model) tickCmd() tea.Cmd {
	return tea.Tick(tickTimeSeconds, func(t time.Time) tea.Msg {
		return tick(t)
	})
}

// Init is the first function that will be called. It returns an optional
// initial command. To not perform an initial command return nil.
func (m model) Init() tea.Cmd {
	return m.tickCmd()
}

// Update is called when a message is received. Use it to inspect messages
// and, in response, update the model and/or send a command.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.containerHeight = msg.Height
		m.containerWidth = msg.Width
		m.progress.Width = min(msg.Width-padding*2-4, maxWidth)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c", "q":
			return m, tea.Quit
		}
		return m, nil
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	case tick:
		currentPercentComplete := m.percentComplete
		if currentPercentComplete >= 1.0 {
			m.state = complete
			return m, tea.Quit
		}

		m.elapsed = time.Time(msg).Sub(starttime)
		m.percentComplete = m.elapsed.Seconds() / m.sessionLength.Seconds()
		cmd := m.progress.IncrPercent(float64(m.percentComplete - currentPercentComplete))
		return m, tea.Batch(m.tickCmd(), cmd)
	default:
		return m, nil
	}
}

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (m model) View() string {
	pad := strings.Repeat(" ", padding)

	var msg string
	var display string
	if m.state == complete {
		msg = pad + "Your session is complete!" + pad
		display = pad + "" + pad
	} else {
		msg = pad + fmt.Sprintf("%dmin session in progress...", int(m.sessionLength.Minutes())) + pad
		display = pad + m.progress.View() + pad
	}

	content := lipgloss.JoinVertical(lipgloss.Left, msg, display)
	help := helpStyle(pad + "Press q or ctrl-c to quit" + pad)

	view := containerStyle("\n" +
		content + "\n\n" +
		help + "\n")

	return lipgloss.Place(
		m.containerWidth,
		m.containerHeight,
		lipgloss.Center,
		lipgloss.Center,
		view,
	)
}
