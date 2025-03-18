package tui

import "github.com/milkymilky0116/jellyfish/internal/mails"

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
