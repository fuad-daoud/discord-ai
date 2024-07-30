package commands

import (
	"github.com/disgoorg/disgo/voice"
	"github.com/fuad-daoud/discord-ai/integrations/cohere"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"github.com/fuad-daoud/discord-ai/platform"
	"golang.org/x/net/context"
)

type Command string

const (
	Join            Command = "command_join"
	Leave           Command = "command_leave"
	Play            Command = "command_play"
	Pause           Command = "command_pause"
	Stop            Command = "command_stop"
	Resume          Command = "command_resume"
	Skip            Command = "command_skip"
	Search          Command = "command_search"
	Queue           Command = "command_queue"
	ClearQueue      Command = "command_clear_queue"
	ToggleLoopQueue Command = "command_toggle_loop_queue"
	ToggleLoop      Command = "command_toggle_loop"
)

func AddCommandsChannelOnReadyHandler() {
	go func() {
		//call := &cohere.CommandCall{
		//	ToolCall: nil,
		//	ExtraProperties: cohere.Properties{
		//		MessageId: "1262490536577601777",
		//		UserId:    468494540852953089,
		//		GuildId:   847908927554322432,
		//	},
		//}
		//play(call)
		for call := range cohere.Call {
			go runCommand(call)
		}
	}()

}

func runCommand(call *cohere.CommandCall) {
	defer rec()
	switch Command(call.Name) {
	case Join:
		join(call)
		break
	case Leave:
		leave(call)
		break
	case Play:
		play(call)
		break
	case Pause:
		pause(call)
		break
	case Stop:
		stop(call)
		break
	case Resume:
		resume(call)
		break
	case Skip:
		skip(call)
		break
	case Queue:
		queue(call)
		break
	case ClearQueue:
		clearQueue(call)
		break
	case ToggleLoopQueue:
		toggleLoopQueue(call)
		break
	case ToggleLoop:
		toggleLoop(call)
		break
	case Search:
		search(call)
		break
	default:
		{
			cohere.Result <- &cohere.CommandResult{
				Call: call.ToolCall,
				Outputs: []map[string]interface{}{
					{
						"Success":     false,
						"Description": "command not implemented",
					},
				},
			}
			break
		}
	}
}

func join(call *cohere.CommandCall) {
	dlog.Log.Info("starting join function")
	voiceState, b := platform.Cache().VoiceState(call.ExtraProperties.GuildId, call.ExtraProperties.UserId)
	if !b {
		cohere.Result <- &cohere.CommandResult{
			Call: call.ToolCall,
			Outputs: []map[string]interface{}{
				{
					"Success":     false,
					"Description": "user is not in a voice channel, the user shall be in a voice channel first",
				},
			},
		}
		return
	}

	botState, botStateOk := platform.Cache().VoiceState(call.ExtraProperties.GuildId, platform.Client().ApplicationID())
	userState, userStateOk := platform.Cache().VoiceState(call.ExtraProperties.GuildId, call.ExtraProperties.UserId)

	if botStateOk && userStateOk && (botState.ChannelID.String() == userState.ChannelID.String()) {
		cohere.Result <- &cohere.CommandResult{
			Call: call.ToolCall,
			Outputs: []map[string]interface{}{
				{
					"Success":     false,
					"Description": "already in the voice channel",
				},
			},
		}
		return
	}

	dlog.Log.Info("joining voice channel")

	conn := platform.Client().VoiceManager().CreateConn(call.ExtraProperties.GuildId)

	if err := conn.Open(context.Background(), *voiceState.ChannelID, false, false); err != nil {
		dlog.Log.Error("error opening voice channel", "error", err)
		cohere.Result <- &cohere.CommandResult{
			Call: call.ToolCall,
			Outputs: []map[string]interface{}{
				{
					"Success":     false,
					"Description": "error connecting to voice channel",
				},
			},
		}
		return
	}
	if err := conn.SetSpeaking(context.Background(), voice.SpeakingFlagMicrophone); err != nil {
		dlog.Log.Error("error setting speaking flag", "error", err)
		cohere.Result <- &cohere.CommandResult{
			Call: call.ToolCall,
			Outputs: []map[string]interface{}{
				{
					"Success":     false,
					"Description": "error setting speaking flag",
				},
			},
		}
		return
	}
	dlog.Log.Info("set speaking successfully")
	if _, err := conn.UDP().Write(voice.SilenceAudioFrame); err != nil {
		dlog.Log.Error("failed to write silence audio frame", "error", err)
		cohere.Result <- &cohere.CommandResult{
			Call: call.ToolCall,
			Outputs: []map[string]interface{}{
				{
					"Success":     false,
					"Description": "error sending silence",
				},
			},
		}
		return
	}
	dlog.Log.Info("wrote silent frame successfully")

	go platform.HandleDeepgramVoicePackets(conn, call.ExtraProperties)

	cohere.Result <- &cohere.CommandResult{
		Call: call.ToolCall,
		Outputs: []map[string]interface{}{
			{
				"Success":     true,
				"Description": "connected",
			},
		},
	}

	//audioProvider, err := elevenlabs.TTS("I am here !!")
	//if err != nil {
	//	dlog.Log.Error("could not get audio provider", "error", err)
	//	return
	//}

	//conn.SetOpusFrameProvider(provider)

	dlog.Log.Info("Finished joining function")
	return
}

func rec() {
	if r := recover(); r != nil {
		dlog.Log.Error("Recovered ", "msg", r)
	}
}
