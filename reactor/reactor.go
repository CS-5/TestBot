package reactor

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

type (

	// Reactor defines an instance of a reaction watcher.
	Reactor struct {
		DefaultExpiration *time.Time

		watchPool map[string][]Watcher
	}

	// Watcher defines the properties to be watched for a specific message,
	// and is what makes up the watchPool.
	Watcher struct {
		Trigger string
		Handler func(ctx *Context)
		Time    time.Time
	}

	// Context defines the Reactor context, including the emoji used, channelID,
	// userID, and more.
	Context struct {
		Session  *discordgo.Session
		Reaction *discordgo.MessageReactionAdd
	}
)

// New creates a new reactor, setting the default expiration to the time
// specified.
func New(defaultExpiration *time.Time) *Reactor {
	return &Reactor{
		DefaultExpiration: defaultExpiration,
		watchPool:         make(map[string][]Watcher),
	}
}

// Handle is passed to DiscordGo to handle reaction add events.
func (r *Reactor) Handle(
	session *discordgo.Session, reaction *discordgo.MessageReactionAdd,
) {

}

// Watch is used by commands or other parts of the bot to request a given
// message be watched for reactions being added to it.
func (r *Reactor) Watch(messageID string, watchers ...Watcher) {
	if len(r.watchPool[messageID]) == 0 {
		r.watchPool[messageID] = watchers
		return
	}

	r.watchPool[messageID] = append(r.watchPool[messageID], watchers...)
}

// Unwatch is used by commands or other parts of the bot to unwatch a specific
// message or messages.
func (r *Reactor) Unwatch(messageID ...string) {
	for _, id := range messageID {
		delete(r.watchPool, id)
	}
}

// ChannelSend is a helper function for easily sending a message to the current
// channel (this is a duplicate of the ChannelSend function in the multiplexer).
func (ctx *Context) ChannelSend(message string) (*discordgo.Message, error) {
	return ctx.Session.ChannelMessageSend(ctx.Reaction.ChannelID, message)
}

// ChannelSendf is a helper function like ChannelSend for sending a formatted
// message to the current channel (this is a duplicate of the ChannelSendf
// function in the multiplexer).
func (ctx *Context) ChannelSendf(
	format string,
	a ...interface{},
) (*discordgo.Message, error) {
	return ctx.Session.ChannelMessageSend(
		ctx.Reaction.ChannelID, fmt.Sprintf(format, a...),
	)
}
