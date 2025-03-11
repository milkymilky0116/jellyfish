package main

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/mail"
	"os"
	"strconv"
	"strings"
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

func main() {
	server := "imap.gmail.com:993"

	conn, err := tls.Dial("tcp", server, nil)
	if err != nil {
		log.Fatal("TLS Dial Error:", err)
	}
	defer conn.Close()

	fmt.Println("✅ Connect to IMAP Server", conn.RemoteAddr())

	mails := InitMails(conn)
	err = mails.Login()
	if err != nil {
		log.Fatalf("fail to login: %v", err)
	}

	code, err := mails.SendMessage("SELECT", "INBOX")
	if err != nil {
		log.Printf("fail to send message: %v", err)
	}
	mails.ReadMessage(code, "SELECT")
	code, err = mails.SendMessage("FETCH", fmt.Sprintf("%d:%d (BODY[HEADER.FIELDS (SUBJECT FROM DATE)])", mails.TotalMails-9, mails.TotalMails))
	if err != nil {
		log.Printf("fail to send message: %v", err)
	}
	ok, err := mails.ReadMessage(code, "FETCH")
	if !ok && err != nil {
		log.Printf("fail to read message: %v", err)
	}
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
	ok, err := m.ReadMessage(code, "LOGIN")
	if err != nil || !ok {
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

func (m *Mails) ReadMessage(code, msgType string) (bool, error) {
	var buffer strings.Builder
	for {
		line, err := m.Reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return false, err
		}

		buffer.WriteString(line)
		if strings.HasPrefix(line, fmt.Sprintf("%s OK", code)) {
			break
		} else if strings.HasPrefix(line, fmt.Sprintf("%s BAD", code)) || strings.HasPrefix(line, fmt.Sprintf("%s NO", code)) {
			return false, errors.New("message return bad response")
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
					return false, err
				}
				m.TotalMails = totalMails
			}
		}
	case "FETCH":
		contents := strings.Join(content, "\n")
		emailContents, err := findEmailContent(contents)
		if err != nil {
			return false, err
		}
		fmt.Println(emailContents)
	}
	return true, nil
}

type Email struct {
	Subject string
	From    string
	Date    time.Time
}

func findEmailContent(content string) ([]Email, error) {
	contents := []Email{}
	for index, letter := range content {
		start, end := -1, -1
		if letter == '*' {
			start = index
			for j := start; j < len(content); j++ {
				if content[j] == '}' {
					end = j
					contentLength, err := strconv.Atoi(content[end-3 : end])
					if err != nil {
						return nil, err
					}
					raw := content[end+1 : end+contentLength]
					emailContents := strings.Split(raw, "\n")
					email := Email{}
					var subjectStrBuilder strings.Builder
					for _, rawLines := range emailContents {
						rawLines = strings.TrimSpace(rawLines)
						switch {
						case strings.HasPrefix(rawLines, "From:"):
							decodedStr, err := DecodeMimeContent(strings.TrimPrefix(rawLines, "From:"))
							if err != nil {
								return nil, err
							}
							email.From = decodedStr
						case strings.HasPrefix(rawLines, "Subject:"):
							decodedStr, err := DecodeMimeContent(strings.TrimPrefix(rawLines, "Subject:"))
							if err != nil {
								return nil, err
							}
							subjectStrBuilder.WriteString(decodedStr)
						case strings.HasPrefix(rawLines, "Date:"):
							parsedTime, err := mail.ParseDate(strings.TrimPrefix(rawLines, "Date: "))
							if err != nil {
								return nil, err
							}
							email.Date = parsedTime
						default:
							decodedStr, err := DecodeMimeContent(rawLines)
							if err != nil {
								return nil, err
							}
							subjectStrBuilder.WriteString(decodedStr)
						}
					}
					email.Subject = subjectStrBuilder.String()
					contents = append(contents, email)
					break
				}
			}
		}
	}
	return contents, nil
}

func DecodeMimeContent(str string) (string, error) {
	decoder := mime.WordDecoder{}
	decodedStr, err := decoder.DecodeHeader(str)
	if err != nil {
		return "", err
	}
	return decodedStr, nil
}
func splitContent(code, content string) ([]string, string) {
	contentLines := []string{}
	resultLines := ""
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, code) {
			resultLines = line
		} else {
			contentLines = append(contentLines, line)
		}
	}
	return contentLines, resultLines
}
