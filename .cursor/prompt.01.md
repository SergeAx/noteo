Noteo: Telegram bot to send notifications
=========================================

Terms
-----

User: a Telegram user, receiving notifications
Project: an entity from which notifications are sent
Publisher: a Telegram user owns one or more Projects
Subscription: a link between user and project
Token: a long random string to authorize API requests

Logic for user
--------------

User recieves from publisher an URL with `/start` comand with the parameter that
designates a project. Is user is not subscribed to the prioject - it presented
with a confirmation dialogue. After confirming, user is subscribed to the project

When a User receives a notification from a Project, there are 3 Snoose buttons:
10 minutes, one hour, 24 hours, plus the button "Unsubscribe".

Logic for publisher
-------------------

First, Publisher should register with the bot with plain `/start` command. They
then should register projects by choosing an unique name for it. If the name is
already used by another Project of the same Publisher - bot returns an error.

After Project is registered, Publisher gets a Token to be used with API.

Publisher may list their Projects, and view/edit each of them.

Editing project allows changing its name and re-generating Token.

Viewing project allows to see current Token and list of Subscriptions

Publisher may call an API handle using Token as Bearer auth, and pass a JSON
structure with one field "body", containing Notification text. It should
support all Telegram emojis and text formatting.

Modules
-------

Project should be structured like this:

    internal\
      app\
        bot\ - code for Telegram bot functions
        api\ - code for REST API
        db\ - code for database storage handling
      domain\
        user\
        publisher\
        project\
        subscription\

Modules inside app\ directory should only handle input, validation and output,
calling modides inside domain\ directory for performing business logic.

Packages
--------

The project package should be `gitlab.com/trum/noteo`

* use `spf13/viper` to manage configuration
* use `tucnak/telebot` for bot API and logic
* use `net/http` for REST API
* use Gorm for data storage
* use `uber-go/dig` for dependency injection

Configuration
-------------

Using `spf13/viper` read config from env variables, .env file or comand line.
Use `NOTEO_` prefix for env variables.

* NOTEO_BOT_TOKEN - Telegram bot token
* NOTEO_PORT - port for REST API to listen on. Default 8080
* NOTEO_DB_DSN - DSN for Gorm database connection. Should start with `sqlite:`
or `postgres:`
