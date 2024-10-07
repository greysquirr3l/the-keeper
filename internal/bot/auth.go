// File: internal/bot/auth.go

package bot

import (
	"fmt"
	"net/url"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// IsAuthorized checks if a user has the required role to use a command
func IsAuthorized(s *discordgo.Session, guildID, userID string) bool {
	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		GetBot().GetLogger().WithFields(logrus.Fields{
			"guildID": guildID,
			"userID":  userID,
		}).WithError(err).Error("Error fetching guild member")
		return false
	}

	config := GetConfig()
	for _, roleID := range member.Roles {
		if roleID == config.Discord.RoleID {
			return true
		}
	}
	return false
}

// GetOAuth2URL generates the Discord OAuth2 authorization URL
func GetOAuth2URL() string {
	clientID := GetConfig().Discord.ClientID
	redirectURI := GetConfig().Discord.RedirectURL
	scopes := "identify guilds bot"

	return fmt.Sprintf(
		"https://discord.com/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s",
		clientID, url.QueryEscape(redirectURI), url.QueryEscape(scopes),
	)
}
