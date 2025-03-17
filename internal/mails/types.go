package mails

import (
	"bufio"
	"time"
)

type Mails struct {
	TotalMails int
	Email      string
	Password   string
	TagSeq     int
	Writer     *bufio.Writer
	Reader     *bufio.Reader
}

type Email struct {
	Subject string
	From    string
	Date    time.Time
}
