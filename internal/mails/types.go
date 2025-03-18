package mails

import (
	"bufio"
	"crypto/tls"
	"time"
)

type MailClient struct {
	TotalMails int
	Email      string
	Password   string
	TagSeq     int
	Writer     *bufio.Writer
	Reader     *bufio.Reader
	Conn       *tls.Conn
	Emails     []Email
	Categories []string
}

type Email struct {
	Subject string
	From    string
	Date    time.Time
}
