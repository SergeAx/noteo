package bot

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/tucnak/telebot"

	"gitlab.com/trum/noteo/internal/domain"
)

type Service struct {
	bot                 *telebot.Bot
	projectService      *domain.ProjectService
	subscriptionService *domain.SubscriptionService
	stateManager        *StateManager

	mainMenu      *mainMenuHandler
	projects      *projectsHandler
	subscriptions *subscriptionsHandler
}

func NewService(
	cfg *Config,
	projectService *domain.ProjectService,
	subscriptionService *domain.SubscriptionService,
	stateManager *StateManager,
) (*Service, error) {
	bot, err := telebot.NewBot(telebot.Settings{
		Token:  cfg.Token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, err
	}

	service := &Service{
		bot:                 bot,
		projectService:      projectService,
		subscriptionService: subscriptionService,
		stateManager:        stateManager,
	}

	// Initialize handlers
	service.mainMenu = newMainMenuHandler(service)
	service.projects = newProjectsHandler(service)
	service.subscriptions = newSubscriptionsHandler(service)

	// Register handlers
	service.registerHandlers()

	return service, nil
}

func (s *Service) registerHandlers() {
	s.mainMenu.register()
	s.projects.register()
	s.subscriptions.register()
}

func (s *Service) getSubscriptionURL(projectID uuid.UUID) string {
	return fmt.Sprintf("t.me/%s?start=%s", s.bot.Me.Username, projectID)
}

func (s *Service) Start() {
	slog.Info("Starting Telegram bot", "username", s.bot.Me.Username, "url", "https://t.me/"+s.bot.Me.Username)
	s.bot.Start()
}
