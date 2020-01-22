package main

import (
	"github.com/CS-5/disgomux"
	"github.com/sirupsen/logrus"
)

type muxLog struct {
	logEntry *logrus.Entry
	logAll   bool
}

func (ml muxLog) Init(mux *disgomux.Mux) {
	// Nothing to init
}

func (ml muxLog) MessageRecieved(ctx *disgomux.Context) {
	if ml.logAll {
		ch, _ := ctx.Session.Channel(ctx.Message.ChannelID)
		gu, _ := ctx.Session.Guild(ctx.Message.GuildID)

		ml.logEntry.WithFields(logrus.Fields{
			"messageGuild":   gu.Name,
			"messageChannel": ch.Name,
			"messageAuthor":  ctx.Message.Author.Username,
			"messageContent": ctx.Message.Content,
		}).Info("Message Recieved")
	}
}

func (ml muxLog) CommandRegistered(cs *disgomux.CommandSettings) {
	ml.logEntry.WithField("command", cs.Command).Info(
		"Multiplexer command sucessfully registered.",
	)
}

func (ml muxLog) InitializeComplete(m *disgomux.Mux) {
	ml.logEntry.Info("Multiplexer initialization complete.")
}
