package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/CS-5/disgomux"
	"github.com/google/go-github/v29/github"
)

type (
	cTip struct {
		Command  string
		HelpText string

		GHClient *github.Client
	}
)

func (t cTip) Init(m *disgomux.Mux) {
	// Nothing to init
}

func (t cTip) Handle(ctx *disgomux.Context) {

	if strings.ToLower(ctx.Arguments[0]) == "add" {
		opts := github.RepositoryContentGetOptions{
			Ref: "master",
		}

		fmt.Printf("%+v\n", &t.GHClient)

		repo, _, _, err := t.GHClient.Repositories.GetContents(context.Background(), "PulseDevelopmentGroup", "0x626f74-data", "TEST.md", &opts)

		if err != nil {
			fmt.Println(err)
		}

		content, err := repo.GetContent()
		if err != nil {
			fmt.Println(err)
		}

		ctx.ChannelSend(content)

		return
	}

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
