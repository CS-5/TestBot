package main

import (
	"os"
	"strings"

	"github.com/CS-5/disgomux"
	"github.com/sirupsen/logrus"
)

func initFile(path string) (*os.File, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		file, err := os.Create(path)

		if err != nil {
			return &os.File{}, err
		}
		return file, err
	}
	return file, err
}

func initLogging(debug bool) *logrus.Logger {
	logrus.SetOutput(os.Stdout)
	log := logrus.New()

	logrus.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})

	if !debug {
		logrus.SetLevel(logrus.InfoLevel)
		log.SetFormatter(&logrus.JSONFormatter{})
	}
	return log
}

func arrayContains(array []string, value string, ignoreCase bool) bool {
	for _, e := range array {
		if ignoreCase {
			e = strings.ToLower(e)
		}

		if e == value {
			return true
		}
	}
	return false
}

func cmdIssue(ctx *disgomux.Context, e error, msg string) {
	cLog.WithFields(logrus.Fields{
		"command": ctx.Command,
		"error":   e.Error(),
	}).Error(msg)

	if env.Debug {
		ctx.ChannelSendf(msg+"\nError: `%s`", e.Error())
		return
	}
	ctx.ChannelSend(msg)
}
