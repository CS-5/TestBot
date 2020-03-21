package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"net/http"
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

var tips []string

func (t cTip) Init(m *disgomux.Mux) {
	resp, err := http.Get(env.TipsURL)
	if err != nil {
		cLog.WithField("error", err).Error("Unable to fetch tips config")
	}

	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		tips = append(tips, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		cLog.WithField("error", err).Error("Unable to parse tips file")
	}
}

func (t cTip) Handle(ctx *disgomux.Context) {

	if len(ctx.Arguments) > 0 && strings.ToLower(ctx.Arguments[0]) == "add" {
		opts := github.RepositoryContentGetOptions{
			Ref: "master",
		}

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
