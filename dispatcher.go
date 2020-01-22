package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/sirupsen/logrus"
)

type (
	multiplexer struct {
		Prefix, ErrorText string
		Commands          map[string]func(ctx *context)
		HelpText          map[string]string
		Logger            *logrus.Logger
		Debug             bool
	}

	context struct {
		Command   string
		Arguments []string
		Session   *discordgo.Session
		Message   *discordgo.MessageCreate
	}
)

// NewMux creates a new multiplexer.
func newMux(
	prefix, errorText string,
	logger *logrus.Logger,
	debug bool,
) (multiplexer, error) {
	if len(prefix) > 1 {
		return multiplexer{},
			fmt.Errorf("prefix %q longer than max length of 1", prefix)
	}

	et := errorText
	if len(errorText) == 0 {
		et = "Command not found."
	}

	if logger == nil {
		return multiplexer{},
			fmt.Errorf("logger invalid")
	}

	return multiplexer{
		Prefix:    prefix,
		ErrorText: et,
		Commands:  make(map[string]func(ctx *context)),
		HelpText:  make(map[string]string),
		Logger:    logger,
		Debug:     debug,
	}, nil
}

// Handle handles commands. Called directly from dg.AddHandler()
func (m *multiplexer) handle(
	session *discordgo.Session,
	message *discordgo.MessageCreate,
) {
	/* Ignore if the message being handled originated from the bot */
	if message.Author.ID == session.State.User.ID {
		return
	}

	/* Ignore if the message is not a regular message */
	if message.Type != discordgo.MessageTypeDefault || len(message.Content) == 0 {
		return
	}

	/* Split the message on the space */
	args := strings.Split(message.Content, " ")
	if args[0][:1] != m.Prefix {
		return
	}

	/* If debugging is enabled, log the message */
	if m.Debug {
		ch, _ := session.Channel(message.ChannelID)
		gu, _ := session.Guild(message.GuildID)
		m.Logger.WithFields(logrus.Fields{
			"messageGuild":   gu.Name,
			"messageChannel": ch.Name,
			"messageAuthor":  message.Author.Username,
			"messageContent": message.Content,
		}).Info("Message Recieved")
	}

	/* Retrieve the handler from the list (if present) */
	command := args[0][1:]
	handler, ok := m.Commands[command]
	if !ok {
		var (
			sb strings.Builder
			cl []string
		)

		for k := range m.Commands {
			cl = append(cl, k)
		}

		sb.WriteString("Command not found. Did you mean:")
		for _, v := range fuzzy.Find(command, cl) {
			sb.WriteString(fmt.Sprintf("\n- `%s`", v))
		}

		session.ChannelMessageSend(message.ChannelID, sb.String())
		return
	}

	/* Check if command was listed as requireing a special role */
	roles, ok := config.permissions[command]
	if !ok {
		/* If it doesn't, just handle it */
		go handler(&context{
			Command:   command,
			Arguments: args[1:],
			Session:   session,
			Message:   message,
		})
		return
	}

	/* If it does, check if the user has the correct permissions */
	member, err := session.GuildMember(message.GuildID, message.Author.ID)
	if err != nil {
		session.ChannelMessageSend(
			message.ChannelID,
			"This should never happen, if you see this: Run.",
		)
		return
	}

	/* Iterate through the roles required for the command and compare */
	for _, r := range roles {
		if arrayContains(member.Roles, r) {
			go handler(&context{
				Command:   command,
				Arguments: args[1:],
				Session:   session,
				Message:   message,
			})
			return
		}
	}

	/* Looks like you don't have permission */
	session.ChannelMessageSend(
		message.ChannelID,
		"You don't have permission to execute that command.",
	)
}

// Register adds a command to the bot
func (m *multiplexer) register(
	command, helpText string,
	handler func(ctx *context),
) error {
	if len(command) == 0 {
		return fmt.Errorf("Command '%v' too short", command)
	}

	m.Commands[command] = handler
	m.HelpText[command] = helpText
	return nil
}

// HandleHelp adds a !help command with auto-generated output. Must be called
// after all register commands
func (m *multiplexer) handleHelp(description string) {
	var fields []*discordgo.MessageEmbedField
	for k, v := range m.HelpText {
		// If there is no description for the command, omit it from help
		if len(v) == 0 {
			continue
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "!" + k,
			Value:  v,
			Inline: true,
		})
	}

	m.register("help", "Lists all commands and their functions.",
		func(ctx *context) {
			ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID,
				&discordgo.MessageEmbed{
					Title:       ":regional_indicator_h::regional_indicator_e::regional_indicator_l::regional_indicator_p:",
					Author:      &discordgo.MessageEmbedAuthor{},
					Color:       0xfdd329,
					Description: description,
					Fields:      fields,
				})
		},
	)
}

// ChannelSend enables easier sending of messages to the channel the command
// was recieved on.
func (c *context) channelSend(message string) {
	c.Session.ChannelMessageSend(c.Message.ChannelID, message)
}
