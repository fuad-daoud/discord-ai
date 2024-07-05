package platform

import (
	"context"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jonas747/ogg"
	"io"
)

type AudioProvider struct {
	decoder *ogg.PacketDecoder
	Source  io.Reader
}

func (p *AudioProvider) ProvideOpusFrame() ([]byte, error) {
	if p.decoder == nil {
		p.decoder = ogg.NewPacketDecoder(ogg.NewDecoder(p.Source))
	}

	data, _, err := p.decoder.Decode()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (p *AudioProvider) Close() {
	if c, ok := p.Source.(io.Closer); ok {
		_ = c.Close()
	}
	//_ = p.Source.Close()
}

func (p *AudioProvider) Wait() error {
	return nil
}
func deafen(guildId *snowflake.ID, channelId *snowflake.ID) func() error {
	return func() error {
		return Client().UpdateVoiceState(context.Background(), *guildId, channelId, false, true)
	}
}
func unDeafen(guildId *snowflake.ID, channelId *snowflake.ID) func() error {
	return func() error {
		return Client().UpdateVoiceState(context.Background(), *guildId, channelId, false, true)
	}
}

type UpdateVoiceState func(ctx context.Context, guildID snowflake.ID, channelID *snowflake.ID, selfMute bool, selfDeaf bool) error
