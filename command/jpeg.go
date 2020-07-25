package command

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"regexp"

	"github.com/PulseDevelopmentGroup/0x626f74/log"
	"github.com/PulseDevelopmentGroup/0x626f74/multiplexer"
	"github.com/PulseDevelopmentGroup/0x626f74/util"
	"github.com/bwmarrin/discordgo"
	"github.com/disintegration/imaging"
)

// JPEG is a bot command
type JPEG struct {
	Command  string
	HelpText string

	Logger *log.Logs
}

var (
	imgSaturation float64 = 100
	imgBlur       float64 = 3
	imgQuality    int     = 1

	urlRE = regexp.MustCompile(`(http(s?):)([/|.|\w|\s|-])*\.(?:jpg|jpeg|png|JPG|JPEG|PNG)`)
)

// Init is called by the multiplexer before the bot starts to initialize any
// variables the command needs.
func (c JPEG) Init(m *multiplexer.Mux) {
	// Nothing to init
}

// Handle is called by the multiplexer whenever a user triggers the command.
func (c JPEG) Handle(ctx *multiplexer.Context) {
	ctx.Session.ChannelTyping(ctx.Message.ChannelID)

	urls, err := c.getURLs(ctx)
	if err != nil {
		ctx.ChannelSendf("Unable to get URLs for JPEGing: `%s`", err)
		return
	}

	for _, url := range urls {
		req, err := http.Get(url)
		if err != nil {
			c.Logger.CmdErr(ctx, err, "There was a problem getting the attachment")
			return
		}
		defer req.Body.Close()

		imgIn, _, err := image.Decode(req.Body)
		if err != nil {
			c.Logger.CmdErr(ctx, err, "There was a problem decoding the image")
			return
		}

		/* Tweak these values to adjust JPEGness */
		img1 := imaging.AdjustSaturation(imgIn, imgSaturation)
		imgOut := imaging.Blur(img1, imgBlur)

		var buf bytes.Buffer // Buffer to return image
		err = jpeg.Encode(&buf, imgOut, &jpeg.Options{
			Quality: imgQuality,
		})
		if err != nil {
			c.Logger.CmdErr(ctx, err, "There was a problem endoding the image")
			return
		}

		ctx.Session.ChannelFileSend(
			ctx.Message.ChannelID,
			"compressed.jpeg",
			&buf,
		)
	}
}

func (c JPEG) getURLs(ctx *multiplexer.Context) ([]string, error) {
	var urls []string

	if len(ctx.Arguments) != 0 && urlRE.MatchString(ctx.Arguments[0]) {
		return append(urls, ctx.Arguments[0]), nil
	}

	message, err := c.getMessage(ctx)
	if err != nil {
		return urls, err
	}

	if len(message.Attachments) == 0 || message.Attachments[0] == nil {
		urls = urlRE.FindAllString(message.Content, -1)
	} else {
		for _, attach := range message.Attachments {
			if urlRE.MatchString(attach.ProxyURL) {
				urls = append(urls, attach.ProxyURL)
			}
		}
	}

	return urls, nil
}

func (c JPEG) getMessage(ctx *multiplexer.Context) (*discordgo.Message, error) {
	if len(ctx.Arguments) == 0 {
		messages, err := ctx.Session.ChannelMessages(
			ctx.Message.ChannelID, 1, ctx.Message.ID, "", "",
		)
		if err != nil {
			return nil, err
		}

		return messages[len(messages)-1], nil
	}

	if util.IsID(ctx.Arguments[0]) {
		message, err := ctx.Session.ChannelMessage(
			ctx.Message.ChannelID, ctx.Arguments[0],
		)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to get message with ID '%s'", ctx.Arguments[0],
			)
		}

		return message, nil
	}

	return nil, fmt.Errorf(
		"'%s' doesn't look like a message ID or URL", ctx.Arguments[0],
	)
}

// HandleHelp is called by whatever help command is in place when a user enters
// "!help [command name]". If the help command is not being handled, return
// false.
func (c JPEG) HandleHelp(ctx *multiplexer.Context) bool {
	ctx.ChannelSend(
		"`!jpeg` to JPEGify the image that was just sent.\n" +
			"`!jpeg [message ID]` to JPEGify a specific image in this channel.",
	)
	return true
}

// Settings is called by the multiplexer on startup to process any settings
// associated with that command.
func (c JPEG) Settings() *multiplexer.CommandSettings {
	return &multiplexer.CommandSettings{
		Command:  c.Command,
		HelpText: c.HelpText,
	}
}
