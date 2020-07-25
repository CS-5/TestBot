package log

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/PulseDevelopmentGroup/0x626f74/multiplexer"
	"github.com/PulseDevelopmentGroup/0x626f74/util"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// Logs defines all the different loggers used within the bot
type Logs struct {
	Primary     *logrus.Logger
	Command     *logrus.Entry
	Multiplexer *logrus.Entry

	debug        bool
	errorChannel string
}

// New creates a new Logs stuct. Accepts a boolean specifying whether
// debug mode is enabled.
func New(debug bool, errorChannel string) *Logs {
	logrus.SetOutput(os.Stdout)
	primary := logrus.New()

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
		primary.SetFormatter(&logrus.TextFormatter{
			ForceColors: true,
		})
	} else {
		logrus.SetLevel(logrus.InfoLevel)
		primary.SetFormatter(&logrus.JSONFormatter{})

	}

	return &Logs{
		Primary:      primary,
		Command:      primary.WithField("type", "command"),
		Multiplexer:  primary.WithField("type", "multiplexer"),
		debug:        debug,
		errorChannel: errorChannel,
	}
}

// MuxMiddleware is the middleware function attached to MuxLog. Accepts the context
// from disgomux.
func (l *Logs) MuxMiddleware(ctx *multiplexer.Context) {
	if l.debug {
		// Ignoring errors here since they're effectivly meaningless
		ch, _ := ctx.Session.Channel(ctx.Message.ChannelID)
		gu, _ := ctx.Session.Guild(ctx.Message.GuildID)

		l.Multiplexer.WithFields(logrus.Fields{
			"messageGuild":   gu.Name,
			"messageChannel": ch.Name,
			"messageAuthor":  ctx.Message.Author.Username,
			"messageContent": ctx.Message.Content,
		}).Info("Message Recieved")
	}
}

// CmdErr is used for handling errors within commands which should be reported
// to the user. Takes a multiplexer context, error message, and user-readable
// message which are sent to the channel where the command was executed.
func (l *Logs) CmdErr(ctx *multiplexer.Context, errMsg error, msg string) {
	// Inform the user of the issue (using a basic message string)
	ctx.ChannelSendf("The bot seems to have encountered an issue: `%s`", msg)

	// Inform the admins of the issue
	msgTime, err := ctx.Message.Timestamp.Parse()
	if err != nil {
		msgTime = time.Time{}
	}

	msgChannel := "unknown"
	channel, err := ctx.Session.Channel(ctx.Message.ChannelID)
	if err == nil {
		msgChannel = channel.Name
	}

	ctx.Session.ChannelMessageSendEmbed(l.errorChannel, &discordgo.MessageEmbed{
		Color: 0xff0000,
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: ctx.Message.Author.AvatarURL(""),
			Name:    ctx.Message.Author.Username,
		},
		Title: fmt.Sprintf("🚧 Error with command `%s%s`", ctx.Prefix, ctx.Command),
		URL: util.GetMsgURL(
			ctx.Message.GuildID, ctx.Message.ChannelID, ctx.Message.ID,
		),

		Timestamp: msgTime.Format("2006-01-02T15:04:05.000Z"),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🚶 User",
				Value:  ctx.Message.Author.Username,
				Inline: true,
			},
			{
				Name:   "#️⃣ Channel",
				Value:  fmt.Sprintf("#%s", msgChannel),
				Inline: true,
			},
			{
				Name:   "🕹️ Command",
				Value:  ctx.Prefix + ctx.Command,
				Inline: true,
			},
			{
				Name:  "✉️ Command Message",
				Value: msg,
			},
			{
				Name:  "⚠️ Error Message",
				Value: errMsg.Error(),
			},
			{
				Name: "🖊️ Command Text",
				Value: ctx.Prefix + ctx.Command +
					" " + strings.Join(ctx.Arguments[:], " "),
			},
		},
	})

	l.Command.Error(errMsg)
}
