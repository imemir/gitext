package ai

import (
	"strings"
)

// trimMessage cleans up the generated commit message
func trimMessage(msg string) string {
	msg = strings.TrimSpace(msg)
	// Remove surrounding quotes if present
	if len(msg) >= 2 && msg[0] == '"' && msg[len(msg)-1] == '"' {
		msg = msg[1 : len(msg)-1]
	}
	if len(msg) >= 2 && msg[0] == '\'' && msg[len(msg)-1] == '\'' {
		msg = msg[1 : len(msg)-1]
	}
	// Remove any trailing newlines or spaces
	msg = strings.TrimSpace(msg)
	return msg
}
