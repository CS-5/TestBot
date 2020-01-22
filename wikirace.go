package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

type articleInfo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type wikiResult struct {
	BatchComplete string `json:"batchComplete"`
	Query         struct {
		Random []articleInfo `json:"random"`
	} `json:"query"`
}

func handleWikirace(ctx *context) {
	/* TODO: Maybe float these erros up to the handler? */
	resp, err := http.Get("https://en.wikipedia.org/w/api.php?action=query&format=json&list=random&rnnamespace=0&rnlimit=2")
	if err != nil {
		log.WithFields(logrus.Fields{
			"error":   err,
			"command": ctx.Command,
		}).Error("Unable to get random wikipedia page")

		ctx.channelSend("Hmm... I seem to have run into an issue... Try again later?")
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error":   err,
			"command": ctx.Command,
		}).Error("Unable to read page")

		ctx.channelSend("Hmm... I seem to have run into an issue... Try again later?")
		return
	}

	var search wikiResult
	err = json.Unmarshal(body, &search)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error":   err,
			"command": ctx.Command,
		}).Error("Unable to unmarshal page")

		ctx.channelSend("Hmm... I seem to have run into an issue... Try again later?")
		return
	}

	articles := search.Query.Random[:2]

	ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &discordgo.MessageEmbed{
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
