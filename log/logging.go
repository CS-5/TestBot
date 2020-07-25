package log

import (
	"os"

	"github.com/PulseDevelopmentGroup/0x626f74/multiplexer"
	"github.com/sirupsen/logrus"
)

// Logs defines all the different loggers used within the bot
type Logs struct {
	Primary     *logrus.Logger
	Command     *logrus.Entry
	Multiplexer *logrus.Entry

	debug bool
}

// New creates a new Logs stuct. Accepts a boolean specifying whether
// debug mode is enabled.
func New(debug bool) *Logs {
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
		Primary:     primary,
		Command:     primary.WithField("type", "command"),
		Multiplexer: primary.WithField("type", "multiplexer"),
		debug:       debug,
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
func (l *Logs) CmdErr(ctx *multiplexer.Context, err error, msg string) {
	ctx.ChannelSendf(
		"That's broken.. maybe :at: Carson or Josiah?\n```Message: %s\nError: %s```",
		msg, err.Error(),
	)
	l.Command.Error(err)
}
