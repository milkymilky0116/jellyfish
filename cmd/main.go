package main

import (
	"log"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/milkymilky0116/jellyfish/internal/mails"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type Model struct {
	CurrentMails []mails.Email
	Table        table.Model
}

func initModel(client *mails.Mails) *Model {
	columns := []table.Column{
		{Title: "From", Width: 50},
		{Title: "Subject", Width: 50},
		{Title: "Date", Width: 50},
	}
	rows := []table.Row{}
	for _, mail := range client.Emails {
		rows = append(rows, table.Row{mail.From, mail.Subject, mail.Date.String()})
	}
	t := table.New(table.WithColumns(columns), table.WithRows(rows), table.WithHeight(10))
	style := table.DefaultStyles()
	style.Header = style.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	style.Selected = style.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(style)
	return &Model{
		CurrentMails: client.Emails,
		Table:        t,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.Table.Focused() {
				m.Table.Blur()
			} else {
				m.Table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Batch(
				tea.Printf("Let's go to %s!", m.Table.SelectedRow()[1]),
			)
		}
	}
	m.Table, cmd = m.Table.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return baseStyle.Render(m.Table.View()) + "\n"
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

	tui := tea.NewProgram(initModel(client))
	if _, err := tui.Run(); err != nil {
		log.Fatalf("Fail to run tui: %v", err)
	}
}
