package youtube

import "github.com/disgoorg/snowflake/v2"

var (
	players = make(map[snowflake.ID]Player)
)

func GetPlayer(guildId snowflake.ID) Player {
	player, ok := players[guildId]
	if !ok {
		dbPlayer := DBPlayer{Id: guildId.String()}
		queue := GetQueue(dbPlayer)
		player = &DefaultPlayer{
			DBPlayer: dbPlayer,
			queue:    queue,
			GuildId:  guildId.String(),
			inst:     IDLE,
		}
		players[guildId] = player
		player.Save()
		queue.Load()
	}
	return player
}
