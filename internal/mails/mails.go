package mails

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/milkymilky0116/jellyfish/internal/db"
	"github.com/milkymilky0116/jellyfish/internal/repository"
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
	for name, category := range mailsClient.Emails {
		// 1. Iterate Each Category, select mailbox
		err = mailsClient.SelectMailBox(name)
		if err != nil {
			return nil, err
		}
		existedCategory, err := mailsClient.CacheRepository.GetCategory(context.TODO(), name)
		if err != nil {
			// 2. if category does not exists on db, create new category, caching emails, save latest modseq
			fmt.Println("Create new category, caching email..")
			decodedName, err := DecodeModifiedUTF7(category.Name)
			if err != nil {
				return nil, err
			}
			createCategoryParam := repository.CreateCategoryParams{
				Name:   decodedName,
				Key:    category.Name,
				Modseq: int64(category.TotalMails),
			}
			newCategory, err := mailsClient.CacheRepository.CreateCategory(context.TODO(), createCategoryParam)
			if err != nil {
				return nil, err
			}
			err = mailsClient.FetchMail(category)
			if err != nil {
				return nil, err
			}
			for _, mail := range category.Mails {
				createEmailParam := repository.CreateEmailParams{
					Seq:       mail.Seq,
					Sender:    mail.Sender,
					Subject:   mail.Subject,
					EmailDate: mail.EmailDate,
				}
				newEmail, err := mailsClient.CacheRepository.CreateEmail(context.TODO(), createEmailParam)
				if err != nil {
					return nil, err
				}
				registerEmailCategoryParam := repository.RegisterEmailAndCategoryParams{
					EmailID:    newEmail.ID,
					CategoryID: newCategory.ID,
				}
				err = mailsClient.CacheRepository.RegisterEmailAndCategory(context.TODO(), registerEmailCategoryParam)
				if err != nil {
					return nil, err
				}
			}
		} else {
			fmt.Println("Category Already Cached")
			// 3. if exists, search latest modseq

			_, err = mailsClient.FindUpdatedEmail(int(existedCategory.Modseq))
			if err != nil {
				return nil, err
			}
			// 4. if modseq is not equal to db's modseq, then search updated emails
			// 5. update db
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

func (m *MailClient) FindModSeq() (int, error) {
	code, err := m.SendMessage("FETCH", fmt.Sprintf("* (MODSEQ)"))

	if err != nil {
		return 0, err
	}
	seq, err := m.ReadModSeqMessage(code)
	if err != nil {
		return 0, err
	}

	return seq, nil
}

func (m *MailClient) FindUpdatedEmail(seq int) ([]int, error) {
	code, err := m.SendMessage("SEARCH", fmt.Sprintf("MODSEQ %d", seq))
	if err != nil {
		return nil, err
	}
	err = m.ReadMessage(code)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (m *MailClient) SelectMailBox(mailbox string) error {
	code, err := m.SendMessage("SELECT", fmt.Sprintf("\"%s\" (CONDSTORE)", mailbox))
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

func (m *MailClient) FetchMail(category *Category) error {
	if category.TotalMails == 0 {
		return nil
	}
	// start := category.TotalMails - ((page - 1) * offset)
	// end := max(start-offset+1, 1)

	code, err := m.SendMessage("FETCH", "1:* (BODY[HEADER.FIELDS (SUBJECT FROM DATE)])")
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
	fmt.Println(imapMsg)
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
	for _, block := range content {
		fmt.Println(block)
	}
	for _, line := range content {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "HIGHESTMODSEQ") {
			index := strings.Index(line, "HIGHESTMODSEQ")
			num := line[index+len("HIGHESTMODSEQ")+1 : len(line)-1]
			totalMails, err := strconv.Atoi(num)
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
		m.Emails[category] = &Category{Name: category, Mails: []repository.Email{}}
	}
	return nil
}

func (m *MailClient) ReadModSeqMessage(code string) (int, error) {
	content, _, err := m.ParseIMAPContent(code)
	if err != nil {
		return 0, err
	}
	latestSeq, err := findModSeq(content)
	if err != nil {
		return 0, err
	}
	return latestSeq, nil
}

func (m *MailClient) ReadMessage(code string) error {
	content, _, err := m.ParseIMAPContent(code)
	if err != nil {
		return err
	}
	for _, block := range content {
		fmt.Println(block)
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
