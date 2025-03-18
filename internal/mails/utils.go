package mails

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/mail"
	"strconv"
	"strings"
	"unicode/utf16"

	"golang.org/x/net/html/charset"
)

func findEmailContent(content string) ([]Email, error) {
	contents := []Email{}
	for index, letter := range content {
		start, end := -1, -1
		if letter != '*' {
			continue
		}
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
					case strings.HasPrefix(rawLines, "FROM:"):
						decodedStr, err := DecodeMimeContent(strings.TrimPrefix(rawLines, "FROM:"))
						if err != nil {
							return nil, err
						}
						email.From = decodedStr
					case strings.HasPrefix(rawLines, "SUBJECT:"):
						decodedStr, err := DecodeMimeContent(strings.TrimPrefix(rawLines, "SUBJECT:"))
						if err != nil {
							return nil, err
						}
						subjectStrBuilder.WriteString(decodedStr)
					case strings.HasPrefix(rawLines, "DATE:"):
						parsedTime, err := mail.ParseDate(strings.TrimPrefix(rawLines, "DATE: "))
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
	return contents, nil
}

func findEmailBox(content string) ([]string, error) {
	categories := []string{}
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		start, end := -1, -1
		for index, letter := range line {
			if letter == '(' {
				start = index
				for j := start; j < len(line); j++ {
					if line[j] == ')' {
						end = j
						break
					}
				}
				if strings.Contains(line[start:end], "\\Noselect") {
					continue
				} else {
					start, end = -1, len(line)-2
					for k := end; k >= 0; k-- {
						if line[k] == '"' {
							start = k
							break
						}
					}
					key := line[start+1 : end+1]
					categories = append(categories, key)
				}
			}
		}
	}
	return categories, nil
}

func DecodeModifiedUTF7(s string) (string, error) {
	var result strings.Builder
	i := 0

	for i < len(s) {
		if s[i] == '&' {
			ampIndex := i
			i++

			if i < len(s) && s[i] == '-' {
				result.WriteByte('&')
				i++
				continue
			}

			dashIndex := strings.IndexByte(s[i:], '-')
			if dashIndex == -1 {
				return "", fmt.Errorf("Wrong Modified UTF-7 Encoding")
			}
			dashIndex += i

			base64Str := s[ampIndex+1 : dashIndex]

			base64Str = strings.ReplaceAll(base64Str, ",", "/")

			paddingNeeded := len(base64Str) % 4
			if paddingNeeded > 0 {
				base64Str += strings.Repeat("=", 4-paddingNeeded)
			}

			decoded, err := base64.StdEncoding.DecodeString(base64Str)
			if err != nil {
				return "", fmt.Errorf("Base64 decode error: %v", err)
			}

			utf16Bytes := make([]uint16, len(decoded)/2)
			for j := 0; j < len(decoded); j += 2 {
				if j+1 < len(decoded) {
					utf16Bytes[j/2] = uint16(decoded[j])<<8 | uint16(decoded[j+1])
				}
			}

			r := utf16.Decode(utf16Bytes)
			for _, char := range r {
				result.WriteRune(char)
			}

			i = dashIndex + 1
		} else {
			result.WriteByte(s[i])
			i++
		}
	}

	return result.String(), nil
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
