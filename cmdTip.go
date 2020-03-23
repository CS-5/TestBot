package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strings"

	"github.com/CS-5/disgomux"
	"github.com/google/go-github/v29/github"
	"github.com/teris-io/shortid"
)

type (
	cTip struct {
		Command  string
		HelpText string

		GHClient *github.Client
	}
)

var tips []string

func (t cTip) LoadTips() error {
	tips = []string{}
	if len(env.TipsURL) == 0 {
		cLog.Info("No Tips URL provided. Skipping initialization")
		return nil
	}

	resp, err := http.Get(env.TipsURL)
	if err != nil {
		cLog.WithField("error", err).Error("Unable to fetch tips config")
		return err
	}

	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		tips = append(tips, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		cLog.WithField("error", err).Error("Unable to parse tips file")
		return err
	}

	return nil
}

func (t cTip) Init(m *disgomux.Mux) {
	t.LoadTips()
}

func (t cTip) Handle(ctx *disgomux.Context) {
	if len(ctx.Arguments) == 0 {
		if len(tips) == 0 {
			ctx.ChannelSend("Sorry, I'm plumb out of knowledge to share :(")
			return
		}

		tipIndex := rand.Intn(len(tips))

		ctx.ChannelSend(tips[tipIndex])

	} else if strings.ToLower(ctx.Arguments[0]) == "add" {
		if len(ctx.Arguments) < 3 {
			ctx.ChannelSend("Hrm... That tip is very... Informative? Try again, but this time, specify a tip :)")
			return
		}

		newTip := strings.Join(ctx.Arguments[1:], " ")

		opts := github.RepositoryContentGetOptions{
			Ref: "master",
		}

		repoContents, _, _, err := t.GHClient.Repositories.GetContents(context.Background(), "PulseDevelopmentGroup", "0x626f74-data", "tips.txt", &opts)

		if err != nil {
			fmt.Println(err)
		}

		content, err := repoContents.GetContent()
		if err != nil {
			fmt.Println(err)
		}

		newContent := content + newTip + "\n"

		branchID, err := shortid.Generate()
		if err != nil {
			fmt.Println(err)
		}

		// Should probably compile this ahead of time...
		branchRegex, _ := regexp.Compile("[^a-zA-Z0-9-]")
		if err != nil {
			fmt.Println(err)
		}

		shortTip := branchRegex.ReplaceAllString(newTip[:16], "")

		branchName := "tip/" + branchID + "-" + shortTip

		currentRef, _, err := t.GHClient.Git.GetRef(context.Background(), "PulseDevelopmentGroup", "0x626f74-data", "refs/heads/master")
		if err != nil {
			fmt.Println(err)
		}

		refData := &github.Reference{
			Ref: github.String("refs/heads/" + branchName),
			Object: &github.GitObject{
				SHA: currentRef.GetObject().SHA,
			},
		}

		_, _, err = t.GHClient.Git.CreateRef(context.Background(), "PulseDevelopmentGroup", "0x626f74-data", refData)
		if err != nil {
			// This could happen if the branch already exists...
			fmt.Println(err)
		}

		newTipsOpts := &github.RepositoryContentFileOptions{
			Message: github.String("[New Tip] " + newTip),
			Content: []byte(newContent),

			Branch: github.String(branchName),
			SHA:    github.String(repoContents.GetSHA()),
		}

		_, _, err = t.GHClient.Repositories.UpdateFile(context.Background(), "PulseDevelopmentGroup", "0x626f74-data", "tips.txt", newTipsOpts)
		if err != nil {
			fmt.Println(err)
			return
		}

		prTipName := newTip

		if len(newTip) > 64 {
			prTipName = newTip[:64]
		}

		newPr := &github.NewPullRequest{
			Title:               github.String("[New Tip Suggestion] " + prTipName),
			Head:                github.String(branchName),
			Base:                github.String("master"),
			Body:                github.String(ctx.Message.Author.Username + " has suggested \n\n>" + newTip + "\n\nas a new tech tip."),
			MaintainerCanModify: github.Bool(true),
		}

		pr, _, err := t.GHClient.PullRequests.Create(context.Background(), "PulseDevelopmentGroup", "0x626f74-data", newPr)
		if err != nil {
			fmt.Println(err)
			return
		}

		ctx.ChannelSend("A pull request has been created at " + pr.GetHTMLURL())
	} else if strings.ToLower(ctx.Arguments[0]) == "reload" {
		err := t.LoadTips()
		if err != nil {
			ctx.ChannelSend("Something went wrong when reloading the tips... Try again later?")
			return
		}

		ctx.ChannelSend("Tips successfully reloaded")
	}
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
