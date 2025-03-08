package bot

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/tucnak/telebot"

	"gitlab.com/trum/noteo/internal/domain"
)

// Menu items for subscription management
var (
	btnManageSubscription  = telebot.InlineButton{Unique: "manage_subscription", Text: "Manage"}
	btnMuteSubscription    = telebot.InlineButton{Unique: "mute_subscription", Text: "🔕 Mute"}
	btnUnmuteSubscription  = telebot.InlineButton{Unique: "unmute_subscription", Text: "🔔 Unmute"}
	btnPauseSubscription   = telebot.InlineButton{Unique: "pause_subscription", Text: "⏸️ Pause for 24 hours"}
	btnResumeSubscription  = telebot.InlineButton{Unique: "resume_subscription", Text: "▶️ Resume"}
	btnUnsubscribe         = telebot.InlineButton{Unique: "unsubscribe", Text: "❌ Unsubscribe"}
	btnResubscribe         = telebot.InlineButton{Unique: "resubscribe", Text: "↩️ Re-subscribe"}
	btnBackToSubscriptions = telebot.ReplyButton{Text: "Back to subscriptions"}

	subscriptionManagementMenu = &telebot.ReplyMarkup{
		ReplyKeyboard: [][]telebot.ReplyButton{
			{btnBackToSubscriptions},
		},
		ResizeReplyKeyboard: true,
	}
)

type subscriptionManagementHandler struct {
	service *Service
}

func newSubscriptionManagementHandler(s *Service) *subscriptionManagementHandler {
	return &subscriptionManagementHandler{service: s}
}

func (h *subscriptionManagementHandler) register() {
	h.service.bot.Handle(&btnBackToSubscriptions, h.handleBackToSubscriptions)
	h.service.bot.Handle(&btnManageSubscription, h.handleManageSubscription)
	h.service.bot.Handle(&btnMuteSubscription, h.handleMuteSubscription)
	h.service.bot.Handle(&btnUnmuteSubscription, h.handleUnmuteSubscription)
	h.service.bot.Handle(&btnPauseSubscription, h.handlePauseSubscription)
	h.service.bot.Handle(&btnResumeSubscription, h.handleResumeSubscription)
	h.service.bot.Handle(&btnUnsubscribe, h.handleUnsubscribe)
	h.service.bot.Handle(&btnResubscribe, h.handleResubscribe)
}

// handleBackToSubscriptions redirects to the subscriptions list
func (h *subscriptionManagementHandler) handleBackToSubscriptions(m *telebot.Message) {
	// Get the subscriptions handler from the service
	h.service.subscriptions.handleMySubscriptions(m)
}

// handleManageSubscription handles the Manage button click for a subscription
func (h *subscriptionManagementHandler) handleManageSubscription(c *telebot.Callback) {
	// Extract project ID from callback data
	projectID, err := uuid.Parse(c.Data)
	if err != nil {
		slog.Error("Invalid project ID in manage subscription callback", "error", err, "data", c.Data)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Invalid subscription. Please try again."})
		return
	}

	// Get project details
	project, err := h.service.projectService.GetByID(projectID)
	if err != nil {
		slog.Error("Failed to get project details", "error", err, "project_id", projectID)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Failed to get subscription details. Please try again."})
		return
	}

	// Get subscription details
	userID := domain.MustNewTelegramUserID(int64(c.Sender.ID))
	subscription, err := h.service.subscriptionService.GetUserSubscriptions(userID)
	if err != nil {
		slog.Error("Failed to get user subscriptions", "error", err)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Failed to get subscription details. Please try again."})
		return
	}

	// Find the specific subscription
	var sub *domain.Subscription
	for _, s := range subscription {
		if s.ProjectID == projectID {
			sub = s
			break
		}
	}

	if sub == nil {
		slog.Error("Subscription not found", "project_id", projectID, "user_id", userID)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Subscription not found. Please try again."})
		return
	}

	// Respond to the callback to remove the loading indicator
	h.service.bot.Respond(c, &telebot.CallbackResponse{})

	// Create inline keyboard with management options
	inlineMarkup := &telebot.ReplyMarkup{}

	// Mute/Unmute button
	var muteBtn telebot.InlineButton
	if sub.Muted {
		muteBtn = btnUnmuteSubscription
	} else {
		muteBtn = btnMuteSubscription
	}
	muteBtn.Data = projectID.String()

	// Pause/Resume button
	var pauseBtn telebot.InlineButton
	now := time.Now()
	isPaused := sub.PausedUntil != nil && sub.PausedUntil.After(now)
	if isPaused {
		pauseBtn = btnResumeSubscription
	} else {
		pauseBtn = btnPauseSubscription
	}
	pauseBtn.Data = projectID.String()

	// Unsubscribe button
	unsubBtn := btnUnsubscribe
	unsubBtn.Data = projectID.String()

	inlineMarkup.InlineKeyboard = [][]telebot.InlineButton{
		{muteBtn},
		{pauseBtn},
		{unsubBtn},
	}

	// Status message
	statusMsg := ""
	if sub.Muted {
		statusMsg += "🔕 Notifications are currently muted\n"
	} else {
		statusMsg += "🔔 Notifications are currently enabled\n"
	}

	if isPaused {
		statusMsg += "⏸️ Notifications are paused until " + sub.PausedUntil.Format("Jan 2, 2006 15:04")
	} else {
		statusMsg += "▶️ Notifications are active"
	}

	message := fmt.Sprintf("Managing subscription to <b>%s</b>\n\n%s", project.Name, statusMsg)

	// Send message with both inline buttons and reply keyboard
	_, err = h.service.bot.Send(c.Sender, message, &telebot.SendOptions{ParseMode: telebot.ModeHTML}, inlineMarkup)
	if err != nil {
		slog.Error("Failed to send subscription management message", "error", err)
	}

	// Send the reply keyboard separately
	_, err = h.service.bot.Send(c.Sender, "Use the buttons below to manage your subscription:", subscriptionManagementMenu)
	if err != nil {
		slog.Error("Failed to send subscription management menu", "error", err)
	}
}

// updateSubscriptionMessage updates the subscription management message with current status
func (h *subscriptionManagementHandler) updateSubscriptionMessage(c *telebot.Callback, projectID uuid.UUID) {
	// Get project details
	project, err := h.service.projectService.GetByID(projectID)
	if err != nil {
		slog.Error("Failed to get project details", "error", err, "project_id", projectID)
		return
	}

	// Get subscription details
	userID := domain.MustNewTelegramUserID(int64(c.Sender.ID))
	subscriptions, err := h.service.subscriptionService.GetUserSubscriptions(userID)
	if err != nil {
		slog.Error("Failed to get user subscriptions", "error", err)
		return
	}

	// Find the specific subscription
	var sub *domain.Subscription
	for _, s := range subscriptions {
		if s.ProjectID == projectID {
			sub = s
			break
		}
	}

	if sub == nil {
		slog.Error("Subscription not found", "project_id", projectID, "user_id", userID)
		return
	}

	// Create inline keyboard with management options
	inlineMarkup := &telebot.ReplyMarkup{}

	// Mute/Unmute button
	var muteBtn telebot.InlineButton
	if sub.Muted {
		muteBtn = btnUnmuteSubscription
	} else {
		muteBtn = btnMuteSubscription
	}
	muteBtn.Data = projectID.String()

	// Pause/Resume button
	var pauseBtn telebot.InlineButton
	now := time.Now()
	isPaused := sub.PausedUntil != nil && sub.PausedUntil.After(now)
	if isPaused {
		pauseBtn = btnResumeSubscription
	} else {
		pauseBtn = btnPauseSubscription
	}
	pauseBtn.Data = projectID.String()

	// Unsubscribe button
	unsubBtn := btnUnsubscribe
	unsubBtn.Data = projectID.String()

	inlineMarkup.InlineKeyboard = [][]telebot.InlineButton{
		{muteBtn},
		{pauseBtn},
		{unsubBtn},
	}

	// Status message
	statusMsg := ""
	if sub.Muted {
		statusMsg += "🔕 Notifications are currently muted\n"
	} else {
		statusMsg += "🔔 Notifications are currently enabled\n"
	}

	if isPaused {
		statusMsg += "⏸️ Notifications are paused until " + sub.PausedUntil.Format("Jan 2, 2006 15:04")
	} else {
		statusMsg += "▶️ Notifications are active"
	}

	message := fmt.Sprintf("Managing subscription to <b>%s</b>\n\n%s", project.Name, statusMsg)

	// Update the message with inline buttons
	_, err = h.service.bot.Edit(c.Message, message, &telebot.SendOptions{ParseMode: telebot.ModeHTML}, inlineMarkup)
	if err != nil {
		slog.Error("Failed to update subscription management message", "error", err)
	}
}

// handleMuteSubscription handles muting a subscription
func (h *subscriptionManagementHandler) handleMuteSubscription(c *telebot.Callback) {
	projectID, err := uuid.Parse(c.Data)
	if err != nil {
		slog.Error("Invalid project ID in mute subscription callback", "error", err, "data", c.Data)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Invalid subscription. Please try again."})
		return
	}

	// Mute the subscription
	userID := domain.MustNewTelegramUserID(int64(c.Sender.ID))
	if err := h.service.subscriptionService.MuteNotifications(userID, projectID); err != nil {
		slog.Error("Failed to mute subscription", "error", err)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Failed to mute subscription. Please try again."})
		return
	}

	// Respond to the callback to remove the loading indicator
	h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Subscription muted"})

	// Update the message with new status and inline buttons
	h.updateSubscriptionMessage(c, projectID)

	// Send a confirmation with the subscription management menu
	h.service.bot.Send(c.Sender, "Notifications have been muted.", subscriptionManagementMenu)
}

// handleUnmuteSubscription handles unmuting a subscription
func (h *subscriptionManagementHandler) handleUnmuteSubscription(c *telebot.Callback) {
	projectID, err := uuid.Parse(c.Data)
	if err != nil {
		slog.Error("Invalid project ID in unmute subscription callback", "error", err, "data", c.Data)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Invalid subscription. Please try again."})
		return
	}

	// Unmute the subscription
	userID := domain.MustNewTelegramUserID(int64(c.Sender.ID))
	if err := h.service.subscriptionService.UnmuteNotifications(userID, projectID); err != nil {
		slog.Error("Failed to unmute subscription", "error", err)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Failed to unmute subscription. Please try again."})
		return
	}

	// Respond to the callback to remove the loading indicator
	h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Subscription unmuted"})

	// Update the message with new status and inline buttons
	h.updateSubscriptionMessage(c, projectID)

	// Send a confirmation with the subscription management menu
	h.service.bot.Send(c.Sender, "Notifications have been unmuted.", subscriptionManagementMenu)
}

// handlePauseSubscription handles pausing a subscription
func (h *subscriptionManagementHandler) handlePauseSubscription(c *telebot.Callback) {
	projectID, err := uuid.Parse(c.Data)
	if err != nil {
		slog.Error("Invalid project ID in pause subscription callback", "error", err, "data", c.Data)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Invalid subscription. Please try again."})
		return
	}

	// Pause the subscription for 24 hours
	userID := domain.MustNewTelegramUserID(int64(c.Sender.ID))
	pauseUntil := time.Now().Add(24 * time.Hour)
	if err := h.service.subscriptionService.PauseNotifications(userID, projectID, pauseUntil); err != nil {
		slog.Error("Failed to pause subscription", "error", err)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Failed to pause subscription. Please try again."})
		return
	}

	// Respond to the callback to remove the loading indicator
	h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Subscription paused for 24 hours"})

	// Update the message with new status and inline buttons
	h.updateSubscriptionMessage(c, projectID)

	// Send a confirmation with the subscription management menu
	h.service.bot.Send(c.Sender, "Notifications have been paused for 24 hours.", subscriptionManagementMenu)
}

// handleResumeSubscription handles resuming a subscription
func (h *subscriptionManagementHandler) handleResumeSubscription(c *telebot.Callback) {
	projectID, err := uuid.Parse(c.Data)
	if err != nil {
		slog.Error("Invalid project ID in resume subscription callback", "error", err, "data", c.Data)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Invalid subscription. Please try again."})
		return
	}

	// Resume the subscription
	userID := domain.MustNewTelegramUserID(int64(c.Sender.ID))
	if err := h.service.subscriptionService.ResumeNotifications(userID, projectID); err != nil {
		slog.Error("Failed to resume subscription", "error", err)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Failed to resume subscription. Please try again."})
		return
	}

	// Respond to the callback to remove the loading indicator
	h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Subscription resumed"})

	// Update the message with new status and inline buttons
	h.updateSubscriptionMessage(c, projectID)

	// Send a confirmation with the subscription management menu
	h.service.bot.Send(c.Sender, "Notifications have been resumed.", subscriptionManagementMenu)
}

// handleUnsubscribe handles unsubscribing from a project
func (h *subscriptionManagementHandler) handleUnsubscribe(c *telebot.Callback) {
	projectID, err := uuid.Parse(c.Data)
	if err != nil {
		slog.Error("Invalid project ID in unsubscribe callback", "error", err, "data", c.Data)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Invalid subscription. Please try again."})
		return
	}

	// Get project details
	project, err := h.service.projectService.GetByID(projectID)
	if err != nil {
		slog.Error("Failed to get project details", "error", err, "project_id", projectID)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Failed to get project details. Please try again."})
		return
	}

	// Unsubscribe the user
	userID := domain.MustNewTelegramUserID(int64(c.Sender.ID))
	if err := h.service.subscriptionService.Unsubscribe(userID, projectID); err != nil {
		slog.Error("Failed to unsubscribe", "error", err)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Failed to unsubscribe. Please try again."})
		return
	}

	// Respond to the callback to remove the loading indicator
	h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Unsubscribed successfully"})

	// Create inline keyboard with re-subscribe button
	inlineMarkup := &telebot.ReplyMarkup{}
	resubBtn := btnResubscribe
	resubBtn.Data = projectID.String()
	inlineMarkup.InlineKeyboard = [][]telebot.InlineButton{
		{resubBtn},
	}

	// Update the message
	message := fmt.Sprintf("You have unsubscribed from <b>%s</b>", project.Name)
	_, err = h.service.bot.Edit(c.Message, message, &telebot.SendOptions{ParseMode: telebot.ModeHTML}, inlineMarkup)
	if err != nil {
		slog.Error("Failed to update subscription message after unsubscribe", "error", err)
	}

	// Send a confirmation with the subscription management menu
	h.service.bot.Send(c.Sender, "You have been unsubscribed from the project.", subscriptionManagementMenu)
}

// handleResubscribe handles re-subscribing to a project
func (h *subscriptionManagementHandler) handleResubscribe(c *telebot.Callback) {
	projectID, err := uuid.Parse(c.Data)
	if err != nil {
		slog.Error("Invalid project ID in resubscribe callback", "error", err, "data", c.Data)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Invalid project. Please try again."})
		return
	}

	// Re-subscribe the user
	userID := domain.MustNewTelegramUserID(int64(c.Sender.ID))
	if err := h.service.subscriptionService.Subscribe(userID, projectID); err != nil {
		slog.Error("Failed to resubscribe", "error", err)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Failed to resubscribe. Please try again."})
		return
	}

	// Respond to the callback to remove the loading indicator
	h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Re-subscribed successfully"})

	// Update the message with the regular subscription management view
	h.updateSubscriptionMessage(c, projectID)

	// Send a confirmation with the subscription management menu
	h.service.bot.Send(c.Sender, "You have been re-subscribed to the project.", subscriptionManagementMenu)
}
