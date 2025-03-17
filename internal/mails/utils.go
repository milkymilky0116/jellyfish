package mails

import (
	"io"
	"mime"
	"net/mail"
	"strconv"
	"strings"

	"golang.org/x/net/html/charset"
)

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
	decoder.CharsetReader = func(encoding string, input io.Reader) (io.Reader, error) {
		return charset.NewReader(input, encoding)
	}
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
