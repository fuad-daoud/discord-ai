package events

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type GuildMessageUpdate struct {
	*events.GenericGuildMessage
	OldMessage discord.Message
}
