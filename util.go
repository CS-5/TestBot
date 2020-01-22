package main

import (
	"os"

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

func arrayContains(array []string, value string) bool {
	for _, e := range array {
		if e == value {
			return true
		}
	}
	return false
}
