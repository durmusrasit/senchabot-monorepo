package webhook

import (
	"net/http"

	"github.com/senchabot-dev/monorepo/apps/twitch-bot/client"
)

type Webhook interface {
	BotJoin(client *client.Clients, joinedChannelList []string, w http.ResponseWriter, r *http.Request)
}

type Webhooks struct {
}

func NewWebhooks() Webhook {
	return &Webhooks{}
}
