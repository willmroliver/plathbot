//go:build emoji && stats
// +build emoji,stats

package stats

import (
	emoji "github.com/willmroliver/plathbot/src/api_emoji"
	stats "github.com/willmroliver/plathbot/src/api_stats"
)

func init() {
	stats.Extensions.ExtendAPI(emoji.Title, emoji.Path, emoji.API().Select)
}
