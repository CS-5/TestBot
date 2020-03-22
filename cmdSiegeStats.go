package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/CS-5/disgomux"
	"github.com/bwmarrin/discordgo"
)

type cSiegeStats struct {
	Command  string
	HelpText string
}

const (
	siegeUserQuery  = "https://r6tab.com/api/search.php?platform=uplay&search=%s"
	siegeIDQuery    = "https://r6tab.com/api/player.php?p_id=%s"
	siegeProfileURL = "https://r6.tracker.network/profile/pc/%s"
	siegeAvatar     = "https://ubisoft-avatars.akamaized.net/%s/default_256_256.png"
)

var (
	siegeOperators = map[string]string{
		"2:1":  "Smoke",
		"2:2":  "Castle",
		"2:3":  "Doc",
		"2:4":  "Glaz",
		"2:5":  "Blitz",
		"2:6":  "Buck",
		"2:7":  "Blackbeard",
		"2:8":  "Capitao",
		"2:9":  "Hibana",
		"2:A":  "Jackal",
		"2:B":  "Ying",
		"2:C":  "Ela",
		"2:D":  "Dokkaebi",
		"2:F":  "Maestro",
		"3:1":  "Mute",
		"3:2":  "Ash",
		"3:3":  "Rook",
		"3:4":  "Fuze",
		"3:5":  "IQ",
		"3:6":  "Frost",
		"3:7":  "Valkyrie",
		"3:8":  "Caveira",
		"3:9":  "Echo",
		"3:A":  "Mira",
		"3:B":  "Lesion",
		"3:C":  "Zofia",
		"3:D":  "Vigil",
		"3:E":  "Lion",
		"3:F":  "Alibi",
		"4:1":  "Sledge",
		"4:2":  "Pulse",
		"4:3":  "Twitch",
		"4:4":  "Kapkan",
		"4:5":  "Jager",
		"4:E":  "Finka",
		"5:1":  "Thatcher",
		"5:2":  "Thermite",
		"5:3":  "Montagne",
		"5:4":  "Tachanka",
		"5:5":  "Bandit",
		"1:5":  "GSG9 Recruit",
		"1:4":  "Spetsnaz Recruit",
		"1:3":  "GIGN Recruit",
		"1:2":  "FBI Recruit",
		"1:1":  "SAS Recruit",
		"2:11": "Nomad",
		"3:11": "Kaid",
		"3:10": "Clash",
		"2:10": "Maverick",
		"2:12": "Gridlock",
		"3:12": "Mozzie",
	}
)

type (
	siegeUser struct {
		Users []siegeUserResult `json:"results"`
	}

	siegeUserResult struct {
		ID        string `json:"p_id"`
		ProfileID string `json:"p_user"`
	}

	siegePlayer struct {
		Name      string `json:"p_name"`
		Found     bool   `json:"playerfound"`
		Level     int    `json:"p_level"`
		Updated   string `json:"updatedon"`
		MMR       int    `json:"p_currentmmr"`
		MaxMMR    int    `json:"p_maxmmr"`
		Rank      int    `json:"rank"`
		MaxRank   int    `json:"maxrank"`
		Data      []int  `json:"data"`
		FavAttack string `json:"favattacker"`
		FavDefend string `json:"facdefender"`
	}
)

func (s cSiegeStats) Init(m *disgomux.Mux) {
	// Nothing to init
}

func (s cSiegeStats) Handle(ctx *disgomux.Context) {
	ctx.Session.ChannelTyping(ctx.Message.ChannelID)

	if len(ctx.Arguments) == 0 || len(ctx.Arguments) > 1 {
		s.HandleHelp(ctx)
		return
	}

	user := new(siegeUser)
	userResp, err := http.Get(fmt.Sprintf(siegeUserQuery, ctx.Arguments[0]))
	if err != nil {
		cmdIssue(ctx, err, "Unable to get user data")
	}
	defer userResp.Body.Close()

	s.decode(ctx, userResp.Body, user)

	if len(user.Users) != 1 {
		ctx.ChannelSendf(
			"Could not find user with username: `%s`", ctx.Arguments[0],
		)
		return
	}

	player := new(siegePlayer)
	playerResp, err := http.Get(fmt.Sprintf(siegeIDQuery, user.Users[0].ID))
	if err != nil {
		cmdIssue(ctx, err, "Unable to get player data")
	}
	defer playerResp.Body.Close()

	s.decode(ctx, playerResp.Body, player)

	if !player.Found {
		ctx.ChannelSendf(
			"Could not find player with ID: `%s`", user.Users[1].ID,
		)
		return
	}

	ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID,
		&discordgo.MessageEmbed{
			Title:       player.Name,
			Description: "Current season stats.",
			URL:         fmt.Sprintf(siegeProfileURL, player.Name),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL:    fmt.Sprintf(siegeAvatar, user.Users[0].ProfileID),
				Width:  256,
				Height: 256,
			},
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   "Level",
					Value:  strconv.Itoa(player.Level),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name: "Overall K/D",
					Value: fmt.Sprintf(
						"%0.2f",
						float32(player.Data[1]+player.Data[6])/float32(player.Data[2]+player.Data[7]),
					),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name: "Overall W/L",
					Value: fmt.Sprintf(
						"%0.2f",
						float32(player.Data[3]+player.Data[8])/float32(player.Data[4]+player.Data[9]),
					),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Ranked MMR",
					Value:  strconv.Itoa(player.MMR),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name: "Ranked K/D",
					Value: fmt.Sprintf(
						"%0.2f",
						float32(player.Data[1])/float32(player.Data[2]),
					),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name: "Ranked W/L",
					Value: fmt.Sprintf(
						"%0.2f",
						float32(player.Data[3])/float32(player.Data[4]),
					),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Total Bullets",
					Value:  strconv.Itoa(player.Data[16]),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Favorite Attacker",
					Value:  siegeOperators[player.FavAttack],
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Favorite Defender",
					Value:  siegeOperators[player.FavDefend],
					Inline: true,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: strings.Replace(
					strings.Replace(player.Updated, "<u>", "", -1),
					"</u>", "", -1,
					// This is really not great, to be honest /shrug
				),
			},
			Color: 0x004080,
		},
	)

}

func (s cSiegeStats) decode(ctx *disgomux.Context, r io.Reader, v interface{}) {
	err := json.NewDecoder(r).Decode(&v)
	if err != nil {
		cmdIssue(ctx, err, "Unable to decode payload")
	}
}

func (s cSiegeStats) HandleHelp(ctx *disgomux.Context) bool {
	ctx.ChannelSend("`!siege [Username]` to get a list of a player's stats")
	return true
}

func (s cSiegeStats) Settings() *disgomux.CommandSettings {
	return &disgomux.CommandSettings{
		Command:  s.Command,
		HelpText: s.HelpText,
	}
}

func (s cSiegeStats) Permissions() *disgomux.CommandPermissions {
	return &disgomux.CommandPermissions{
		RoleIDs: config.permissions[s.Command],
	}
}
