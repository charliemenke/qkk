package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
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

const usage = `Usage of qkk:
  qkk [-n NAMESPACE] -r RESOURCE [-p PATTERN] ACTION ...
Options:
  -n, --namespace NAMESPACE        Search and take action in k8s namespace NAMESPACE. default: 'default'
  -r, --resource RESOURCE          Search and take action on k8 resource RESOURCE.
  -p, --pattern PATTERN            Search k8 RESOURCE by pattern PATTERN. default: ''
  -h, --help                       prints help information 
`

func (s state) Init() tea.Cmd {
	return nil
}

func (s state) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return s, tea.Quit
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < len(s.choices)-1 {
				s.cursor++
			}
		case "enter", " ":
			s.selected = strings.Fields(s.choices[s.cursor])[0]
			return s, tea.Quit
		}
	}

	return s, nil
}

func (s state) View() string {
	str := "select which resource you want to run your action on:\n\n"
	// str += fmt.Sprintf("cursor: %d\n", s.cursor)
	// str += fmt.Sprintf("mod cursor: %d\n", (s.cursor - (15 * (s.cursor / 15))))
	chunkedChoices := s.choices[15*(s.cursor/15):]

	for i, choice := range chunkedChoices {
		if i == 0 && s.cursor >= 15 {
			str += fmt.Sprintln("  ...")
		}
		cursor := " "
		if (s.cursor - (15 * (s.cursor / 15))) == i {
			cursor = ">"
		}

		// Render the row
		str += fmt.Sprintf("%s %s\n", cursor, choice)

		if i == 14 {
			str += fmt.Sprintln("  ...")
			break
		}
	}

	return str
}

func initState(resource, namespace, pattern string) (state, error) {
	cmdStr := fmt.Sprintf(`kubectl get %s --no-headers -n %s`, resource, namespace)
	if pattern != "" {
		cmdStr += fmt.Sprintf(` | grep -i "%s"`, pattern)
	}

	out, err := exec.Command("bash", "-c", cmdStr).Output()
	if err != nil {
		return state{}, errors.Join(fmt.Errorf("%s", string(out)), err)
	}

	choices := strings.Split(strings.Trim(string(out), "\n"), "\n")
	if len(choices) == 0 {
		fmt.Printf("no %s results found with pattern '%s'\n", resource, pattern)
		os.Exit(0)
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
	// init cli args
	var resource string
	flag.StringVar(&resource, "resource", "", "which kubernetes resource to get logs from")
	flag.StringVar(&resource, "r", "", "which kubernetes resource to get logs from")
	var namespace string
	flag.StringVar(&namespace, "namespace", "default", "which kubernetes namespace to search in")
	flag.StringVar(&namespace, "n", "default", "which kubernetes namespace to search in")
	var pattern string
	flag.StringVar(&pattern, "pattern", "", "grep pattern to search across specified kubernetes resource")
	flag.StringVar(&pattern, "p", "", "grep pattern to search across specified kubernetes resource")

	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	if resource == "" {
		fmt.Println("missing required flag --resource or -r")
		os.Exit(1)
	}

	actions := flag.Args()
	if len(actions) == 0 {
		fmt.Println("missing kubectl action like 'log' or 'edit'")
		os.Exit(1)
	}

	// init cli tool
	model, err := initState(resource, namespace, pattern)
	if err != nil {
		fmt.Printf("failed to start qklog: %v\n", err)
		os.Exit(1)
	}

	qklog := tea.NewProgram(model)
	returnedModel, err := qklog.Run()
	if err != nil {
		fmt.Printf("error running qklog: %v", err)
	}

	fmt.Println("")

	if state, ok := returnedModel.(state); ok && state.selected != "" {
		// fmt.Printf("action: %s\n", strings.Join(actions, " "))
		// fmt.Printf("resource: %s\n", resource)
		// fmt.Printf("selected: %s\n", state.selected)

		// logs action has different pattern
		var cmd *exec.Cmd
		if actions[0] == "logs" {
			cmd = exec.Command("kubectl", append(actions, "-n", namespace, state.selected)...)
		} else {
			cmd = exec.Command("kubectl", append(actions, "-n", namespace, state.resource, state.selected)...)
		}
		fmt.Printf("running: %s\n\n\n", cmd.String())

		stdout, _ := cmd.StdoutPipe()
		stdoerr, _ := cmd.StderrPipe()

		// if action is edit we need to redirect the command std outputs to simulate a terminal
		if actions[0] == "edit" {
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}

		err := cmd.Start()
		if err != nil {
			fmt.Printf("error running kubectl action '%s' on resource '%s': %v", strings.Join(actions, " "), state.resource, err)
			stdout.Close()
			os.Exit(1)
		}
		buf := bufio.NewScanner(io.MultiReader(stdout, stdoerr))
		for buf.Scan() {
			fmt.Println(buf.Text())
		}
		err = cmd.Wait()
		if err != nil {
			fmt.Printf("error running kubectl action '%s' on resource '%s': %v", strings.Join(actions, " "), state.resource, err)
			os.Exit(1)
		}
	}
}
