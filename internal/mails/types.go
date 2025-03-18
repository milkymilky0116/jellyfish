package mails

import (
	"bufio"
	"crypto/tls"
	"time"

	"github.com/milkymilky0116/jellyfish/internal/db"
)

type MailClient struct {
	ClienEmail      string
	ClientPassword  string
	TagSeq          int
	Writer          *bufio.Writer
	Reader          *bufio.Reader
	Conn            *tls.Conn
	CurrentMailBox  string
	Emails          map[string]*Category
	CacheRepository db.IRepository
}

type Category struct {
	TotalMails int
	Name       string
	Mails      []Email
}

type Email struct {
	Subject string
	From    string
	Date    time.Time
}
