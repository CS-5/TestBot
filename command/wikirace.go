package command

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/PulseDevelopmentGroup/0x626f74/log"
	"github.com/PulseDevelopmentGroup/0x626f74/multiplexer"
	"github.com/bwmarrin/discordgo"
	"github.com/patrickmn/go-cache"
)

// Wiki is a command
type Wiki struct {
	Command  string
	HelpText string

	Logger *log.Logs

	RateLimitMax int
	RateLimitDB  *cache.Cache
}

type (
	articleInfo struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
	}

	wikiResult struct {
		BatchComplete string `json:"batchComplete"`
		Query         struct {
			Random []articleInfo `json:"random"`
		} `json:"query"`
	}
)

// Init is called by the multiplexer before the bot starts to initialize any
// variables the command needs.
func (c Wiki) Init(m *multiplexer.Mux) {
	// Nothing to init
}

// Handle is called by the multiplexer whenever a user triggers the command.
func (c Wiki) Handle(ctx *multiplexer.Context) {
	resp, err := http.Get("https://en.wikipedia.org/w/api.php?action=query&format=json&list=random&rnnamespace=0&rnlimit=2")
	if err != nil {
		c.Logger.CmdErr(ctx, err, "Unable to get random wikipedia page")
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Logger.CmdErr(ctx, err, "Unable to read page")
		return
	}

	var search wikiResult
	err = json.Unmarshal(body, &search)
	if err != nil {
		c.Logger.CmdErr(ctx, err, "Unable to unmarshal page")
		return
	}

	articles := search.Query.Random[:2]

	ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID,
		&discordgo.MessageEmbed{
			Title:       "Wikipedia Race",
			Author:      &discordgo.MessageEmbedAuthor{},
			Color:       0x0080ff,
			Description: "Start at the start and use only blue links in the article to get to the end page!",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name: "Start:vertical_traffic_light:",
					Value: fmt.Sprintf(
						"[%s](%s%d)",
						articles[0].Title,
						"https://en.wikipedia.org/?curid=",
						articles[0].ID,
					),
					Inline: false,
				},
				{
					Name: "End :checkered_flag:",
					Value: fmt.Sprintf(
						"[%s](%s%d)",
						articles[1].Title,
						"https://en.wikipedia.org/?curid=",
						articles[1].ID,
					),
					Inline: false,
				},
			},
		})
}

// HandleHelp is called by whatever help command is in place when a user enters
// "!help [command name]". If the help command is not being handled, return
// false.
func (c Wiki) HandleHelp(ctx *multiplexer.Context) bool {
	var sb strings.Builder
	sb.WriteString(
		"Use `!wikirace` to start a new race! The rules are simple:\n",
	)
	sb.WriteString("1. Only blue links _within_ the article are allowed\n")
	sb.WriteString("2. You cannot use the back button or the search function\n")
	sb.WriteString("3. Whoever gets to end article in the fewest clicks wins\n")

	ctx.ChannelSend(sb.String())
	return true
}

// Settings is called by the multiplexer on startup to process any settings
// associated with that command.
func (c Wiki) Settings() *multiplexer.CommandSettings {
	return &multiplexer.CommandSettings{
		Command:      c.Command,
		HelpText:     c.HelpText,
		RateLimitMax: c.RateLimitMax,
		RateLimitDB:  c.RateLimitDB,
	}
}
