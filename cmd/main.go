package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/milkymilky0116/jellyfish/internal/mails"
)

var (
	panelStyle         = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#FAFAFA"))
	selectedPanelStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#7D56F4"))
	listStyle          = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA"))
	selectedListStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#04B575"))
)

type Panel struct {
	id             int
	title          string
	list           []string
	currentElement int
	width          int
	height         int
}

type Model struct {
	Panels       []Panel
	CurrentPanel int
	CurrentList  []string
	Client       *mails.MailClient
}

func initModel(client *mails.MailClient) (*Model, error) {
	categoryList := []string{}
	for _, category := range client.Categories {
		decodedCategory, err := mails.DecodeModifiedUTF7(category)
		if err != nil {
			return nil, err
		}
		categoryList = append(categoryList, decodedCategory)
	}
	categoryPanel := Panel{
		id:    0,
		title: "Category",
		list:  categoryList,
	}
	mailList := []string{}
	for _, mail := range client.Emails {
		mailList = append(mailList, mail.Subject)
	}
	emailPanel := Panel{
		id:    1,
		title: "Email",
		list:  mailList,
	}
	return &Model{
		Panels: []Panel{categoryPanel, emailPanel},
		Client: client,
	}, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Panels[0].width = msg.Width/4 - 2
		m.Panels[0].height = msg.Height - 2
		m.Panels[1].width = (msg.Width / 4 * 3) - 2
		m.Panels[1].height = msg.Height - 2
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.CurrentPanel = (m.CurrentPanel + 1) % len(m.Panels)
			m.CurrentList = m.Panels[m.CurrentPanel].list
		case "enter":
			panel := &m.Panels[m.CurrentPanel]
			switch panel.title {
			case "Category":
				panel := &m.Panels[m.CurrentPanel]
				err := m.Client.SelectMailBox(m.Client.Categories[panel.currentElement])
				if err != nil {
					log.Println(err)
				}
				m.Client.FetchMail(1, 10)
				mailList := []string{}
				for _, mail := range m.Client.Emails {
					mailList = append(mailList, mail.Subject)
				}
				m.Panels[1].list = mailList
			}
		case "j":
			panel := &m.Panels[m.CurrentPanel]
			panel.currentElement = (panel.currentElement + 1) % len(panel.list)
		case "k":
			panel := &m.Panels[m.CurrentPanel]
			panel.currentElement = (len(panel.list) + panel.currentElement - 1) % len(panel.list)
		}
	}
	return m, cmd
}

func (m Model) View() string {
	panels := []string{}
	for _, panel := range m.Panels {
		if m.CurrentPanel == panel.id {
			panels = append(panels, renderSelectedPanel(panel))
		} else {
			panels = append(panels, renderPanel(panel))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, panels...)
}

func renderSelectedPanel(panel Panel) string {
	content := fmt.Sprintf("%s\n", panel.title)
	for index, element := range panel.list {
		if panel.currentElement == index {
			content += fmt.Sprintf(">> %s\n", selectedListStyle.Render(element))
		} else {
			content += fmt.Sprintf("%s\n", listStyle.Render(element))
		}
	}
	return selectedPanelStyle.Width(panel.width).Height(panel.height).Render(lipgloss.JoinVertical(lipgloss.Center, content))
}

func renderPanel(panel Panel) string {
	content := fmt.Sprintf("%s\n%s", panel.title, listStyle.Render(strings.Join(panel.list, "\n")))
	return panelStyle.Width(panel.width).Height(panel.height).Render(lipgloss.JoinVertical(lipgloss.Center, content))
}

func main() {
	server := os.Getenv("IMAP_URL")

	client, err := mails.InitMailClient(server)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Conn.Close()
	err = client.FetchMail(1, 10)
	if err != nil {
		log.Fatal(err)
	}
	model, err := initModel(client)
	if err != nil {
		log.Fatal(err)
	}
	tui := tea.NewProgram(model)
	if _, err := tui.Run(); err != nil {
		log.Fatalf("Fail to run tui: %v", err)
	}
}
