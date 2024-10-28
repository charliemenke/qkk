package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type state struct {
	pattern  string   // pattern to search k8s resources with
	resource string   // what k8s resource we are targeting
	choices  []string // resulting choices based on pattern search
	selected string   // selected choice
	cursor   int      // current item selected in choices
}

func (s state) Init() tea.Cmd {
	return nil
}

func (s state) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return s, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if s.cursor < len(s.choices)-1 {
				s.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			s.selected = strings.Fields(s.choices[s.cursor])[0]
			fmt.Printf("getting logs...\n\n")
			return s, tea.Quit
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return s, nil
}

func (s state) View() string {
	str := "select which resource you want to see logs for:\n\n"
	for i, choice := range s.choices {
		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if s.cursor == i {
			cursor = ">" // cursor!
		}

		// Render the row
		str += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	return str
}

func initState(resource, pattern string) (state, error) {
	out, err := exec.Command("bash", "-c", fmt.Sprintf(`kubectl get %s --no-headers | grep -i "%s"`, resource, pattern)).Output()
	if err != nil {
		return state{}, err
	}

	return state{
		pattern:  pattern,
		resource: resource,
		choices:  strings.Split(strings.Trim(string(out), "\n"), "\n"),
		selected: "",
		cursor:   0,
	}, nil
}

func main() {
	fmt.Println("starting qklog")

	// init cli args
	resourceFlg := flag.String("resource", "", "which kubernetes resource to get logs from")
	patternFlg := flag.String("pattern", "", "grep pattern to search across specified kubernetes resource")
	flag.Parse()

	if *resourceFlg == "" {
		fmt.Println("missing required flag --resource")
		os.Exit(1)

	}
	if *patternFlg == "" {
		fmt.Println("missing required flag --pattern")
		os.Exit(1)

	}

	// init cli tool
	model, err := initState(*resourceFlg, *patternFlg)
	if err != nil {
		fmt.Printf("failed to start qklog: %v", err)
		os.Exit(1)
	}

	qklog := tea.NewProgram(model)
	returnedModel, err := qklog.Run()
	if err != nil {
		fmt.Printf("error running qklog: %v", err)
	}

	if state, ok := returnedModel.(state); ok && state.selected != "" {
		cmd := exec.Command("kubectl", "logs", state.selected)
		stdout, _ := cmd.StdoutPipe()
		err := cmd.Start()
		if err != nil {
			fmt.Printf("error getting logs from %s: %v", state.selected, err)
			stdout.Close()
			os.Exit(1)
		}
		buf := bufio.NewScanner(stdout)
		for buf.Scan() {
			fmt.Println(buf.Text())
		}
		err = cmd.Wait()
		if err != nil {
			fmt.Printf("error getting logs from %s: %v", state.selected, err)
			os.Exit(1)
		}
	}
}
