package mails

import (
	"bufio"
	"crypto/tls"
	"database/sql"

	"github.com/milkymilky0116/jellyfish/internal/db"
	"github.com/milkymilky0116/jellyfish/internal/repository"
)

type MailClient struct {
	ClienEmail      string
	ClientPassword  string
	TagSeq          int
	Writer          *bufio.Writer
	Reader          *bufio.Reader
	Conn            *tls.Conn
	DB              *sql.DB
	CurrentMailBox  string
	Emails          map[string]*Category
	CacheRepository db.IRepository
}

type Category struct {
	TotalMails int
	Name       string
	Mails      []repository.Email
}
