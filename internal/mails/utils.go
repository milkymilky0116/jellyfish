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

	"github.com/milkymilky0116/jellyfish/internal/repository"
	"golang.org/x/net/html/charset"
)

func findEmailContent(content string) ([]repository.Email, error) {
	contents := []repository.Email{}
	if len(content) > 5 {
		emailDatas := strings.Split(content, "\n* ")
		emailDatas[0] = emailDatas[0][1:]
		for _, block := range emailDatas {
			idIndex := strings.Index(block, "FETCH")
			id, err := strconv.Atoi(strings.TrimSpace(block[0:idIndex]))
			if err != nil {
				return nil, err
			}
			contentLengthStart, end := 0, 0
			for index, letter := range block {
				if letter == '{' {
					contentLengthStart = index
				}
				if letter == '}' {
					end = index
					contentLength, err := strconv.Atoi(block[contentLengthStart+1 : end])
					if err != nil {
						return nil, err
					}
					raw := block[end+1 : end+contentLength]
					emailContents := strings.Split(raw, "\n")
					email := repository.Email{}
					email.Seq = int64(id)
					var subjectStrBuilder strings.Builder
					for _, rawLines := range emailContents {
						rawLines = strings.TrimSpace(rawLines)
						switch {
						case strings.HasPrefix(rawLines, "From:"):
							decodedStr, err := DecodeMimeContent(strings.TrimPrefix(rawLines, "From:"))
							if err != nil {
								return nil, err
							}
							email.Sender = decodedStr
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
							email.EmailDate = parsedTime
						case strings.HasPrefix(rawLines, "FROM:"):
							decodedStr, err := DecodeMimeContent(strings.TrimPrefix(rawLines, "FROM:"))
							if err != nil {
								return nil, err
							}
							email.Sender = decodedStr
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
							email.EmailDate = parsedTime
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
func findModSeq(content []string) (int, error) {
	for _, line := range content {
		if len(line) < 2 || !strings.Contains(line, "FETCH") {
			continue
		}
		index := strings.Index(line, "MODSEQ")
		result, err := strconv.Atoi(line[index+8 : len(line)-3])
		if err != nil {
			return 0, err
		}
		return result, nil
	}
	return 0, nil
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
