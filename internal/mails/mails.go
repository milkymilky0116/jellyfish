package mails

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/milkymilky0116/jellyfish/internal/db"
)

func InitMailClient(url string, repo db.IRepository) (*MailClient, error) {
	conn, err := tls.Dial("tcp", url, nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("✅ Connect to IMAP Server", conn.RemoteAddr())
	mailsClient := InitMails(conn, repo)
	err = mailsClient.Login()
	if err != nil {
		return nil, err
	}
	err = mailsClient.ListMailBox()
	if err != nil {
		return nil, err
	}
	for key, value := range mailsClient.Emails {
		err := mailsClient.SelectMailBox(key)
		if err != nil {
			return nil, err
		}

		if data, err := mailsClient.CacheRepository.Get(key); err != nil {
			log.Println("Not caching..")
			err = mailsClient.FetchMail(value, 1, 10)
			if err != nil {
				return nil, err
			}
			bytes, err := json.Marshal(*value)
			if err != nil {
				return nil, err
			}
			err = mailsClient.CacheRepository.Set(key, bytes)
			if err != nil {
				return nil, err
			}
		} else {
			log.Println("Cached..")
			var category Category
			err = json.Unmarshal(data, &category)
			if err != nil {
				return nil, err
			}
			mailsClient.Emails[key] = &category
		}
	}

	return mailsClient, nil
}

func InitMails(conn *tls.Conn, repo db.IRepository) *MailClient {
	return &MailClient{
		Writer:          bufio.NewWriter(conn),
		Reader:          bufio.NewReader(conn),
		Conn:            conn,
		ClienEmail:      os.Getenv("IMAP_EMAIL"),
		ClientPassword:  os.Getenv("IMAP_PASSWORD"),
		CacheRepository: repo,
		Emails:          make(map[string]*Category),
	}
}

func (m *MailClient) NextTag() string {
	m.TagSeq++
	return fmt.Sprintf("a%03d", m.TagSeq)
}

func (m *MailClient) Login() error {
	code, err := m.SendMessage("LOGIN", fmt.Sprintf("\"%s\" \"%s\"", m.ClienEmail, m.ClientPassword))
	if err != nil {
		log.Printf("fail to send message: %v", err)
	}
	err = m.ReadMessage(code)
	if err != nil {
		return errors.New("fail to login")
	}
	return nil
}

func (m *MailClient) ListMailBox() error {
	code, err := m.SendMessage("LIST", "\"\" \"*\"")
	if err != nil {
		return err
	}
	err = m.ReadListMessage(code)
	if err != nil {
		return err
	}
	return nil
}

func (m *MailClient) SelectMailBox(mailbox string) error {
	code, err := m.SendMessage("SELECT", fmt.Sprintf("\"%s\"", mailbox))
	if err != nil {
		return err
	}
	err = m.ReadSelectMessage(mailbox, code)
	if err != nil {
		return err
	}
	m.CurrentMailBox = mailbox
	return nil
}

func (m *MailClient) FetchMail(category *Category, page, offset int) error {
	if category.TotalMails == 0 {
		return nil
	}
	start := category.TotalMails - ((page - 1) * offset)
	end := max(start-offset+1, 1)

	code, err := m.SendMessage("FETCH", fmt.Sprintf("%d:%d (BODY[HEADER.FIELDS (SUBJECT FROM DATE)])", end, start))
	if err != nil {
		return err
	}
	err = m.ReadFetchMessage(category.Name, code)
	if err != nil {
		return err
	}
	return nil
}

func (m *MailClient) SendMessage(msgType, msg string) (string, error) {
	code := m.NextTag()
	imapMsg := fmt.Sprintf("%s %s %s\r\n", code, msgType, msg)
	log.Println(imapMsg)
	if _, err := m.Writer.WriteString(imapMsg); err != nil {
		return "", err
	}
	return code, m.Writer.Flush()
}

func (m *MailClient) ReadSelectMessage(inbox, code string) error {
	content, _, err := m.ParseIMAPContent(code)
	if err != nil {
		return err
	}
	for _, line := range content {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "EXISTS") {
			totalMails, err := strconv.Atoi(line[2 : len(line)-7])
			if err != nil {
				return err
			}
			if entry, ok := m.Emails[inbox]; ok {
				entry.TotalMails = totalMails
				m.Emails[inbox] = entry
			}
		}
	}
	return nil
}

func (m *MailClient) ReadFetchMessage(inbox, code string) error {
	content, _, err := m.ParseIMAPContent(code)
	if err != nil {
		return err
	}
	contents := strings.Join(content, "\n")
	emailContents, err := findEmailContent(contents)
	if err != nil {
		return err
	}
	fmt.Println(emailContents)
	if entry, ok := m.Emails[inbox]; ok {
		entry.Mails = emailContents
		m.Emails[inbox] = entry
	}
	return nil
}

func (m *MailClient) ReadListMessage(code string) error {
	content, _, err := m.ParseIMAPContent(code)
	if err != nil {
		return err
	}
	contents := strings.Join(content, "\n")
	categories, err := findEmailBox(contents)
	if err != nil {
		return err
	}
	for _, category := range categories {
		m.Emails[category] = &Category{Name: category, Mails: []Email{}}
	}
	return nil
}

func (m *MailClient) ReadMessage(code string) error {
	_, _, err := m.ParseIMAPContent(code)
	if err != nil {
		return err
	}
	return nil
}

func (m *MailClient) ParseIMAPContent(code string) ([]string, string, error) {
	var buffer strings.Builder
	for {
		line, err := m.Reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, "", err
		}

		buffer.WriteString(line)
		if strings.HasPrefix(line, fmt.Sprintf("%s OK", code)) {
			break
		} else if strings.HasPrefix(line, fmt.Sprintf("%s BAD", code)) || strings.HasPrefix(line, fmt.Sprintf("%s NO", code)) {
			return nil, "", errors.New(fmt.Sprintf("message return bad response: %v", buffer.String()))
		}
	}
	content, result := splitContent(code, buffer.String())
	return content, result, nil
}
