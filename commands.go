package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/CS-5/disgomux"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

/* === Command Template === */
/*

type cName struct {
	Command  string
	HelpText string
}

func (n cName) Init(m *disgomux.Mux) {
	// Initialization called before bot starts
}

func (n cName) Handle(ctx *disgomux.Context) {
	// Handle called when command is recieved
}

func (n cName) HandleHelp(ctx *disgomux.Context) {
	// HandleHelp is called by the help command (if there is one) to output
	// specific help info
}

func (n cName) Settings() *disgomux.CommandSettings {
	// Settings are called as-needed by the multiplexer to get configuration
	// information

	return &disgomux.CommandSettings{
		Command:  n.Command,
		HelpText: n.HelpText,
	}
}

func (d cName) Permissions() *disgomux.CommandPermissions {
	// Permissions are called as-needed by the multiplexer to get permissions
	// needed for the command

	// For no permissions:
	return &disgomux.CommandPermissions{}
}
*/

/* === Start Debug Command === */

type cDebug struct {
	Command  string
	HelpText string
}

func (d cDebug) Init(m *disgomux.Mux) {
	// Nothing to init
}

func (d cDebug) Handle(ctx *disgomux.Context) {
	if len(ctx.Arguments) == 0 {
		d.HandleHelp(ctx)
		return
	}

	switch ctx.Arguments[0] {
	case "config":
		var sb strings.Builder
		sb.WriteString(
			fmt.Sprintf("`Simple Commands: %+v`\n\n", config.simpleCommands))
		sb.WriteString(
			fmt.Sprintf("`Permissions: %+v`", config.permissions))
		ctx.ChannelSend(sb.String())
	case "args":
		ctx.ChannelSend(fmt.Sprintf("%+v", ctx.Arguments))
	default:
		ctx.ChannelSend("Debug")
	}
}

func (d cDebug) HandleHelp(ctx *disgomux.Context) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"`%s%s config`: Returns the contents of the JSON config file.\n",
		ctx.Prefix, d.Command,
	))
	sb.WriteString(fmt.Sprintf(
		"`%s%s args`: Returns the supplied arguments.",
		ctx.Prefix, d.Command,
	))
	ctx.ChannelSend(sb.String())
}

func (d cDebug) Settings() *disgomux.CommandSettings {
	return &disgomux.CommandSettings{
		Command:  d.Command,
		HelpText: d.HelpText,
	}
}

func (d cDebug) Permissions() *disgomux.CommandPermissions {
	return &disgomux.CommandPermissions{
		RoleIDs: config.permissions[d.Command],
	}
}

/* === End Debug Command === */

/* === Start WikiRace Command === */

type (
	cWiki struct {
		Command  string
		HelpText string
	}

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

const issueText = "Hmm... I seem to have run into an issue... Try again later?"

func (w cWiki) Init(m *disgomux.Mux) {
	// Nothing to init
}

func (w cWiki) Handle(ctx *disgomux.Context) {
	/* TODO: Maybe float these erros up to the handler? */
	resp, err := http.Get("https://en.wikipedia.org/w/api.php?action=query&format=json&list=random&rnnamespace=0&rnlimit=2")
	if err != nil {
		cLog.WithFields(logrus.Fields{
			"error":   err,
			"command": ctx.Command,
		}).Errorf("Unable to get random wikipedia page")

		ctx.ChannelSend(issueText)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		cLog.WithFields(logrus.Fields{
			"error":   err,
			"command": ctx.Command,
		}).Error("Unable to read page")

		ctx.ChannelSend(issueText)
		return
	}

	var search wikiResult
	err = json.Unmarshal(body, &search)
	if err != nil {
		cLog.WithFields(logrus.Fields{
			"error":   err,
			"command": ctx.Command,
		}).Error("Unable to unmarshal page")

		ctx.ChannelSend(issueText)
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
				&discordgo.MessageEmbedField{
					Name: "Start:vertical_traffic_light:",
					Value: fmt.Sprintf(
						"[%s](%s%d)",
						articles[0].Title,
						"https://en.wikipedia.org/?curid=",
						articles[0].ID,
					),
					Inline: false,
				},
				&discordgo.MessageEmbedField{
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

func (w cWiki) HandleHelp(ctx *disgomux.Context) {
	var sb strings.Builder
	sb.WriteString(
		"Use `!wikirace` to start a new race! The rules are simple:\n",
	)
	sb.WriteString("1. Only blue links _within_ the article are allowed\n")
	sb.WriteString("2. You cannot use the back button or the search function\n")
	sb.WriteString("3. Whoever gets to end article in the fewest clicks wins\n")

	ctx.ChannelSend(sb.String())
}

func (w cWiki) Settings() *disgomux.CommandSettings {
	return &disgomux.CommandSettings{
		Command:  w.Command,
		HelpText: w.HelpText,
	}
}

func (w cWiki) Permissions() *disgomux.CommandPermissions {
	return &disgomux.CommandPermissions{
		RoleIDs: config.permissions[w.Command],
	}
}

/* === End WikiRace Command */

/* === Start Gatekeeper Command === */

type cGate struct {
	Command  string
	HelpText string
}

const (
	unknownCommand = "Unknown command. Usage: `!role [give|take] [Role Name]`"
	backendError   = "Backend error: `%q`"
)

func (g cGate) Init(m *disgomux.Mux) {
	// Nothing to init
}

func (g cGate) Handle(ctx *disgomux.Context) {
	guildID := ctx.Message.GuildID
	roles, err := ctx.Session.GuildRoles(guildID)
	if err != nil {
		cLog.WithFields(logrus.Fields{
			"error":   err,
			"command": ctx.Command,
		}).Errorf(
			"Problem getting roles for guild `%v`", guildID)
		return
	}

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
	case "give", "g":
		give = true
	case "take", "t":
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
		ctx.ChannelSend(fmt.Sprintf(backendError, err))
	}

	/* Check to see if the requested role is valid */
	req := strings.ToLower(ctx.Arguments[1])

	roleID, ok := requestableRoles[req]
	if !ok {
		ctx.ChannelSend(
			fmt.Sprintf("Unable to find role `%s`", req),
		)
		return
	}

	hasRole := false
	if arrayContains(member.Roles, roleID, false) {
		hasRole = true
	}

	/* Give a role */
	if give {
		if hasRole {
			ctx.ChannelSend(fmt.Sprintf(
				"You appear to already have that role, %s", member.Mention(),
			))
			return
		}
		ctx.Session.GuildMemberRoleAdd(guildID, userID, roleID)
		ctx.ChannelSend(
			fmt.Sprintf(
				"You have been given role `%s`, %s", req, member.Mention(),
			),
		)
		return
	}

	/* Take a role */
	if !hasRole {
		ctx.ChannelSend(fmt.Sprintf(
			"You don't have that role... How do you expect me to take it, %s?",
			member.Mention(),
		))
		return
	}
	ctx.Session.GuildMemberRoleRemove(guildID, userID, roleID)
	ctx.ChannelSend(fmt.Sprintf(
		"Taking role `%s` away, %s", req, member.Mention(),
	))
}

func (g cGate) HandleHelp(ctx *disgomux.Context) {
	//TODO: Finish this
}

func (g cGate) Settings() *disgomux.CommandSettings {
	return &disgomux.CommandSettings{
		Command:  g.Command,
		HelpText: g.HelpText,
	}
}

func (g cGate) Permissions() *disgomux.CommandPermissions {
	return &disgomux.CommandPermissions{
		RoleIDs: config.permissions[g.Command],
	}
}

func idToName(ctx *disgomux.Context, id *string) (string, error) {
	role, err := ctx.Session.State.Role(ctx.Message.GuildID, *id)
	if err != nil {
		return "", err
	}
	return role.Name, nil
}

func nameToID(ctx *disgomux.Context, name *string) (string, error) {
	roles, err := ctx.Session.GuildRoles(ctx.Message.GuildID)
	if err != nil {
		return "", errors.New("Unable to fetch roles for this guild")
	}

	for _, r := range roles {
		if strings.ToLower(r.Name) == *name {
			return r.ID, nil
		}
	}
	return "", fmt.Errorf("Role name '%s' does not exist", *name)
}

/* === End Gatekeeper Command === */

/* === Start Help Command === */

type cHelp struct {
	Command  string
	HelpText string
}

var (
	helpHandlers = make(map[string]func(ctx *disgomux.Context))
	helpFields   []*discordgo.MessageEmbedField
)

func (h cHelp) Init(m *disgomux.Mux) {
	i := 0
	for k, v := range m.Commands {
		msg := v.Settings().HelpText

		/* If there is no description, omit command from help */
		if len(msg) == 0 {
			continue
		}

		helpHandlers[k] = v.HandleHelp
		helpFields = append(helpFields, &discordgo.MessageEmbedField{
			Name:   m.Prefix + k,
			Value:  msg,
			Inline: true,
		})
		i++
	}

	cLog.WithField("command", h.Command).Infof(
		"Loaded help handlers and messages for %d commands", i,
	)
}

func (h cHelp) Handle(ctx *disgomux.Context) {
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

	command, ok := helpHandlers[ctx.Arguments[0]]
	if !ok {
		ctx.ChannelSend(fmt.Sprintf(
			"Unable to find help info for command `%s`", ctx.Arguments[0],
		))
		return
	}

	command(ctx)
}

func (h cHelp) HandleHelp(ctx *disgomux.Context) {
	ctx.ChannelSend("Are you sure _you_ don't need help?")
}

func (h cHelp) Settings() *disgomux.CommandSettings {
	return &disgomux.CommandSettings{
		Command:  h.Command,
		HelpText: h.HelpText,
	}
}

func (h cHelp) Permissions() *disgomux.CommandPermissions {
	return &disgomux.CommandPermissions{}
}

/* === End Help Command === */

/* === Begin InspiroBot Command === */

type cInspire struct {
	Command  string
	HelpText string
}

func (i cInspire) Init(m *disgomux.Mux) {
	// Nothing to init
}

func (i cInspire) Handle(ctx *disgomux.Context) {
	resp, err := http.Get("http://inspirobot.me/api?generate=true")
	if err != nil {
		cLog.WithFields(logrus.Fields{
			"error":   err,
			"command": ctx.Command,
		}).Errorf("Unable to get inspirational quote")

		ctx.ChannelSend(issueText)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			cLog.WithFields(logrus.Fields{
				"error":   err,
				"command": ctx.Command,
			}).Errorf("Unable to parse InspiroBot response")

			ctx.ChannelSend(issueText)
			return
		}

		ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID,
			&discordgo.MessageEmbed{
				Description: "~~[Favorite]() | [Delete]()~~ _Work in progress_",
				Color:       0x6dd3ff,
				Image: &discordgo.MessageEmbedImage{
					URL: string(body),
				},
			},
		)
		return
	}
	ctx.ChannelSend("Sorry, I couldn't chat with InspiroBot. Maybe try again later?")
}

func (i cInspire) HandleHelp(ctx *disgomux.Context) {
	ctx.ChannelSend("`!inspire` for a free inspirational message generated by inspirobot.me! (Use at your own risk)")
}

func (i cInspire) Settings() *disgomux.CommandSettings {
	return &disgomux.CommandSettings{
		Command:  i.Command,
		HelpText: i.HelpText,
	}
}

func (i cInspire) Permissions() *disgomux.CommandPermissions {
	return &disgomux.CommandPermissions{}
}

/* === End InspiroBot Command === */
