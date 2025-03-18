package utils

import (
	"encoding/json"

	"github.com/milkymilky0116/jellyfish/internal/mails"
)

func ParseValue(value []byte) (*map[string][]mails.Email, error) {
	var content map[string][]mails.Email
	err := json.Unmarshal(value, &content)
	if err != nil {
		return nil, err
	}
	return &content, nil
}
