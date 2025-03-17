package mails

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

func InitMailClient(url string) error {
	conn, err := tls.Dial("tcp", url, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	fmt.Println("✅ Connect to IMAP Server", conn.RemoteAddr())

	mailsClient := InitMails(conn)
	err = mailsClient.Login()
	if err != nil {
		return err
	}

	code, err := mailsClient.SendMessage("SELECT", "INBOX")
	if err != nil {
		return err
	}
	mailsClient.ReadMessage(code, "SELECT")
	code, err = mailsClient.SendMessage("FETCH", fmt.Sprintf("%d:%d (BODY[HEADER.FIELDS (SUBJECT FROM DATE)])", mailsClient.TotalMails-9, mailsClient.TotalMails))
	if err != nil {
		return err
	}
	err = mailsClient.ReadMessage(code, "FETCH")
	if err != nil {
		return err
	}
	return nil
}

func InitMails(conn *tls.Conn) *Mails {
	return &Mails{
		Writer:   bufio.NewWriter(conn),
		Reader:   bufio.NewReader(conn),
		Email:    os.Getenv("SMTP_EMAIL"),
		Password: os.Getenv("SMTP_PASSWORD"),
	}
}

func (m *Mails) NextTag() string {
	m.TagSeq++
	return fmt.Sprintf("a%03d", m.TagSeq)
}

func (m *Mails) Login() error {
	code, err := m.SendMessage("LOGIN", fmt.Sprintf("\"%s\" \"%s\"", m.Email, m.Password))
	if err != nil {
		log.Printf("fail to send message: %v", err)
	}
	err = m.ReadMessage(code, "LOGIN")
	if err != nil {
		return errors.New("fail to login")
	}
	return nil
}

func (m *Mails) SendMessage(msgType, msg string) (string, error) {
	code := m.NextTag()
	imapMsg := fmt.Sprintf("%s %s %s\r\n", code, msgType, msg)
	if _, err := m.Writer.WriteString(imapMsg); err != nil {
		return "", err
	}
	return code, m.Writer.Flush()
}

func (m *Mails) ReadMessage(code, msgType string) error {
	var buffer strings.Builder
	for {
		line, err := m.Reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		buffer.WriteString(line)
		if strings.HasPrefix(line, fmt.Sprintf("%s OK", code)) {
			break
		} else if strings.HasPrefix(line, fmt.Sprintf("%s BAD", code)) || strings.HasPrefix(line, fmt.Sprintf("%s NO", code)) {
			return errors.New("message return bad response")
		}
	}
	content, _ := splitContent(code, buffer.String())

	switch msgType {
	case "SELECT":
		for _, line := range content {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "EXISTS") {
				totalMails, err := strconv.Atoi(line[2 : len(line)-7])
				if err != nil {
					return err
				}
				m.TotalMails = totalMails
			}
		}
	case "FETCH":
		contents := strings.Join(content, "\n")
		emailContents, err := findEmailContent(contents)
		if err != nil {
			return err
		}
		fmt.Println(emailContents)
	}
	return nil
}
