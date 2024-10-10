// File: internal/bot/auth.go

package bot

import (
	"fmt"
	"net/url"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// IsAuthorized checks if a user has the required role to use admin commands
func (b *Bot) IsAuthorized(s *discordgo.Session, guildID, userID string) bool {
	b.logger.WithFields(logrus.Fields{
		"user_id":        userID,
		"guild_id":       guildID,
		"config_role_id": b.Config.Discord.RoleID,
	}).Info("Checking authorization")

	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		b.logger.WithError(err).Error("Error fetching guild member")
		return false
	}

	b.logger.WithFields(logrus.Fields{
		"user_roles":    member.Roles,
		"required_role": b.Config.Discord.RoleID,
	}).Info("Comparing user roles")

	for _, roleID := range member.Roles {
		if roleID == b.Config.Discord.RoleID {
			b.logger.Info("User is authorized")
			return true
		}
	}

	b.logger.Info("User is not authorized")
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
