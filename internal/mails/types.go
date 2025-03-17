package mails

import (
	"bufio"
	"crypto/tls"
	"time"
)

type Mails struct {
	TotalMails int
	Email      string
	Password   string
	TagSeq     int
	Writer     *bufio.Writer
	Reader     *bufio.Reader
	Conn       *tls.Conn
	Emails     []Email
}

type Email struct {
	Subject string
	From    string
	Date    time.Time
}
