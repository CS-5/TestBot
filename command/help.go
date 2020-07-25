package command

import (
	"strings"

	"github.com/PulseDevelopmentGroup/0x626f74/log"
	"github.com/PulseDevelopmentGroup/0x626f74/multiplexer"
	"github.com/bwmarrin/discordgo"
)

// Help is a bot command
type Help struct {
	Command  string
	HelpText string

	Logger *log.Logs
}

var (
	helpHandlers = make(map[string]func(ctx *multiplexer.Context) bool)
	helpFields   []*discordgo.MessageEmbedField
)

// Init is called by the multiplexer before the bot starts to initialize any
// variables the command needs.
func (c Help) Init(m *multiplexer.Mux) {
	i := 0
	for k, v := range m.Commands {
		msg := v.Settings().HelpText

		/* If there is no description, omit command from help */
		if len(msg) == 0 {
			continue
		}

		helpHandlers[strings.ToLower(k)] = v.HandleHelp

		helpFields = append(helpFields, &discordgo.MessageEmbedField{
			Name:   m.Prefix + k,
			Value:  msg,
			Inline: true,
		})
		i++
	}

	c.Logger.Command.WithField("command", c.Command).Infof(
		"Loaded help handlers and messages for %d commands", i,
	)
}

// Handle is called by the multiplexer whenever a user triggers the command.
func (c Help) Handle(ctx *multiplexer.Context) {
	if len(ctx.Arguments) == 0 {
		ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID,
			&discordgo.MessageEmbed{
				Title:       ":regional_indicator_h::regional_indicator_e::regional_indicator_l::regional_indicator_p:",
				Author:      &discordgo.MessageEmbedAuthor{},
				Color:       0xfdd329,
				Description: "Available commands:",
				Fields:      helpFields,
			})
		return
	}

	cmd := strings.ToLower(ctx.Arguments[0])
	handler, ok := helpHandlers[cmd]
	if !ok {
		ctx.ChannelSendf("Unable to find help handler for command: %s", cmd)
		return
	}

	if !handler(ctx) {
		ctx.ChannelSend(
			"Help handler not specified... guess you're out of luck",
		)
	}
}

// HandleHelp is called by whatever help command is in place when a user enters
// "!help [command name]". If the help command is not being handled, return
// false.
func (c Help) HandleHelp(ctx *multiplexer.Context) bool {
	ctx.ChannelSend("Are you sure _you_ don't need help?")
	return true
}

// Settings is called by the multiplexer on startup to process any settings
// associated with that command.
func (c Help) Settings() *multiplexer.CommandSettings {
	return &multiplexer.CommandSettings{
		Command:  c.Command,
		HelpText: c.HelpText,
	}
}
