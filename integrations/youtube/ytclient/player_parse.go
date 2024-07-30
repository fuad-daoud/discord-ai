package ytclient

import (
	"context"
	"errors"
	"fmt"
	"regexp"
)

type playerConfig []byte

var basejsPattern = regexp.MustCompile(`(/s/player/\w+/player_ias.vflset/\w+/base.js)`)

func (c *Client) getPlayerConfig(ctx context.Context, videoID string) (playerConfig, error) {
	embedURL := fmt.Sprintf("https://youtube.com/embed/%s?hl=en", videoID)
	embedBody, err := c.httpGetBodyBytes(ctx, embedURL)
	if err != nil {
		return nil, err
	}

	// example: /s/player/f676c671/player_ias.vflset/en_US/base.js
	playerPath := string(basejsPattern.Find(embedBody))
	if playerPath == "" {
		return nil, errors.New("unable to find basejs URL in playerConfig")
	}

	config := c.playerCache.Get(playerPath)
	if config != nil {
		return config, nil
	}

	config, err = c.httpGetBodyBytes(ctx, "https://youtube.com"+playerPath)
	if err != nil {
		return nil, err
	}

	c.playerCache.Set(playerPath, config)
	return config, nil
}
