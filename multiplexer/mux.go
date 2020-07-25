package multiplexer

import (
	"fmt"
	"strings"

	"github.com/PulseDevelopmentGroup/0x626f74/util"
	"github.com/bwmarrin/discordgo"
	"github.com/patrickmn/go-cache"
	"github.com/sahilm/fuzzy"
)

type (
	// Mux is the multiplexer object. Initialized with New().
	Mux struct {
		Prefix         string
		Commands       map[string]Command
		SimpleCommands map[string]SimpleCommand
		Middleware     []Middleware
		options        *Options
		fuzzyMatch     bool
		commandNames   []string
		errorTexts     *ErrorTexts
		permissions    map[string]*CommandPermissions
	}

	// Command specifies the functions for a multiplexed command
	Command interface {
		Init(m *Mux)
		Handle(ctx *Context)
		HandleHelp(ctx *Context) bool
		Settings() *CommandSettings
	}

	// CommandPermissions holds the specific ID arrays for a given command in whitelist
	// format. UserID takes priority over all other permissions. RoleID takes
	// priority over ChanID.
	CommandPermissions struct {
		UserIDs []string
		RoleIDs []string
		ChanIDs []string
	}

	// CommandSettings contain command-specific settings the multiplexer should
	// know.
	CommandSettings struct {
		Command, HelpText string

		RateLimitMax int
		RateLimitDB  *cache.Cache
	}

	// SimpleCommand contains the content and helptext of a logic-less command.
	// Simple commands have no support for permissions.
	SimpleCommand struct {
		Command, Content, HelpText string
	}

	// ErrorTexts holds strings used when an error occurs
	ErrorTexts struct {
		CommandNotFound, NoPermissions, RateLimited string
	}

	// Context is the contexual values supplied to middlewares and handlers
	Context struct {
		Prefix, Command string
		Arguments       []string
		Session         *discordgo.Session
		Message         *discordgo.MessageCreate
	}

	// Middleware specifies a special middleware function that is called anytime
	// handle() is called from DiscordGo
	Middleware func(*Context)

	// Options is a set of config options to use when handling a message. All
	// properties true by default.
	Options struct {
		IgnoreBots       bool
		IgnoreDMs        bool
		IgnoreEmpty      bool
		IgnoreNonDefault bool
	}
)

// New initlaizes a new Mux object
func New(prefix string) (*Mux, error) {
	if len(prefix) > 1 {
		return &Mux{}, fmt.Errorf("prefix %s greater than 1 character", prefix)
	}

	return &Mux{
		Prefix:         prefix,
		Commands:       make(map[string]Command),
		SimpleCommands: make(map[string]SimpleCommand),
		Middleware:     []Middleware{},
		errorTexts: &ErrorTexts{
			CommandNotFound: "Command not found.",
			NoPermissions:   "You do not have permission to use that command.",
		},
		options:     &Options{true, true, true, true},
		permissions: make(map[string]*CommandPermissions),
		fuzzyMatch:  false,
	}, nil
}

// SetOptions allows configuration of the multiplexer. Must be called before
// Initialize()
func (m *Mux) SetOptions(opt *Options) {
	m.options = opt
}

// SetPermissions allows defining permissions for each command. Must be called
// before Initialize()
func (m *Mux) SetPermissions(perms map[string]*CommandPermissions) {
	m.permissions = perms
}

// UseMiddleware adds a middleware to the multiplexer. Middlewares are called
// before a command is handled.
func (m *Mux) UseMiddleware(mw Middleware) {
	m.Middleware = append(m.Middleware, mw)
}

// SetErrors sets the error texts for the multiplexer using the supplied struct
func (m *Mux) SetErrors(errorTexts *ErrorTexts) {
	m.errorTexts = errorTexts
}

// Register registers one or more commands to the multiplexer
func (m *Mux) Register(commands ...Command) {
	for _, c := range commands {
		cString := c.Settings().Command
		if len(cString) != 0 {
			m.Commands[cString] = c
		}
	}
}

// RegisterSimple registers one or more simple commands to the multiplexer
func (m *Mux) RegisterSimple(simpleCommands ...SimpleCommand) {
	for _, c := range simpleCommands {
		cString := c.Command
		if len(cString) != 0 {
			m.SimpleCommands[cString] = c
		}
	}
}

func (m *Mux) ClearSimple() {
	m.SimpleCommands = make(map[string]SimpleCommand)
}

// UseFuzzy both enables and builds a list of commands to fuzzy match
// against. May result in a small performance hit
func (m *Mux) UseFuzzy() {
	m.fuzzyMatch = true

	for k := range m.Commands {
		m.commandNames = append(m.commandNames, k)
	}
}

// Initialize calls the init functions of all registered commands to do any
// preloading or setup before commands are to be handled. Must be called before
// Mux.Handle() and after Mux.Register()
func (m *Mux) Initialize() {
	/* If no commands are loaded, and none are specified, return */
	if len(m.Commands) == 0 {
		return
	}

	for _, c := range m.Commands {
		c.Init(m)
	}
}

// Handle is passed to DiscordGo to handle actions
func (m *Mux) Handle(
	session *discordgo.Session,
	message *discordgo.MessageCreate,
) {
	/* Ignore if the message being handled originated from the bot */
	if message.Author.ID == session.State.User.ID {
		return
	}

	/* Ignore if the message has no content */
	if m.options.IgnoreEmpty && len(message.Content) == 0 {
		return
	}

	/* Ignore if the message is not default */
	if m.options.IgnoreNonDefault &&
		message.Type != discordgo.MessageTypeDefault {
		return
	}

	/* Ignore if the message originated from a bot */
	if m.options.IgnoreBots && message.Author.Bot {
		return
	}

	/* Ignore if the message is in a DM */
	if m.options.IgnoreDMs && message.GuildID == "" {
		return
	}

	/* Ignore if the message doesn't have the prefix */
	if !strings.HasPrefix(message.Content, m.Prefix) {
		return
	}

	/* Split the message on the space */
	args := strings.Split(message.Content, " ")
	command := strings.ToLower(args[0][1:])

	simple, ok := m.SimpleCommands[command]
	if ok {
		session.ChannelMessageSend(message.ChannelID, simple.Content)
		return
	}

	handler, ok := m.Commands[command]
	/* If command does not exist, attempt to fuzzy match it */
	if !ok {
		if m.fuzzyMatch {
			var sb strings.Builder

			for _, fzy := range fuzzy.Find(command, m.commandNames) {
				sb.WriteString("- `!" + fzy.Str + "`\n")
			}

			if sb.Len() != 0 {
				session.ChannelMessageSend(
					message.ChannelID,
					fmt.Sprintf(
						"Command not found. Did you mean: \n%s", sb.String(),
					),
				)
				return
			}

		}

		session.ChannelMessageSend(
			message.ChannelID,
			m.errorTexts.CommandNotFound,
		)

		return
	}

	/* Form context */
	settings := handler.Settings()
	ctx := &Context{
		Prefix:    m.Prefix,
		Command:   command,
		Arguments: args[1:],
		Session:   session,
		Message:   message,
	}

	if !settings.checkLimit(message.Author.ID) {
		ctx.ChannelSend(m.errorTexts.RateLimited)
		return
	}

	// TODO: Move away from middlewares and more closely integrate logging
	/* Call middlewares */
	if len(m.Middleware) > 0 {
		for _, mw := range m.Middleware {
			go mw(ctx)
		}
	}

	/* If permissions have been specified, check them */
	p, ok := m.permissions[command]
	if ok {
		member, err := session.GuildMember(message.GuildID, message.Author.ID)
		if err != nil {
			ctx.ChannelSend("There was a weird issue.")
			return
		}

		/* Check the permissions struct against the context */
		if !CheckPermissions(
			p, member.User.ID, member.Roles, message.ChannelID,
		) {
			/* The user doesn't have the correct permissions */
			ctx.ChannelSend(m.errorTexts.NoPermissions)
			return
		}
	}

	/* User has permissions or it doesnt require them? Run it */
	go handler.Handle(ctx)
}

/* === Helper Functions === */

// checkLimit checks the supplied command settings' rate limiter to see if
// the user is allowed to run the command.
func (cs *CommandSettings) checkLimit(id string) bool {
	/* No rate limiter set? Ignore */
	if cs.RateLimitDB == nil {
		return true
	}

	/* Increment limiter. Does not increase elements which do not exist */
	cs.RateLimitDB.IncrementInt(id, 1)

	uses, found := cs.RateLimitDB.Get(id)
	/* If not found, initialize */
	if !found {
		cs.RateLimitDB.SetDefault(id, 1)
		return true
	}

	/* Check uses (add 1 since we've incremented the value) */
	if uses.(int) < cs.RateLimitMax+1 {
		return true
	}

	return false
}

// ChannelSend is a helper function for easily sending a message to the current
// channel.
func (ctx *Context) ChannelSend(message string) (*discordgo.Message, error) {
	return ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, message)
}

// ChannelSendf is a helper function like ChannelSend for sending a formatted
// message to the current channel.
func (ctx *Context) ChannelSendf(
	format string,
	a ...interface{},
) (*discordgo.Message, error) {
	return ctx.Session.ChannelMessageSend(
		ctx.Message.ChannelID, fmt.Sprintf(format, a...),
	)
}

// CheckPermissions takes the user, role(s), and channel IDs and checks them
// against the supplied permissions struct.
func CheckPermissions(
	perms *CommandPermissions,
	userID string, roleIDs []string, chanID string,
) bool {
	// TODO: Fix this gnarly logic
	if len(perms.UserIDs) == 0 && len(perms.RoleIDs) == 0 && len(perms.ChanIDs) == 0 {
		return true
	}

	if util.ArrayContains(perms.UserIDs, userID, true) {
		return true
	}

	for _, id := range roleIDs {
		if util.ArrayContains(perms.RoleIDs, id, true) {
			return true
		}
	}

	return util.ArrayContains(perms.ChanIDs, chanID, true)
}
