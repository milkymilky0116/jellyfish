package main

import (
	"bufio"
	"log"

	"github.com/milkymilky0116/jellyfish/internal/mails"
)

type Mails struct {
	TotalMails int
	Email      string
	Password   string
	TagSeq     int
	Writer     *bufio.Writer
	Reader     *bufio.Reader
}

func main() {
	server := "imap.gmail.com:993"

	err := mails.InitMailClient(server)
	if err != nil {
		log.Fatal(err)
	}
}
