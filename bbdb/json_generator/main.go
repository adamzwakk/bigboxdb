package main

import (
	"fmt"
	"os"
	"strings"
	"strconv"
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"

	"github.com/adamzwakk/bigboxdb/server/db"
	"github.com/adamzwakk/bigboxdb/server/models"
	"github.com/adamzwakk/bigboxdb/tools"
)

type stepKind int

const (
	textInput stepKind = iota
	singleSelect
	multiSelect
)

type step struct {
	kind        stepKind
	title       string
	options     []string // for select types
}

type model struct {
	steps       []step
	currentStep int
	combined     tools.ImportData

	// Text input state
	textValue  string
	cursorPos  int

	// Single select state
	cursor int

	// Multi select state
	selected map[int]bool

	// Completion
	done bool
}

func initialModel() model {
	d := db.GetDB()

	var boxtypes []models.BoxType
	d.Find(&boxtypes)
	boxTypeNames := make([]string, len(boxtypes))
	for i, p := range boxtypes {
		boxTypeNames[i] = p.Name 
	}

	var regions []models.Region
	d.Find(&regions)
	regionNames := make([]string, len(regions))
	for i, p := range regions {
		regionNames[i] = p.Name 
	}

	var platforms []models.Platform
	d.Find(&platforms)
	platformNames := make([]string, len(platforms))
	for i, p := range platforms {
		platformNames[i] = p.Name
	}

	return model{
		steps: []step{
			{
				kind:        textInput,
				title:       "What is the game title?",
			},
			{
				kind:        textInput,
				title:       "What is the game region?",
			},
			{
				kind:        textInput,
				title:       "When was the game released?",
			},
			{
				kind:        textInput,
				title:       "What kind of variant is it?",
			},
			{
				kind:        textInput,
				title:       "Should it sort by a different name?",
			},
			{
				kind:        textInput,
				title:       "Who is the Developer?",
			},
			{
				kind:        textInput,
				title:       "Who is the Publisher?",
			},
			{
				kind:        textInput,
				title:       "What is the platform",
			},
			{
				kind:        singleSelect,
				title:       "What is the box type?",
				options:     boxTypeNames,
			},
			{
				kind:        textInput,
				title:       "Box width?",
			},
			{
				kind:        textInput,
				title:       "Box height?",
			},
			{
				kind:        textInput,
				title:       "Box depth?",
			},
			{
				kind:        textInput,
				title:       "Does the gatefold need a transparent flag?",
			},
			{
				kind:        textInput,
				title:       "Any notes about the scan?",
			},
			{
				kind:        textInput,
				title:       "Please enter the IGDB",
			},
			{
				kind:        textInput,
				title:       "Please enter the MobyGamesID",
			},
		},
		selected: make(map[int]bool),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		if msg, ok := msg.(tea.KeyMsg); ok {
			if msg.String() == "q" || msg.String() == "ctrl+c" || msg.String() == "enter" {
				return m, tea.Quit
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			if m.currentStep > 0 {
				m.currentStep--
				m.resetStepState()
			}
			return m, nil
		}

		currentStep := m.steps[m.currentStep]

		switch currentStep.kind {
		case textInput:
			return m.updateTextInput(msg)
		case singleSelect:
			return m.updateSingleSelect(msg)
		case multiSelect:
			return m.updateMultiSelect(msg)
		}
	}

	return m, nil
}

func (m model) updateTextInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		if len(strings.TrimSpace(m.textValue)) > 0 {
			m.saveCurrentStep()
			m.advanceStep()
		}
	case tea.KeyBackspace:
		if m.cursorPos > 0 {
			m.textValue = m.textValue[:m.cursorPos-1] + m.textValue[m.cursorPos:]
			m.cursorPos--
		}
	case tea.KeyLeft:
		if m.cursorPos > 0 {
			m.cursorPos--
		}
	case tea.KeyRight:
		if m.cursorPos < len(m.textValue) {
			m.cursorPos++
		}
	case tea.KeyRunes:
		ch := string(msg.Runes)
		m.textValue = m.textValue[:m.cursorPos] + ch + m.textValue[m.cursorPos:]
		m.cursorPos += len(ch)
	}
	return m, nil
}

func (m model) updateSingleSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	step := m.steps[m.currentStep]
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(step.options)-1 {
			m.cursor++
		}
	case "enter":
		m.saveCurrentStep()
		m.advanceStep()
	}
	return m, nil
}

func (m model) updateMultiSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	step := m.steps[m.currentStep]
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(step.options)-1 {
			m.cursor++
		}
	case " ":
		m.selected[m.cursor] = !m.selected[m.cursor]
	case "enter":
		m.saveCurrentStep()
		m.advanceStep()
	}
	return m, nil
}

func strPtr(s string) *string {
    return &s
}

func boolPtr(s string) *bool {
    s = strings.ToLower(strings.TrimSpace(s))
    v := s == "true" || s == "yes" || s == "y" || s == "1"
    return &v
}

func intPtr(s string) *int {
    val, err := strconv.Atoi(strings.TrimSpace(s))
    if err != nil {
        return nil
    }
    return &val
}

func (m *model) saveCurrentStep() {
	step := m.steps[m.currentStep]
	d := db.GetDB()

	switch m.currentStep {
		case 0:
			m.combined.Title = strings.TrimSpace(m.textValue)
		case 1:
			m.combined.Region = strPtr(strings.TrimSpace(m.textValue))
		case 2:
			m.combined.Year, _ = strconv.Atoi(strings.TrimSpace(m.textValue))
		case 3:
			m.combined.Variant = strings.TrimSpace(m.textValue)
		case 4:
			m.combined.SeriesSort = strings.TrimSpace(m.textValue)
		case 5:
			m.combined.Developer = tools.FirstString(strings.TrimSpace(m.textValue))
		case 6:
			m.combined.Publisher = tools.FirstString(strings.TrimSpace(m.textValue))
		case 7:
			m.combined.Platform = strings.TrimSpace(m.textValue)
		case 8:
			var bt models.BoxType
			d.Where(models.BoxType{Name: step.options[m.cursor]}).First(&bt)

			m.combined.BoxType = bt.ID
		case 9:
			val, err := strconv.ParseFloat(strings.TrimSpace(m.textValue),32)
			if err == nil {
				m.combined.Width = float32(val)
			}
		case 10:
			val, err := strconv.ParseFloat(strings.TrimSpace(m.textValue),32)
			if err == nil {
				m.combined.Height = float32(val)
			}
		case 11:
			val, err := strconv.ParseFloat(strings.TrimSpace(m.textValue),32)
			if err == nil {
				m.combined.Depth = float32(val)
			}
		case 12:
			m.combined.GatefoldTransparent = boolPtr(strings.TrimSpace(m.textValue))
		case 13:
			m.combined.ScanNotes = strings.TrimSpace(m.textValue)
		case 14:
			m.combined.IGDBId = intPtr(m.textValue)
		case 15:
			m.combined.MobygamesId = intPtr(m.textValue)
	}
}

func (m *model) advanceStep() {
	if m.currentStep < len(m.steps)-1 {
		m.currentStep++
		m.resetStepState()
	} else {
		m.done = true
	}
}

func (m *model) resetStepState() {
	m.textValue = ""
	m.cursorPos = 0
	m.cursor = 0
	m.selected = make(map[int]bool)
}

func (m model) View() string {
	if m.done {
		return m.viewResult()
	}

	step := m.steps[m.currentStep]
	progress := fmt.Sprintf("Step %d of %d", m.currentStep+1, len(m.steps))

	var s strings.Builder
	s.WriteString(progress + "\n\n")
	s.WriteString(step.title + "\n\n")

	switch step.kind {
	case textInput:
		s.WriteString(m.viewTextInput())
	case singleSelect:
		s.WriteString(m.viewSingleSelect())
	case multiSelect:
		s.WriteString(m.viewMultiSelect())
	}

	help := "\nenter: confirm"
	if m.currentStep > 0 {
		help += " | esc: back"
	}
	help += " | ctrl+c: quit"
	s.WriteString(help + "\n")

	return s.String()
}

func (m model) viewTextInput() string {
	before := m.textValue[:m.cursorPos]
	after := m.textValue[m.cursorPos:]
	return "> " + before + "█" + after + "\n"
}

func (m model) viewSingleSelect() string {
	step := m.steps[m.currentStep]
	var s strings.Builder
	for i, opt := range step.options {
		if i == m.cursor {
			s.WriteString("❯ " + opt + "\n")
		} else {
			s.WriteString("  " + opt + "\n")
		}
	}
	return s.String()
}

func (m model) viewMultiSelect() string {
	step := m.steps[m.currentStep]
	var s strings.Builder
	for i, opt := range step.options {
		cursor := "  "
		if i == m.cursor {
			cursor = "❯ "
		}
		checked := "○"
		if m.selected[i] {
			checked = "●"
		}
		s.WriteString(fmt.Sprintf("%s%s %s\n", cursor, checked, opt))
	}
	return s.String()
}

func (m model) viewResult() string {
	var s strings.Builder
	s.WriteString("Done! Press enter or q to quit.\n")
	return s.String()
}

func main(){
	godotenv.Load("./../.env")
	fmt.Println("BigBoxDB JSON Generator")

	p := tea.NewProgram(initialModel())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	m := finalModel.(model)
	if m.done {
		b, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Print(string(b))
	}
}