package tokenmanager

import (
	"fmt"
	"log"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/goccy/go-yaml"
)

type TokenManager struct {
	refreshInterval time.Duration
	file            string
	tokens          map[string]tokenRecord
	mutex           sync.RWMutex
}

type tokenRecord struct {
	Token string `yaml:"token"`
}

func Run(tokenFile string, refreshInterval time.Duration) (*TokenManager, error) {
	tm := &TokenManager{
		file:            tokenFile,
		tokens:          map[string]tokenRecord{},
		refreshInterval: refreshInterval,
	}
	err := tm.reloadTokens()
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			time.Sleep(tm.refreshInterval)
			err := tm.reloadTokens()
			if err != nil {
				log.Printf("failed reloading tokens: %s\n", err.Error())
			}
		}
	}()
	return tm, nil
}

func (tm *TokenManager) HasToken(token string) bool {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	_, ok := tm.tokens[token]
	return ok
}

func (tm *TokenManager) reloadTokens() error {
	data, err := os.ReadFile(tm.file)
	if err != nil {
		return err
	}
	var newTokensRecords []tokenRecord
	err = yaml.Unmarshal(data, &newTokensRecords)
	if err != nil {
		return err
	}

	newTokens := []string{}
	for idx, k := range newTokensRecords {
		if k.Token == "" {
			return fmt.Errorf("token number %d in file %s is empty", idx+1, tm.file)
		}
		newTokens = append(newTokens, k.Token)

	}
	if len(newTokens) < 1 {
		return fmt.Errorf("no tokens found in %s", tm.file)
	}

	tm.mutex.RLock()
	currentTokens := []string{}
	for k, _ := range tm.tokens {
		currentTokens = append(currentTokens, k)
	}
	tm.mutex.RUnlock()

	if len(newTokens) == len(currentTokens) {
		slices.Sort(newTokens)
		slices.Sort(currentTokens)
		if slices.Equal(newTokens, currentTokens) {
			return nil
		}
	}

	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	for k := range tm.tokens {
		delete(tm.tokens, k)
	}

	for _, tokenRecord := range newTokensRecords {
		tm.tokens[tokenRecord.Token] = tokenRecord
	}
	log.Printf("updated tokens from %s\n", tm.file)
	return nil

}
