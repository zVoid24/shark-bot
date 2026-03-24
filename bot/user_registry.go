package bot

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const knownUsersFilePath = "data/known_users.txt"

var knownUsersMu sync.Mutex

func (b *Bot) trackKnownUser(u *tgbotapi.User) {
	if u == nil {
		return
	}

	userID := fmt.Sprintf("%d", u.ID)
	if userID == "" {
		return
	}

	fullName := strings.TrimSpace(strings.TrimSpace(u.FirstName) + " " + strings.TrimSpace(u.LastName))
	if fullName == "" {
		fullName = "Unknown"
	}

	if _, err := b.userSvc.EnsureUser(userID, fullName); err != nil {
		log.Warn("failed to ensure user while tracking", "user_id", userID, "err", err)
	}

	if err := addKnownUserID(userID); err != nil {
		log.Warn("failed to persist known user", "user_id", userID, "err", err)
	}
}

func mergeUniqueUserIDs(primary, secondary []string) []string {
	seen := make(map[string]struct{}, len(primary)+len(secondary))
	out := make([]string, 0, len(primary)+len(secondary))

	appendUnique := func(ids []string) {
		for _, id := range ids {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}

	appendUnique(primary)
	appendUnique(secondary)
	return out
}

func getKnownUserIDs() ([]string, error) {
	knownUsersMu.Lock()
	defer knownUsersMu.Unlock()

	file, err := os.Open(knownUsersFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer file.Close()

	seen := map[string]struct{}{}
	out := []string{}

	s := bufio.NewScanner(file)
	for s.Scan() {
		id := strings.TrimSpace(s.Text())
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func addKnownUserID(userID string) error {
	knownUsersMu.Lock()
	defer knownUsersMu.Unlock()

	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil
	}

	existing := map[string]struct{}{}
	if file, err := os.Open(knownUsersFilePath); err == nil {
		s := bufio.NewScanner(file)
		for s.Scan() {
			id := strings.TrimSpace(s.Text())
			if id != "" {
				existing[id] = struct{}{}
			}
		}
		_ = file.Close()
	}

	if _, ok := existing[userID]; ok {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(knownUsersFilePath), 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(knownUsersFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(userID + "\n")
	return err
}
