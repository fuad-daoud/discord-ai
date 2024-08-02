package platform

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	db2 "github.com/fuad-daoud/discord-ai/layers/db"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"strings"
)

func dbThreadCreateHandler(event *events.ThreadCreate) {
	dlog.Log.Debug("Starting db ThreadCreateHandler")
	if event.Thread.OwnerID != event.Client().ApplicationID() {
		return
	}
	err := db2.Transaction(func(write db2.Write) error {
		write(db2.MergeN("t", db2.Thread{
			TextChannel: db2.TextChannel{
				Id:          event.ThreadID.String(),
				Name:        event.Thread.Name(),
				CreatedDate: event.Thread.CreatedAt().String(),
			},
		}))

		write(db2.MatchN("t", db2.Thread{TextChannel: db2.TextChannel{Id: event.ThreadID.String()}}),
			db2.MatchN("m", db2.Message{Id: event.ThreadID.String()}),
			db2.Merge("(m)-[:STARTED]->(t)-[:CONTAINS]->(m)"))

		write(db2.MatchN("mb", db2.Member{Id: event.Thread.OwnerID.String()}),
			db2.MatchN("t", db2.Thread{
				TextChannel: db2.TextChannel{
					Id: event.Thread.ID().String(),
				},
			}),
			db2.Merge("(mb)-[:CREATED]->(t)"),
		)
		return nil
	})
	if err != nil {
		dlog.Log.Error("Error in dbThreadCreateHandler: ", err)
		return
	}
	dlog.Log.Debug("Finished db ThreadCreateHandler")
}

func dbMessageUpdateHandler(event *events.GuildMessageUpdate) {
	dlog.Log.Debug("Starting db MessageUpdateHandler")
	if event.Message.Flags == discord.MessageFlagHasThread && event.OldMessage.ChannelID == event.ChannelID {
		return
	}

	result, err := db2.Query(db2.MatchN("m", db2.Message{Id: event.Message.ID.String()}), db2.Return("m"))
	if err != nil {
		dlog.Log.Error("Error in dbMessageUpdateHandler: ", err)
		return
	}
	oldMessage, _ := db2.ParseKey[db2.Message]("m", result.Records)

	err = db2.Transaction(func(write db2.Write) error {
		set, err := db2.Set("m", db2.Message{
			Id:          oldMessage.Id,
			Text:        strings.ReplaceAll(event.Message.Content, "\"", "'"),
			UpdatedDate: event.Message.CreatedAt.String(),
			CreatedDate: oldMessage.CreatedDate,
		})
		if err != nil {
			return err
		}
		return write(db2.MatchN("m", oldMessage),
			set,
			db2.Return("m"))
	})
	if err != nil {
		dlog.Log.Error("Error in dbMessageUpdateHandler: ", err)
		return
	}
	dlog.Log.Debug("Finished db MessageUpdateHandler")
}

func dbMessageCreateHandler(event *events.GuildMessageCreate) {
	if event.Message.Type == discord.MessageTypeThreadStarterMessage {
		return
	}
	dlog.Log.Debug("Starting db messageCreateHandler")
	restClient := event.Client().Rest()
	channel, _ := restClient.GetChannel(event.ChannelID)

	message := db2.Message{
		Id:          event.Message.ID.String(),
		Text:        strings.ReplaceAll(event.Message.Content, "\"", "'"),
		CreatedDate: event.Message.CreatedAt.String(),
	}

	member := db2.Member{
		Id: event.Message.Author.ID.String(),
	}
	err := db2.Transaction(func(write db2.Write) error {
		if channel.Type() == discord.ChannelTypeGuildPublicThread {
			threadMessageCreateHandler(write, channel, member, message)
		} else {
			createMessageHandler(write, channel, event.GuildID.String(), member, message)
		}
		return nil
	})
	if err != nil {
		dlog.Log.Error("Error in dbMessageCreateHandler: ", err)
		return
	}
	dlog.Log.Debug("Finished db messageCreateHandler")
}

func createMessageHandler(write db2.Write, channel discord.Channel, guildId string, member db2.Member, message db2.Message) {
	textChannel := db2.TextChannel{
		Id:   channel.ID().String(),
		Name: channel.Name(),
	}
	dbGuild := db2.Guild{Id: guildId}
	write(db2.MatchN("g", dbGuild),
		db2.MatchN("mb", member),
		db2.MergeN("c", textChannel),
		db2.CreateN("m", message),
		db2.Merge("(g)-[:HAS]->(c)"),
		db2.Merge("(c)-[:CONTAINS]->(m)-[:AUTHOR]->(mb)-[:CREATED]->(m)"))
}

func threadMessageCreateHandler(write db2.Write, channel discord.Channel, member db2.Member, message db2.Message) {
	textChannel := db2.TextChannel{
		Id: channel.(discord.GuildThread).ParentID().String(),
	}
	dbThread := db2.Thread{
		TextChannel: db2.TextChannel{
			Id:   channel.(discord.GuildThread).ID().String(),
			Name: channel.(discord.GuildThread).Name(),
		},
	}
	write(db2.MatchN("c", textChannel),
		db2.MatchN("mb", member),
		db2.MergeN("t", dbThread),
		db2.CreateN("m", message),
		db2.Merge("(c)-[:CHILD]->(t)-[:CONTAINS]->(m)-[:AUTHOR]->(mb)-[:CREATED]->(m)"),
		db2.Merge("(t)-[:PARENT]->(c)"))
}

func dbReadyHandler(event *events.Ready) {
	err := db2.Transaction(func(write db2.Write) error {
		for _, guild := range event.Guilds {
			dlog.Log.Debug("Merging guild", "ID", guild.ID)
			guild, err := event.Client().Rest().GetGuild(guild.ID, false)
			if err != nil {
				dlog.Log.Error(err.Error())
				return err
			}
			dlog.Log.Debug("Found guild", "name", guild.Name)
			guildNode := db2.Guild{
				Id:   guild.ID.String(),
				Name: guild.Name,
			}
			write(db2.MergeN("g", guildNode))

			dlog.Log.Debug("Merged guild", "name", guild.Name)

			err = addMembers(write, err, event, guild, guildNode)
			if err != nil {
				dlog.Log.Error("Error in dbReadyHandler: ", err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		dlog.Log.Error("Error in dbReadyHandler: ", err)
		return
	}
}

func addMembers(write db2.Write, err error, event *events.Ready, guild *discord.RestGuild, guildNode db2.Guild) error {
	members, err := event.Client().Rest().GetMembers(guild.ID, 1000, snowflake.MustParse("0"))
	if err != nil {
		dlog.Log.Error("Failed to get members", "guild", guild.ID, "error", err)
		return err
	}
	for _, member := range members {
		dlog.Log.Debug("Found member", "name", member.User.Username, "id", member.User.ID)

		var avatar string
		if member.User.AvatarURL() == nil {
			avatar = member.User.DefaultAvatarURL()
		} else {
			avatar = *member.User.AvatarURL()
		}
		memberNode := db2.Member{
			Id:        member.User.ID.String(),
			Name:      member.EffectiveName(),
			AvatarUrl: avatar,
		}
		write(db2.MergeN("mb", memberNode))

		dlog.Log.Debug("Merged members", "name", member.EffectiveName())

		write(db2.MatchN("g", guildNode),
			db2.MatchN("m", memberNode),
			db2.Merge("(g)-[:HAS]->(m)-[:MEMBER_OF]->(g)"),
		)
	}
	return nil
}
