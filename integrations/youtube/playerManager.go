package youtube

import "github.com/disgoorg/snowflake/v2"

var (
	players = make(map[snowflake.ID]Player)
)

func AddPlayer(player Player, guildId snowflake.ID) {
	players[guildId] = player
}

func GetPlayer(guildId snowflake.ID) Player {
	return players[guildId]
}
