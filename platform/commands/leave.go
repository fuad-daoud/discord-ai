package commands

import (
	"github.com/fuad-daoud/discord-ai/integrations/cohere"
	"github.com/fuad-daoud/discord-ai/platform"
	"golang.org/x/net/context"
)

func leave(call *cohere.CommandCall) {
	toolCall := call.ToolCall
	_, botStateOk := platform.Cache().VoiceState(call.ExtraProperties.GuildId, platform.Client().ApplicationID())
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
	conn := platform.Client().VoiceManager().GetConn(call.ExtraProperties.GuildId)
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
