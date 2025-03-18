package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/milkymilky0116/jellyfish/internal/db"
	"github.com/milkymilky0116/jellyfish/internal/mails"
	"github.com/milkymilky0116/jellyfish/internal/tui"
)

func main() {
	server := os.Getenv("IMAP_URL")
	badgerDB, err := db.InitBadgerDB()
	if err != nil {
		log.Fatal(err)
	}
	repo := db.InitBadgerRepository(badgerDB)
	client, err := mails.InitMailClient(server, repo)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Conn.Close()
	model, err := tui.InitModel(client)
	if err != nil {
		log.Fatal(err)
	}
	tui := tea.NewProgram(model)
	if _, err := tui.Run(); err != nil {
		log.Fatalf("Fail to run tui: %v", err)
	}

}
