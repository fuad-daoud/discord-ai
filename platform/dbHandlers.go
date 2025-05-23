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
	err := db.InTransaction(func(write db.Write) error {
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

	result, err := db.Query(cypher.MatchN("m", db.Message{Id: event.Message.ID.String()}), cypher.Return("m"))
	if err != nil {
		dlog.Log.Error("Error in dbMessageUpdateHandler: ", err)
		return
	}
	oldMessage, _ := cypher.ParseKey[db.Message]("m", result)

	err = db.InTransaction(func(write db.Write) error {
		set, err := cypher.Set("m", db.Message{
			Id:          oldMessage.Id,
			Text:        strings.ReplaceAll(event.Message.Content, "\"", "'"),
			UpdatedDate: event.Message.CreatedAt.String(),
			CreatedDate: oldMessage.CreatedDate,
		})
		if err != nil {
			return err
		}
		_, err = write(cypher.MatchN("m", oldMessage),
			set,
			cypher.Return("m"))
		if err != nil {
			return err
		}
		return nil
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

	message := db.Message{
		Id:          event.Message.ID.String(),
		Text:        strings.ReplaceAll(event.Message.Content, "\"", "'"),
		CreatedDate: event.Message.CreatedAt.String(),
	}

	member := db.Member{
		Id: event.Message.Author.ID.String(),
	}
	err := db.InTransaction(func(write db.Write) error {
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
	err := db.InTransaction(func(write db.Write) error {
		for _, guild := range event.Guilds {
			dlog.Log.Debug("Merging guild", "ID", guild.ID)
			guild, err := event.Client().Rest().GetGuild(guild.ID, false)
			if err != nil {
				dlog.Log.Error(err.Error())
				return err
			}
			dlog.Log.Debug("Found guild", "name", guild.Name)
			guildNode := db.Guild{
				Id:   guild.ID.String(),
				Name: guild.Name,
			}
			write(cypher.MergeN("g", guildNode))

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

func addMembers(write db.Write, err error, event *events.Ready, guild *discord.RestGuild, guildNode db.Guild) error {
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
	return nil
}
