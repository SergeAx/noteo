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
			{btnBackToSubscriptions, btnBackToMenu},
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
	h.service.bot.Handle(&btnBackToSubscriptions, h.service.subscriptions.handleMySubscriptions)
	h.service.bot.Handle(&btnManageSubscription, h.handleManageSubscription)
	h.service.bot.Handle(&btnMuteSubscription, h.handleMuteSubscription)
	h.service.bot.Handle(&btnUnmuteSubscription, h.handleUnmuteSubscription)
	h.service.bot.Handle(&btnPauseSubscription, h.handlePauseSubscription)
	h.service.bot.Handle(&btnResumeSubscription, h.handleResumeSubscription)
	h.service.bot.Handle(&btnUnsubscribe, h.handleUnsubscribe)
	h.service.bot.Handle(&btnResubscribe, h.handleResubscribe)
}

// parseProjectID parses a project ID from callback data and handles errors
func (h *subscriptionManagementHandler) parseProjectID(c *telebot.Callback, action string) (uuid.UUID, bool) {
	projectID, err := uuid.Parse(c.Data)
	if err != nil {
		slog.Error("Invalid project ID in "+action+" callback", "error", err, "data", c.Data)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Invalid subscription. Please try again."})
		return uuid.Nil, false
	}
	return projectID, true
}

// getUserID extracts the user ID from a callback
func (h *subscriptionManagementHandler) getUserID(c *telebot.Callback) domain.TelegramUserID {
	return domain.MustNewTelegramUserID(int64(c.Sender.ID))
}

func (h *subscriptionManagementHandler) findSubscription(userID domain.TelegramUserID, projectID uuid.UUID) (*domain.Subscription, *domain.Project, error) {
	project, err := h.service.projectService.GetByID(projectID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get project details: %w", err)
	}

	subscriptions, err := h.service.subscriptionService.GetUserSubscriptions(userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user subscriptions: %w", err)
	}

	for _, s := range subscriptions {
		if s.ProjectID == projectID {
			return s, project, nil
		}
	}

	return nil, nil, fmt.Errorf("subscription not found")
}

// createSubscriptionButtons creates the inline keyboard buttons for subscription management
func (h *subscriptionManagementHandler) createSubscriptionButtons(sub *domain.Subscription, projectID uuid.UUID) *telebot.ReplyMarkup {
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
	if sub.Paused() {
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

	return inlineMarkup
}

// createResubscribeButton creates an inline keyboard with just the resubscribe button
func (h *subscriptionManagementHandler) createResubscribeButton(projectID uuid.UUID) *telebot.ReplyMarkup {
	inlineMarkup := &telebot.ReplyMarkup{}
	resubBtn := btnResubscribe
	resubBtn.Data = projectID.String()
	inlineMarkup.InlineKeyboard = [][]telebot.InlineButton{
		{resubBtn},
	}
	return inlineMarkup
}

// createStatusMessage creates a status message for a subscription
func (h *subscriptionManagementHandler) createStatusMessage(sub *domain.Subscription, project *domain.Project) string {
	// Status message
	statusMsg := ""
	if sub.Muted {
		statusMsg += "🔕 Notifications are currently muted\n"
	} else {
		statusMsg += "🔔 Notifications are currently enabled\n"
	}

	if sub.Paused() {
		statusMsg += "⏸️ Notifications are paused until " + sub.PausedUntil.Format("Jan 2, 2006 15:04")
	} else {
		statusMsg += "▶️ Notifications are active"
	}

	return fmt.Sprintf("Managing subscription to <b>%s</b>\n\n%s", project.Name, statusMsg)
}

// handleSubscriptionAction performs a subscription action and handles common response patterns
func (h *subscriptionManagementHandler) handleSubscriptionAction(
	c *telebot.Callback,
	action string,
	actionFunc func(userID domain.TelegramUserID, projectID uuid.UUID) error,
	successMessage string,
	confirmationMessage string,
) {
	projectID, ok := h.parseProjectID(c, action)
	if !ok {
		return
	}

	userID := h.getUserID(c)
	if err := actionFunc(userID, projectID); err != nil {
		slog.Error("Failed to "+action+" subscription", "error", err)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Failed to " + action + " subscription. Please try again."})
		return
	}

	// Respond to the callback to remove the loading indicator
	h.service.bot.Respond(c, &telebot.CallbackResponse{Text: successMessage})

	// Update the message with new status and inline buttons
	h.updateSubscriptionMessage(c, projectID)

	// Send a confirmation with the subscription management menu
	h.service.bot.Send(c.Sender, confirmationMessage, subscriptionManagementMenu)
}

// handleManageSubscription handles the Manage button click for a subscription
func (h *subscriptionManagementHandler) handleManageSubscription(c *telebot.Callback) {
	// Extract project ID from callback data
	projectID, ok := h.parseProjectID(c, "manage subscription")
	if !ok {
		return
	}

	userID := h.getUserID(c)
	sub, project, err := h.findSubscription(userID, projectID)
	if err != nil {
		slog.Error("Failed to find subscription", "error", err)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Failed to get subscription details. Please try again."})
		return
	}

	// Respond to the callback to remove the loading indicator
	h.service.bot.Respond(c, &telebot.CallbackResponse{})

	// Create inline keyboard with management options
	inlineMarkup := h.createSubscriptionButtons(sub, projectID)

	// Create status message
	message := h.createStatusMessage(sub, project)

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
	userID := h.getUserID(c)
	sub, project, err := h.findSubscription(userID, projectID)
	if err != nil {
		slog.Error("Failed to find subscription for update", "error", err)
		return
	}

	// Create inline keyboard with management options
	inlineMarkup := h.createSubscriptionButtons(sub, projectID)

	// Create status message
	message := h.createStatusMessage(sub, project)

	// Update the message with inline buttons
	_, err = h.service.bot.Edit(c.Message, message, &telebot.SendOptions{ParseMode: telebot.ModeHTML}, inlineMarkup)
	if err != nil {
		slog.Error("Failed to update subscription management message", "error", err)
	}
}

// handleMuteSubscription handles muting a subscription
func (h *subscriptionManagementHandler) handleMuteSubscription(c *telebot.Callback) {
	h.handleSubscriptionAction(
		c,
		"mute",
		h.service.subscriptionService.MuteNotifications,
		"Subscription muted",
		"Notifications have been muted.",
	)
}

// handleUnmuteSubscription handles unmuting a subscription
func (h *subscriptionManagementHandler) handleUnmuteSubscription(c *telebot.Callback) {
	h.handleSubscriptionAction(
		c,
		"unmute",
		h.service.subscriptionService.UnmuteNotifications,
		"Subscription unmuted",
		"Notifications have been unmuted.",
	)
}

// handlePauseSubscription handles pausing a subscription
func (h *subscriptionManagementHandler) handlePauseSubscription(c *telebot.Callback) {
	projectID, ok := h.parseProjectID(c, "pause subscription")
	if !ok {
		return
	}

	// Pause the subscription for 24 hours
	userID := h.getUserID(c)
	pauseUntil := time.Now().Add(24 * time.Hour)

	// We need a custom function to handle the additional pauseUntil parameter
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
	h.handleSubscriptionAction(
		c,
		"resume",
		h.service.subscriptionService.ResumeNotifications,
		"Subscription resumed",
		"Notifications have been resumed.",
	)
}

// handleUnsubscribe handles unsubscribing from a project
func (h *subscriptionManagementHandler) handleUnsubscribe(c *telebot.Callback) {
	projectID, ok := h.parseProjectID(c, "unsubscribe")
	if !ok {
		return
	}

	userID := h.getUserID(c)

	// Get project details before unsubscribing
	project, err := h.service.projectService.GetByID(projectID)
	if err != nil {
		slog.Error("Failed to get project details", "error", err, "project_id", projectID)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Failed to get project details. Please try again."})
		return
	}

	// Unsubscribe the user
	if err := h.service.subscriptionService.Unsubscribe(userID, projectID); err != nil {
		slog.Error("Failed to unsubscribe", "error", err)
		h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Failed to unsubscribe. Please try again."})
		return
	}

	// Respond to the callback to remove the loading indicator
	h.service.bot.Respond(c, &telebot.CallbackResponse{Text: "Unsubscribed successfully"})

	// Create inline keyboard with re-subscribe button
	inlineMarkup := h.createResubscribeButton(projectID)

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
	h.handleSubscriptionAction(
		c,
		"resubscribe",
		h.service.subscriptionService.Subscribe,
		"Re-subscribed successfully",
		"You have been re-subscribed to the project.",
	)
}
