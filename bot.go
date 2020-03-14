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
		Fuzzy   bool   `env:"USE_FUZZY" envDefault:"false"`
	}
)

var (
	env    = environment{}
	log    *logrus.Logger
	cLog   *logrus.Entry // Log for commanbds
	mLog   *logrus.Entry // Log for multiplexer
	config *botConfig

	prefix = "!"
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
		log.WithError(err).Error("Problem starting bot")
	}
	log.Info("Bot started")

	/* Initialize Mux */
	dMux, err := disgomux.New(prefix)
	if err != nil {
		log.WithError(err).Fatalf("Unable to create multixplexer")
	}

	/* Setup Logging */
	logMW := &muxLog{
		logAll:   env.Debug,
		logEntry: mLog,
	}

	dMux.UseMiddleware(logMW.Logger)

	/* Setup Errors */
	dMux.SetErrors(disgomux.ErrorTexts{
		CommandNotFound: "Command not found.",
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
			HelpText: "Get an inspirational quote from inspirobot.me (use at your own risk)",
		},
		cJPEG{
			Command:  "jpeg",
			HelpText: "More JPEG for the last image. 'nuff said",
		},
		cTip{
			Command:  "tip",
			HelpText: "Tech is hard. We're here to help",
		},
	)

	dMux.Options(&disgomux.Options{
		IgnoreDMs:        true,
		IgnoreBots:       true,
		IgnoreNonDefault: true,
		IgnoreEmpty:      true,
	})

	dMux.Initialize()

	if env.Fuzzy {
		dMux.InitializeFuzzy()
	}

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
		log.WithError(err).Error(
			"Problem opening websocket connection.",
		)
		return
	}

	dg.UpdateStatusComplex(discordgo.UpdateStatusData{
		Game: &discordgo.Game{
			Name: "you",
			Type: discordgo.GameTypeWatching,
			Assets: discordgo.Assets{
				LargeImageID: "watching",
				LargeText:    "Watching...",
			},
		},
		Status: "online",
	})

	defer dg.Close()

	/* Wait for interrupt */
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
