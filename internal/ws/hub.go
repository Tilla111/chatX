package ws

import (
	"encoding/json"
	"sync"
	"time"
)

type Hub struct {
	mu         sync.RWMutex
	Clients    map[string]*Client
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.ID] = client
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.ID]; ok {
				delete(h.Clients, client.ID)
				close(client.Send)
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) BroadcastChatMessage(chatID int64, chatName, senderID, senderName, content string, recipientIDs []string) {
	payload := map[string]interface{}{
		"type":        "new_message",
		"chat_id":     chatID,
		"chat_name":   chatName,
		"sender_id":   senderID,
		"sender_name": senderName,
		"content":     content,
		"created_at":  time.Now().Format("2006-01-02 15:04:05"),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, id := range recipientIDs {

		if id == senderID {
			continue
		}

		if client, ok := h.Clients[id]; ok {
			select {
			case client.Send <- data:
			default:

			}
		}
	}
}

func (h *Hub) BroadcastReadStatus(chatID int64, readerID string, recipientID string) {
	payload := map[string]interface{}{
		"type":      "messages_read",
		"chat_id":   chatID,
		"reader_id": readerID,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if client, ok := h.Clients[recipientID]; ok {
		select {
		case client.Send <- data:
		default:
		}
	}
}

func (h *Hub) broadcastToRecipients(recipients []string, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, id := range recipients {
		if client, ok := h.Clients[id]; ok {
			select {
			case client.Send <- data:
			default:

			}
		}
	}
}

// BroadcastMessageUpdate - xabar tahrirlanganini tarqatadi
func (h *Hub) BroadcastMessageUpdate(chatID, msgID int64, newText string, recipients []string) {
	payload := map[string]interface{}{
		"type":         "message_updated",
		"chat_id":      chatID,
		"message_id":   msgID,
		"message_text": newText,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	h.broadcastToRecipients(recipients, data)
}

// BroadcastMessageDelete - xabar o'chirilganini tarqatadi
func (h *Hub) BroadcastMessageDelete(chatID, msgID int64, recipients []string) {
	payload := map[string]interface{}{
		"type":       "message_deleted",
		"chat_id":    chatID,
		"message_id": msgID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	h.broadcastToRecipients(recipients, data)
}

// BroadcastMemberAdded - groupga yangi a'zo qo'shilganini tarqatadi
func (h *Hub) BroadcastMemberAdded(chatID, userID, addedByID int64, username, addedByName string, recipients []string) {
	payload := map[string]interface{}{
		"type":          "member_added",
		"chat_id":       chatID,
		"user_id":       userID,
		"username":      username,
		"added_by_id":   addedByID,
		"added_by_name": addedByName,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	h.broadcastToRecipients(recipients, data)
}

// BroadcastChatDelete - chat o'chirilganini tarqatadi
func (h *Hub) BroadcastChatDelete(chatID, deletedByID int64, deletedByName string, recipients []string) {
	payload := map[string]interface{}{
		"type":            "chat_deleted",
		"chat_id":         chatID,
		"deleted_by_id":   deletedByID,
		"deleted_by_name": deletedByName,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	h.broadcastToRecipients(recipients, data)
}
