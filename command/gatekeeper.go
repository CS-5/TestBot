package command

import (
	"fmt"
	"strings"

	"github.com/PulseDevelopmentGroup/0x626f74/log"
	"github.com/PulseDevelopmentGroup/0x626f74/multiplexer"
	"github.com/PulseDevelopmentGroup/0x626f74/util"
)

// Gatekeeper is a bot command
type Gatekeeper struct {
	Command  string
	HelpText string

	Logger *log.Logs
}

const (
	unknownCommand = "Unknown command. Usage: `!role [give|take] [Role Name]`"
	backendError   = "Backend error: `%q`"
)

// Init is called by the multiplexer before the bot starts to initialize any
// variables the command needs.
func (c Gatekeeper) Init(m *multiplexer.Mux) {
	// Nothing to init
}

// Handle is called by the multiplexer whenever a user triggers the command.
func (c Gatekeeper) Handle(ctx *multiplexer.Context) {
	guildID := ctx.Message.GuildID
	roles, err := ctx.Session.GuildRoles(guildID)
	if err != nil {
		c.Logger.CmdErr(ctx, err, "There was a problem getting the roles of the guild")
		return
	}

	// TODO: The way this works should probably be re-evaluated
	requestableRoles := make(map[string]string)
	printNames := []string{}
	for _, r := range roles {
		if strings.HasPrefix(r.Name, ":") {
			requestableRoles[strings.ToLower(r.Name[1:])] = r.ID
			printNames = append(printNames, r.Name[1:])
		}
	}

	/* If there are no arguments (Give/Take). Provide the user with options */
	if len(ctx.Arguments) == 0 {
		var msg strings.Builder
		msg.WriteString("Available roles are: ")

		for _, n := range printNames {
			msg.WriteString(fmt.Sprintf("\n- `%s`", n))
		}

		ctx.ChannelSend(msg.String())
		return
	}

	/* If there was an argument, was it give, take, or something invalid? */
	var give bool
	switch strings.ToLower(ctx.Arguments[0]) {
	case "give", "g", "gib":
		give = true
	case "take", "t", "tek":
		give = false
	default:
		ctx.ChannelSend(unknownCommand)
		return
	}

	/* If there was just one argument, inform the user */
	if len(ctx.Arguments) < 2 {
		ctx.ChannelSend(unknownCommand)
		return
	}

	/* Get the user ID and the user object */
	userID := ctx.Message.Author.ID
	member, err := ctx.Session.GuildMember(ctx.Message.GuildID, userID)
	if err != nil {
		c.Logger.CmdErr(ctx, err, "There was a problem getting the user id")
	}

	/* Check to see if the requested role is valid */
	req := strings.ToLower(ctx.Arguments[1])

	roleID, ok := requestableRoles[req]
	if !ok {
		ctx.ChannelSendf("Unable to find role `%s`", req)
		return
	}

	hasRole := false
	if util.ArrayContains(member.Roles, roleID, false) {
		hasRole = true
	}

	/* Give a role */
	if give {
		if hasRole {
			ctx.ChannelSendf(
				"You appear to already have that role, %s", member.Mention(),
			)
			return
		}
		ctx.Session.GuildMemberRoleAdd(guildID, userID, roleID)
		ctx.ChannelSendf(
			"You have been given role `%s`, %s", req, member.Mention(),
		)
		return
	}

	/* Take a role */
	if !hasRole {
		ctx.ChannelSendf(
			"You don't have that role... How do you expect me to take it, %s?",
			member.Mention(),
		)
		return
	}
	ctx.Session.GuildMemberRoleRemove(guildID, userID, roleID)
	ctx.ChannelSendf(
		"Taking role `%s` away, %s", req, member.Mention(),
	)
}

// HandleHelp is called by whatever help command is in place when a user enters
// "!help [command name]". If the help command is not being handled, return
// false.
func (c Gatekeeper) HandleHelp(ctx *multiplexer.Context) bool {
	ctx.ChannelSendf(
		"This server has a number of opt-in roles common interests.\n\n"+
			"To see a list of all available roles, use the `!%s` command. To "+
			"join an opt-in, use the `!%s give [opt-in name]` command. To "+
			"leave, use the `!%s take [opt-in name]` command.",
		c.Command, c.Command, c.Command,
	)
	return false
}

// Settings is called by the multiplexer on startup to process any settings
// associated with that command.
func (c Gatekeeper) Settings() *multiplexer.CommandSettings {
	return &multiplexer.CommandSettings{
		Command:  c.Command,
		HelpText: c.HelpText,
	}
}
