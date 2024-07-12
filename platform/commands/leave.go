package commands

import (
	"github.com/disgoorg/snowflake/v2"
	"github.com/fuad-daoud/discord-ai/integrations/cohere"
	"github.com/fuad-daoud/discord-ai/platform"
	"golang.org/x/net/context"
)

func leave(call *cohere.CommandCall) {
	toolCall := call.ToolCall
	guildId := snowflake.MustParse(call.ExtraProperties.GuildId)
	_, botStateOk := platform.Cache().VoiceState(guildId, platform.Client().ApplicationID())
	if !botStateOk {
		cohere.Result <- &cohere.CommandResult{
			Call: toolCall,
			Outputs: []map[string]interface{}{
				{
					"Success":     false,
					"Description": "not in a voice channel",
				},
			},
		}
		return
	}
	conn := platform.Client().VoiceManager().GetConn(guildId)
	conn.Close(context.Background())

	cohere.Result <- &cohere.CommandResult{
		Call: toolCall,
		Outputs: []map[string]interface{}{
			{
				"Success":     true,
				"Description": "left voice channel",
			},
		},
	}
}
