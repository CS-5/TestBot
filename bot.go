package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/CS-5/disgomux"
	"github.com/bwmarrin/discordgo"
	goenv "github.com/caarlos0/env/v6"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sirupsen/logrus"
)

type (
	environment struct {
		Token   string `env:"BOT_TOKEN"`
		Debug   bool   `env:"DEBUG" envDefault:"false"`
		DataDir string `env:"DATA_DIR" envDefault:"data/"`
	}
)

var (
	env    = environment{}
	log    *logrus.Logger
	cLog   *logrus.Entry // Log for commanbds
	mLog   *logrus.Entry // Log for multiplexer
	config *botConfig
)

func init() {
	/* Parse enviorment variables */
	if err := goenv.Parse(&env); err != nil {
		fmt.Printf("%+v\n", err)
	}

	/* Define logging setup */
	log = initLogging(env.Debug)

	/* Parse config */
	var err error
	config, err = getConfig(env.DataDir + "config.json")
	if err != nil {
		log.WithField("error", err).Error("Problem executing config command")
	}

	cLog = log.WithField("type", "command")
	mLog = log.WithField("type", "multiplexer")
}

func main() {
	/* Initialize DiscordGo */
	log.Info("Starting Bot...")
	dg, err := discordgo.New("Bot " + env.Token)
	if err != nil {
		log.WithField("error", err).Error("Problem starting bot")
	}
	log.Info("Bot started")

	/* Initialize mux */
	dMux, err := disgomux.New("!")
	if err != nil {
		log.WithField("error", err).Fatalf("Unable to create multixplexer")
	}

	dMux.Logger(muxLog{
		logEntry: mLog, logAll: true,
	})

	dMux.SetErrors(disgomux.ErrorTexts{
		CommandNotFound: "Command not found D:",
		NoPermissions:   "You do not have permissions to execute that command.",
	})

	/* === Register all the things === */

	dMux.Register(
		cDebug{
			Command:  "debug",
			HelpText: "Debuging info for bot-wranglers",
		},
		cWiki{
			Command:  "wikirace",
			HelpText: "Start a wikirace",
		},
		cGate{
			Command:  "role",
			HelpText: "Manage your access to roles, and their related channels",
		},
		cHelp{
			Command:  "help",
			HelpText: "Displays help information regarding the bot's commands",
		},
		cInspire{
			Command:  "inspire",
			HelpText: "Get an inspirational quote from inspirobot.me",
		},
		cJPEG{
			Command:  "jpeg",
			HelpText: "More JPEG for the last image. 'nuff said",
		},
	)

	dMux.Initialize()

	/* Register commands from the config file */
	for k := range config.simpleCommands {
		k := k
		dMux.RegisterSimple(disgomux.SimpleCommand{
			Command:  k,
			Content:  config.simpleCommands[k],
			HelpText: "This is a simple command",
		})
	}

	/* === End Register === */

	/* Handle commands and start DiscordGo */
	dg.AddHandler(dMux.Handle)
	err = dg.Open()
	if err != nil {
		log.WithField("error", err).Error(
			"Problem opening websocket connection.",
		)
		return
	}
	defer dg.Close()

	/* Wait for interrupt */
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
