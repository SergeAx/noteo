package bot

import (
	"log/slog"

	"github.com/google/uuid"
	"github.com/tucnak/telebot"
)

var (
	btnBackToMenu = telebot.ReplyButton{Text: "Back to main menu"}

	mainMenu = &telebot.ReplyMarkup{
		ReplyKeyboard: [][]telebot.ReplyButton{
			{btnMyProjects, btnMySubscriptions},
		},
		ResizeReplyKeyboard: true,
	}
)

type mainMenuHandler struct {
	service *Service
}

func newMainMenuHandler(s *Service) *mainMenuHandler {
	return &mainMenuHandler{service: s}
}

func (h *mainMenuHandler) register() {
	h.service.bot.Handle("/start", h.handleStart)
	h.service.bot.Handle(&btnBackToMenu, h.handleBackToMenu)
	h.service.bot.Handle(telebot.OnText, h.handleTextMessage)
}

func (h *mainMenuHandler) handleStart(m *telebot.Message) {
	slog.Info("Received /start command", "user_id", m.Sender.ID, "payload", m.Payload)
	h.service.stateManager.ClearState(m.Sender.ID)

	if m.Payload == "" {
		h.service.bot.Send(m.Sender, "Welcome to Noteo! Choose an option:", mainMenu)
		return
	}

	projectID, err := uuid.Parse(m.Payload)
	if err != nil {
		slog.Error("Invalid project ID in subscription link", "error", err, "payload", m.Payload)
		h.service.bot.Send(m.Sender, "Sorry, this subscription link is invalid.", mainMenu)
		return
	}

	err = h.service.subscriptions.handleSubscriptionLink(m, projectID)
	if err != nil {
		slog.Error("Failed to handle subscription link", "error", err)
		h.service.bot.Send(m.Sender, "Sorry, failed to process your subscription. Please try again later.", mainMenu)
	}
}

func (h *mainMenuHandler) handleBackToMenu(m *telebot.Message) {
	h.service.stateManager.ClearState(m.Sender.ID)
	h.service.bot.Send(m.Sender, "Main menu:", mainMenu)
}

func (h *mainMenuHandler) handleTextMessage(m *telebot.Message) {
	state, exists := h.service.stateManager.GetState(m.Sender.ID)
	if !exists {
		slog.Debug("Received unhandled text message",
			"text", m.Text,
			"user_id", m.Sender.ID)
		h.service.bot.Send(m.Sender, "Please use the menu buttons.", mainMenu)
		return
	}

	switch state.State {
	case StateCreatingProject:
		if err := h.service.projects.handleProjectCreation(m); err != nil {
			slog.Error("Failed to create project", "error", err)
			h.service.bot.Send(m.Sender, "Sorry, failed to create project. Please try again.", projectsMenu)
		}
		h.service.stateManager.ClearState(m.Sender.ID)

	default:
		slog.Error("Unknown state", "state", state.State)
		h.service.stateManager.ClearState(m.Sender.ID)
		h.service.bot.Send(m.Sender, "Something went wrong. Please try again.", mainMenu)
	}
}
