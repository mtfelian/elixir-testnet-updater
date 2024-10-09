package notifier

import (
	"fmt"
	"log"
	"os"
	"sync/atomic"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TGBot is a Telegram Bot message sender
type TGBot struct {
	chatIDs     map[int64]bool // Stores chat IDs in memory
	forceChatID atomic.Int64
	bot         *tgbotapi.BotAPI
	instanceID  string
}

// TGBotParams represents TG bot client parameters
type TGBotParams struct {
	BotToken    string
	ForceChatID int64
	InstanceID  string
}

// NewTGBot creates a new TGBot instance
func NewTGBot(p TGBotParams) (*TGBot, error) {
	bot, err := tgbotapi.NewBotAPI(p.BotToken)
	if err != nil {
		return nil, err
	}

	botClient := &TGBot{
		chatIDs:    make(map[int64]bool),
		bot:        bot,
		instanceID: p.InstanceID,
	}
	botClient.forceChatID.Store(p.ForceChatID)

	if err := botClient.loadChatIDs(); err != nil {
		log.Printf("Error loading chat IDs: %v", err)
	}

	if botClient.forceChatID.Load() == 0 {
		go botClient.listenAndStoreChatIDs()
	}

	return botClient, nil
}

// saveChatID saves a chat ID to the file and memory
func (bot *TGBot) saveChatID(chatID int64) error {
	if bot.chatIDs[chatID] { // already exists
		return nil
	}
	bot.chatIDs[chatID] = true

	file, err := os.OpenFile("chat_ids.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%d\n", chatID))
	return err
}

// loadChatIDs loads stored chat IDs from the file
func (bot *TGBot) loadChatIDs() error {
	if forceChatID := bot.forceChatID.Load(); forceChatID != 0 {
		bot.chatIDs[forceChatID] = true
		return nil
	}

	file, err := os.Open("chat_ids.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	var chatID int64
	for {
		if _, err := fmt.Fscanf(file, "%d\n", &chatID); err != nil {
			break
		}
		bot.chatIDs[chatID] = true

		// !!! it's enough really to have 1 subscribed account
		bot.forceChatID.Store(chatID)
		return nil

	}

	return nil
}

// ListenAndStoreChatIDs listens for new messages and stores chat IDs dynamically
func (bot *TGBot) listenAndStoreChatIDs() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		chatID := update.Message.Chat.ID

		if err := bot.saveChatID(chatID); err != nil {
			log.Printf("Error saving chat ID: %v", err)
		}

		msg := tgbotapi.NewMessage(chatID, "Chat ID stored, you will receive broadcasts.")
		if err := bot.send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}

		// !!! it's enough really to have 1 subscribed account
		bot.forceChatID.Store(chatID)
		bot.bot.StopReceivingUpdates()
		return
	}
}

// send the message according to it's configuration
func (bot *TGBot) send(msg tgbotapi.MessageConfig) error {
	msg.Text = fmt.Sprintf("[%s] %s", bot.instanceID, msg.Text)
	_, err := bot.bot.Send(msg)
	return err
}

// SendBroadcastMessage sends a message to all stored chat IDs
func (bot *TGBot) SendBroadcastMessage(message string) {
	for chatID := range bot.chatIDs {
		msg := tgbotapi.NewMessage(chatID, message)
		if err := bot.send(msg); err != nil {
			log.Printf("Error sending message to chat %d: %v", chatID, err)
			continue
		}
		log.Printf("Message sent to chat %d", chatID)
	}
}
