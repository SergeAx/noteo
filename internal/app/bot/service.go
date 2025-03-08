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

	mainMenu               *mainMenuHandler
	projects               *projectsHandler
	subscriptions          *subscriptionsHandler
	subscriptionManagement *subscriptionManagementHandler
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
	service.subscriptionManagement = newSubscriptionManagementHandler(service)

	// Register handlers
	service.registerHandlers()

	return service, nil
}

// SendMessage implements the queue.MessageSender interface
func (s *Service) SendMessage(msg domain.Message) error {
	sendParams := &telebot.SendOptions{}
	if msg.Muted {
		sendParams.DisableNotification = true
	}

	_, err := s.bot.Send(&telebot.Chat{ID: msg.UserID.Int64()}, msg.Text, sendParams)
	return err
}

func (s *Service) registerHandlers() {
	s.mainMenu.register()
	s.projects.register()
	s.subscriptions.register()
	s.subscriptionManagement.register()
}

func (s *Service) getSubscriptionURL(projectID uuid.UUID) string {
	return fmt.Sprintf("https://t.me/%s?start=%s", s.bot.Me.Username, projectID.String())
}

func (s *Service) Start() {
	slog.Info("Starting Telegram bot", "username", s.bot.Me.Username, "url", "https://t.me/"+s.bot.Me.Username)
	s.bot.Start()
}
