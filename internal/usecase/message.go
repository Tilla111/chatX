package service

import (
	"chatX/internal/store"
	"context"
)

type Message struct {
	ID          int64  `json:"id"`
	ChatID      int64  `json:"chat_id"`
	SenderID    int64  `json:"sender_id"`
	MessageText string `json:"message_text"`
	IsRead      bool   `json:"is_read"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	SenderName  string `json:"sender_name"`
	ChatName    string `json:"chat_name"`
}

type MessageSRV struct {
	repo *store.Storage
}

func (s *MessageSRV) Create(ctx context.Context, msg Message) (*Message, error) {
	req := store.Message{
		ChatID:      msg.ChatID,
		SenderID:    msg.SenderID,
		MessageText: msg.MessageText,
	}

	message, err := s.repo.MessageStorage.Create(ctx, &req)
	if err != nil {
		return nil, err
	}

	m := Message{
		ID:          message.ID,
		ChatID:      message.ChatID,
		SenderID:    message.SenderID,
		MessageText: message.MessageText,
		IsRead:      message.IsRead,
		CreatedAt:   message.CreatedAt,
		UpdatedAt:   message.UpdatedAt,
		SenderName:  message.SenderName,
		ChatName:    message.ChatName,
	}

	return &m, nil
}

type MessageDetail struct {
	ID         int64  `json:"id"`
	Content    string `json:"content"`
	SenderID   int64  `json:"sender_id"`
	SenderName string `json:"sender_name"`
	CreatedAt  string `json:"created_at"`
	IsRead     bool   `json:"is_read"`
}

func (s *MessageSRV) GetByID(ctx context.Context, id int64) (*Message, error) {
	message, err := s.repo.MessageStorage.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:          message.ID,
		ChatID:      message.ChatID,
		SenderID:    message.SenderID,
		MessageText: message.MessageText,
		IsRead:      message.IsRead,
		CreatedAt:   message.CreatedAt,
		UpdatedAt:   message.UpdatedAt,
	}, nil
}

func (s *MessageSRV) GetByChatID(ctx context.Context, chatID int64) ([]MessageDetail, error) {

	var messags []MessageDetail

	messag, err := s.repo.MessageStorage.GetMessages(ctx, chatID)
	if err != nil {
		return nil, err
	}

	for _, msg := range messag {
		mes := MessageDetail{
			ID:         msg.ID,
			Content:    msg.Content,
			SenderID:   msg.SenderID,
			SenderName: msg.SenderName,
			CreatedAt:  msg.CreatedAt,
			IsRead:     msg.IsRead,
		}

		messags = append(messags, mes)
	}

	return messags, nil
}

func (s *MessageSRV) MarkChatAsRead(ctx context.Context, chatID, userID int64) error {
	return s.repo.MessageStorage.MarkAsRead(ctx, chatID, userID)
}

func (s *MessageSRV) UpdateMessage(ctx context.Context, msgID, userID int64, newText string) error {
	return s.repo.MessageStorage.Update(ctx, msgID, userID, newText)
}

func (s *MessageSRV) DeleteMessage(ctx context.Context, msgID, userID int64) error {
	return s.repo.MessageStorage.Delete(ctx, msgID, userID)
}
