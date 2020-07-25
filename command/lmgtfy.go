package command

import (
	"strings"

	"github.com/PulseDevelopmentGroup/0x626f74/multiplexer"
	"github.com/patrickmn/go-cache"
)

// LMGTFY is a command
type LMGTFY struct {
	Command  string
	HelpText string

	RateLimitMax int
	RateLimitDB  *cache.Cache
}

var query = "https://lmgtfy.com/?q=%s&iie=1"

// Init is called by the multiplexer before the bot starts to initialize any
// variables the command needs.
func (c LMGTFY) Init(m *multiplexer.Mux) {
	// Nothing to init
}

// Handle is called by the multiplexer whenever a user triggers the command.
func (c LMGTFY) Handle(ctx *multiplexer.Context) {
	if len(ctx.Arguments) == 0 {
		ctx.ChannelSend("Maybe if you showed me what you were trying to google I'd be more helpful...")
		return
	}

	var sb strings.Builder
	for _, w := range ctx.Arguments {
		sb.WriteString(w + " ")
	}

	ctx.ChannelSendf(query, strings.ReplaceAll(strings.TrimSpace(sb.String()), " ", "+"))
}

// HandleHelp is called by whatever help command is in place when a user enters
// "!help [command name]". If the help command is not being handled, return
// false.
func (c LMGTFY) HandleHelp(ctx *multiplexer.Context) bool {
	ctx.ChannelSendf("It's simple. Does someone have a question they should've Googled? Just prefix their question with `!%s` and the bot will take care of the rest!", c.Command)
	return true
}

// Settings is called by the multiplexer on startup to process any settings
// associated with that command.
func (c LMGTFY) Settings() *multiplexer.CommandSettings {
	return &multiplexer.CommandSettings{
		Command:      c.Command,
		HelpText:     c.HelpText,
		RateLimitDB:  c.RateLimitDB,
		RateLimitMax: c.RateLimitMax,
	}
}
