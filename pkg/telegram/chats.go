package telegram

import (
	"encoding/json"
	"fmt"

	"github.com/docker/libkv/store"
	tb "gopkg.in/tucnak/telebot.v2"
)

const telegramChatsDirectory = "telegram/chats"

// ChatStore writes the users to a libkv store backend
type ChatStore struct {
	kv store.Store
}

// NewChatStore stores telegram chats in the provided kv backend
func NewChatStore(kv store.Store) (*ChatStore, error) {
	return &ChatStore{kv: kv}, nil
}

// List all chats saved in the kv backend
func (s *ChatStore) List() ([]tb.Chat, error) {
	kvPairs, err := s.kv.List(telegramChatsDirectory)
	if err != nil {
		return nil, err
	}

	var chats []tb.Chat
	for _, kv := range kvPairs {
		var c tb.Chat
		if err := json.Unmarshal(kv.Value, &c); err != nil {
			return nil, err
		}
		chats = append(chats, c)
	}

	return chats, nil
}

// Add a telegram chat to the kv backend
func (s *ChatStore) Add(c tb.Chat) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s/%d", telegramChatsDirectory, c.ID)

	return s.kv.Put(key, b, nil)
}

// Remove a telegram chat from the kv backend
func (s *ChatStore) Remove(c tb.Chat) error {
	key := fmt.Sprintf("%s/%d", telegramChatsDirectory, c.ID)
	return s.kv.Delete(key)
}
