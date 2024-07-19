package platform

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/fuad-daoud/discord-ai/db"
	"github.com/fuad-daoud/discord-ai/db/cypher"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"strings"
)

func dbThreadCreateHandler(event *events.ThreadCreate) {
	dlog.Log.Debug("Starting db ThreadCreateHandler")
	if event.Thread.OwnerID != event.Client().ApplicationID() {
		return
	}
	db.InTransaction(func(write db.Write) {
		write(cypher.MergeN("t", db.Thread{
			TextChannel: db.TextChannel{
				Id:          event.ThreadID.String(),
				Name:        event.Thread.Name(),
				CreatedDate: event.Thread.CreatedAt().String(),
			},
		}))

		write(cypher.MatchN("t", db.Thread{TextChannel: db.TextChannel{Id: event.ThreadID.String()}}),
			cypher.MatchN("m", db.Message{Id: event.ThreadID.String()}),
			cypher.Merge("(m)-[:STARTED]->(t)-[:CONTAINS]->(m)"))

		write(cypher.MatchN("mb", db.Member{Id: event.Thread.OwnerID.String()}),
			cypher.MatchN("t", db.Thread{
				TextChannel: db.TextChannel{
					Id: event.Thread.ID().String(),
				},
			}),
			cypher.Merge("(mb)-[:CREATED]->(t)"),
		)
	})
	dlog.Log.Debug("Finished DB ThreadCreateHandler")
}

func dbMessageUpdateHandler(event *events.GuildMessageUpdate) {
	dlog.Log.Debug("Starting DB MessageUpdateHandler")
	if event.Message.Flags == discord.MessageFlagHasThread && event.OldMessage.ChannelID == event.ChannelID {
		return
	}

	result := db.Query(cypher.MatchN("m", db.Message{Id: event.Message.ID.String()}), cypher.Return("m"))

	oldMessage, _ := cypher.ParseKey[db.Message]("m", result)

	db.InTransaction(func(write db.Write) {
		write(cypher.MatchN("m", oldMessage),
			cypher.Set("m", db.Message{
				Id:          oldMessage.Id,
				Text:        strings.ReplaceAll(event.Message.Content, "\"", "'"),
				UpdatedDate: event.Message.CreatedAt.String(),
				CreatedDate: oldMessage.CreatedDate,
			}),
			cypher.Return("m"))
	})
	dlog.Log.Debug("Finished DB MessageUpdateHandler")
}

func dbMessageCreateHandler(event *events.GuildMessageCreate) {
	if event.Message.Type == discord.MessageTypeThreadStarterMessage {
		return
	}
	dlog.Log.Debug("Starting DB messageCreateHandler")
	restClient := event.Client().Rest()
	channel, _ := restClient.GetChannel(event.ChannelID)

	message := db.Message{
		Id:          event.Message.ID.String(),
		Text:        strings.ReplaceAll(event.Message.Content, "\"", "'"),
		CreatedDate: event.Message.CreatedAt.String(),
	}

	member := db.Member{
		Id: event.Message.Author.ID.String(),
	}
	db.InTransaction(func(write db.Write) {
		if channel.Type() == discord.ChannelTypeGuildPublicThread {
			threadMessageCreateHandler(write, channel, member, message)
		} else {
			createMessageHandler(write, channel, event.GuildID.String(), member, message)
		}
	})
	dlog.Log.Debug("Finished DB messageCreateHandler")
}

func createMessageHandler(write db.Write, channel discord.Channel, guildId string, member db.Member, message db.Message) {
	textChannel := db.TextChannel{
		Id:   channel.ID().String(),
		Name: channel.Name(),
	}
	dbGuild := db.Guild{Id: guildId}
	write(cypher.MatchN("g", dbGuild),
		cypher.MatchN("mb", member),
		cypher.MergeN("c", textChannel),
		cypher.CreateN("m", message),
		cypher.Merge("(g)-[:HAS]->(c)"),
		cypher.Merge("(c)-[:CONTAINS]->(m)-[:AUTHOR]->(mb)-[:CREATED]->(m)"))
}

func threadMessageCreateHandler(write db.Write, channel discord.Channel, member db.Member, message db.Message) {
	textChannel := db.TextChannel{
		Id: channel.(discord.GuildThread).ParentID().String(),
	}
	dbThread := db.Thread{
		TextChannel: db.TextChannel{
			Id:   channel.(discord.GuildThread).ID().String(),
			Name: channel.(discord.GuildThread).Name(),
		},
	}
	write(cypher.MatchN("c", textChannel),
		cypher.MatchN("mb", member),
		cypher.MergeN("t", dbThread),
		cypher.CreateN("m", message),
		cypher.Merge("(c)-[:CHILD]->(t)-[:CONTAINS]->(m)-[:AUTHOR]->(mb)-[:CREATED]->(m)"),
		cypher.Merge("(t)-[:PARENT]->(c)"))
}

func dbReadyHandler(event *events.Ready) {
	db.InTransaction(func(write db.Write) {
		for _, guild := range event.Guilds {
			dlog.Log.Debug("Merging guild", "ID", guild.ID)
			guild, err := event.Client().Rest().GetGuild(guild.ID, false)
			if err != nil {
				dlog.Log.Error(err.Error())
				panic(err)
			}
			dlog.Log.Debug("Found guild", "name", guild.Name)
			guildNode := db.Guild{
				Id:   guild.ID.String(),
				Name: guild.Name,
			}
			write(cypher.MergeN("g", guildNode))

			dlog.Log.Debug("Merged guild", "name", guild.Name)

			addMembers(write, err, event, guild, guildNode)
		}
	})
}

func addMembers(write db.Write, err error, event *events.Ready, guild *discord.RestGuild, guildNode db.Guild) {
	members, err := event.Client().Rest().GetMembers(guild.ID, 1000, snowflake.MustParse("0"))
	if err != nil {
		dlog.Log.Error("Failed to get members", "guild", guild.ID, "error", err)
		panic(err)
	}
	for _, member := range members {
		dlog.Log.Debug("Found member", "name", member.User.Username, "id", member.User.ID)

		var avatar string
		if member.User.AvatarURL() == nil {
			avatar = member.User.DefaultAvatarURL()
		} else {
			avatar = *member.User.AvatarURL()
		}
		memberNode := db.Member{
			Id:        member.User.ID.String(),
			Name:      member.EffectiveName(),
			AvatarUrl: avatar,
		}
		write(cypher.MergeN("mb", memberNode))

		dlog.Log.Debug("Merged members", "name", member.EffectiveName())

		write(cypher.MatchN("g", guildNode),
			cypher.MatchN("m", memberNode),
			cypher.Merge("(g)-[:HAS]->(m)-[:MEMBER_OF]->(g)"),
		)
	}
}
