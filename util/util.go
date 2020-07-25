package util

import (
	"net/url"
	"os"
	"regexp"
	"strings"
)

/* === Helpers === */

// InitFile opens a file at the specified path. If that file does not exist,
// it creates a new one.
func InitFile(path string) (*os.File, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		file, err := os.Create(path)

		if err != nil {
			return &os.File{}, err
		}
		return file, err
	}
	return file, err
}

// ArrayContains checks a string array for a given string.
func ArrayContains(array []string, value string, ignoreCase bool) bool {
	for _, e := range array {
		if ignoreCase {
			e = strings.ToLower(e)
		}

		if e == value {
			return true
		}
	}
	return false
}

// IsURL checks the provided string to see if it's a valid URL.
func IsURL(test string) bool {
	_, err := url.ParseRequestURI(test)
	if err != nil {
		return false
	}

	u, err := url.Parse(test)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

func GetMsgURL(guildID, channelID, messageID string) string {
	return "https://discordapp.com/channels/" +
		guildID + "/" + channelID + "/" + messageID
}

var (
	idRE      = regexp.MustCompile(`^\d{18}`)
	botRE     = regexp.MustCompile(`<@&\d{18}>`)
	userRE    = regexp.MustCompile(`<@!*\d{18}>`)
	channelRE = regexp.MustCompile(`<#\d{18}>`)

	idExtractRE = regexp.MustCompile(`\d{18}`)
)

// IsID checks if the supplied string is a Discord ID
func IsID(test string) bool {
	return idRE.MatchString(test)
}

// GetID returns the ID (if there is one) from the supplied string
func GetID(input string) string {
	return idExtractRE.FindString(input)
}

// IsBot checks if the supplied string is mentioning a bot
func IsBot(test string) bool {
	return botRE.MatchString(test)
}

// IsUser checks if the supplied string is mentioning a user
func IsUser(test string) bool {
	return userRE.MatchString(test)
}

// IsChannel checks if the supplied string is mentioning a channel
func IsChannel(test string) bool {
	return channelRE.MatchString(test)
}
