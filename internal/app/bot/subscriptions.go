package bot

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/tucnak/telebot"

	"gitlab.com/trum/noteo/internal/domain"
)

// Menu items
var (
	btnMySubscriptions = telebot.ReplyButton{Text: "My Subscriptions"}

	subscriptionsMenu = &telebot.ReplyMarkup{
		ReplyKeyboard: [][]telebot.ReplyButton{
			{btnBackToMenu},
		},
		ResizeReplyKeyboard: true,
	}
)

type subscriptionsHandler struct {
	service *Service
}

func newSubscriptionsHandler(s *Service) *subscriptionsHandler {
	return &subscriptionsHandler{service: s}
}

func (h *subscriptionsHandler) register() {
	h.service.bot.Handle(&btnMySubscriptions, h.handleMySubscriptions)
}

func (h *subscriptionsHandler) handleMySubscriptions(m *telebot.Message) {
	userID := domain.MustNewTelegramUserID(int64(m.Sender.ID))

	subs, err := h.service.subscriptionService.GetUserSubscriptions(userID)
	if err != nil {
		slog.Error("Failed to get subscriptions", "error", err)
		h.service.bot.Send(m.Sender, "Sorry, failed to get your subscriptions. Please try again.", subscriptionsMenu)
		return
	}

	if len(subs) == 0 {
		h.service.bot.Send(m.Sender, "You don't have any subscriptions yet.", subscriptionsMenu)
		return
	}

	// Send each subscription as a separate message with a Manage button
	for i, sub := range subs {
		project, err := h.service.projectService.GetByID(sub.ProjectID)
		if err != nil {
			slog.Error("Failed to get project details", "error", err, "project_id", sub.ProjectID)
			continue
		}
		
		// Create inline keyboard with Manage button
		markup := &telebot.ReplyMarkup{}
		btn := btnManageSubscription
		btn.Data = sub.ProjectID.String() // Store project ID in button data
		markup.InlineKeyboard = [][]telebot.InlineButton{
			{btn},
		}
		
		message := fmt.Sprintf("%d. <b>%s</b>", i+1, project.Name)
		
		// Add status indicators
		statusFlags := []string{}
		if sub.Muted {
			statusFlags = append(statusFlags, "🔕 Muted")
		}

		if sub.Paused() {
			statusFlags = append(statusFlags, "⏸️ Paused")
		}
		
		if len(statusFlags) > 0 {
			message += " [" + strings.Join(statusFlags, ", ") + "]"
		}
		
		h.service.bot.Send(m.Sender, message, &telebot.SendOptions{ParseMode: telebot.ModeHTML}, markup)
	}
	
	// Send the main menu after all subscriptions
	h.service.bot.Send(m.Sender, "Use the buttons below to manage your subscriptions.", subscriptionsMenu)
}

func (h *subscriptionsHandler) handleSubscriptionLink(m *telebot.Message, projectID uuid.UUID) error {
	project, err := h.service.projectService.GetByID(projectID)
	if err != nil {
		return fmt.Errorf("failed to get project by ID: %w", err)
	}

	userID := domain.MustNewTelegramUserID(int64(m.Sender.ID))
	err = h.service.subscriptionService.Subscribe(userID, project.ID)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			h.service.bot.Send(m.Sender, fmt.Sprintf("You are already subscribed to project <b>%s</b>", project.Name),
				&telebot.SendOptions{ParseMode: telebot.ModeHTML}, mainMenu)
			return nil
		}
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	h.service.bot.Send(m.Sender, fmt.Sprintf("You have successfully subscribed to project <b>%s</b>!", project.Name),
		&telebot.SendOptions{ParseMode: telebot.ModeHTML}, mainMenu)
	return nil
}
