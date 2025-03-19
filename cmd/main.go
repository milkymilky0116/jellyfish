package main

import (
	"context"
	"log"
	"os"

	"github.com/milkymilky0116/jellyfish/internal/db"
	"github.com/milkymilky0116/jellyfish/internal/mails"
	"github.com/milkymilky0116/jellyfish/internal/repository"
)

func main() {
	server := os.Getenv("IMAP_URL")
	ctx := context.Background()
	db, err := db.InitSqliteDB(ctx)
	if err != nil {
		log.Fatal(err)
	}
	repo := repository.New(db)
	client, err := mails.InitMailClient(server, repo)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Conn.Close()
	// model, err := tui.InitModel(client)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// tui := tea.NewProgram(model)
	// if _, err := tui.Run(); err != nil {
	// 	log.Fatalf("Fail to run tui: %v", err)
	// }

}
