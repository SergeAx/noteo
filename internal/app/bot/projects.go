package bot

import (
	"fmt"
	"log/slog"

	"github.com/tucnak/telebot"

	"github.com/sergeax/noteo/internal/domain"
)

var (
	btnMyProjects = telebot.ReplyButton{Text: "My Projects"}
	btnCreateNew  = telebot.ReplyButton{Text: "Create new"}
	btnCancel     = telebot.ReplyButton{Text: "Cancel"}

	projectsMenu = &telebot.ReplyMarkup{
		ReplyKeyboard: [][]telebot.ReplyButton{
			{btnCreateNew, btnBackToMenu},
		},
		ResizeReplyKeyboard: true,
	}

	cancelMenu = &telebot.ReplyMarkup{
		ReplyKeyboard: [][]telebot.ReplyButton{
			{btnCancel},
		},
		ResizeReplyKeyboard: true,
	}
)

type projectsHandler struct {
	service *Service
}

func newProjectsHandler(s *Service) *projectsHandler {
	return &projectsHandler{service: s}
}

func (h *projectsHandler) register() {
	h.service.bot.Handle(&btnMyProjects, h.handleMyProjects)
	h.service.bot.Handle(&btnCreateNew, h.handleCreateProject)
	h.service.bot.Handle(&btnCancel, h.handleCancel)
}

func (h *projectsHandler) handleMyProjects(m *telebot.Message) {
	userID := domain.MustNewTelegramUserID(int64(m.Sender.ID))
	projects, err := h.service.projectService.GetByPublisher(userID)
	if err != nil {
		slog.Error("Failed to get projects", "error", err)
		h.service.bot.Send(m.Sender, "Sorry, failed to get your projects. Please try again.", projectsMenu)
		return
	}

	if len(projects) == 0 {
		h.service.bot.Send(m.Sender, "You don't have any projects yet.", projectsMenu)
		return
	}

	var message string
	for i, project := range projects {
		message += fmt.Sprintf("%d. <b>%s</b>\n   Token: <code>%s</code>\n   Share link: %s\n\n",
			i+1, project.Name, project.Token, h.service.getSubscriptionURL(project.ID))
	}

	h.service.bot.Send(m.Sender, message, &telebot.SendOptions{ParseMode: telebot.ModeHTML}, projectsMenu)
}

func (h *projectsHandler) handleCreateProject(m *telebot.Message) {
	h.service.stateManager.SetState(m.Sender.ID, StateCreatingProject, nil)
	h.service.bot.Send(m.Sender, "Please enter the name for your new project:", cancelMenu)
}

func (h *projectsHandler) handleCancel(m *telebot.Message) {
	h.service.stateManager.ClearState(m.Sender.ID)
	h.service.bot.Send(m.Sender, "Operation cancelled.", projectsMenu)
}

func (h *projectsHandler) handleProjectCreation(m *telebot.Message) error {
	userID := domain.MustNewTelegramUserID(int64(m.Sender.ID))
	project, err := h.service.projectService.Create(userID, m.Text)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	message := fmt.Sprintf("Project created successfully!\n\n<b>Name:</b> %s\n<b>Token:</b> <code>%s</code>\n\n"+
		"Share this link to let users subscribe to your project:\n%s",
		project.Name, project.Token, h.service.getSubscriptionURL(project.ID))

	h.service.bot.Send(m.Sender, message, &telebot.SendOptions{ParseMode: telebot.ModeHTML}, projectsMenu)
	return nil
}
