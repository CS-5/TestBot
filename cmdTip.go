package main

import (
	"bufio"
	"math/rand"
	"os"

	"github.com/CS-5/disgomux"
)

type (
	cTip struct {
		Command  string
		HelpText string
	}
)

func (t cTip) Init(m *disgomux.Mux) {
	// Nothing to init
}

func (t cTip) Handle(ctx *disgomux.Context) {

	f, err := os.Open("./tips.txt")
	if err != nil {
		ctx.ChannelSend("We all have bad days. Unfortunately, this is one of mine")
		return
	}
	defer f.Close()

	var tips []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		tips = append(tips, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		ctx.ChannelSend("I dun broked D:")
	}

	tipIndex := rand.Intn(len(tips))

	ctx.ChannelSend(tips[tipIndex])
}

func (t cTip) HandleHelp(ctx *disgomux.Context) bool {
	return false
}

func (t cTip) Settings() *disgomux.CommandSettings {
	return &disgomux.CommandSettings{
		Command:  t.Command,
		HelpText: t.HelpText,
	}
}

func (t cTip) Permissions() *disgomux.CommandPermissions {
	return &disgomux.CommandPermissions{
		RoleIDs: config.permissions[t.Command],
	}
}
