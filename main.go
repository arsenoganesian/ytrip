package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var videoURLRegexp = regexp.MustCompile(`(?:https?:\/\/)?(?:www\.)?youtu\.?be(?:\.com)?\/?.*(?:watch|embed)?(?:.*v=|v\/|\/)([\w-_]+)`)
var filenameRegexp = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

func main() {
	botToken := os.Getenv("TG_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Starting bot")

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Println("Processing:", update.Message.Chat.ID, update.Message.Text)

		if videoURLRegexp.MatchString(update.Message.Text) {
			log.Println("Downloading:", update.Message.Text)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Downloading ..."))
			audioFile, err := downloadAudio(update.Message.Text)
			if err != nil {
				log.Println("Error:", err)
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Download failed"))
				continue
			}
			if _, err := bot.Send(tgbotapi.NewAudioUpload(update.Message.Chat.ID, audioFile)); err == nil {
				log.Println("Sent:", audioFile)
			} else {
				log.Println("Error:", err)
			}
		} else {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome to ytrip, send yt link here"))
			log.Println("Sent:", update.Message.Chat.ID, "Default")
		}
	}
}

func downloadAudio(videoURL string) (string, error) {
	filename, err := getFilename(videoURL)

	if err != nil {
		return "", fmt.Errorf("failed to get filename: %v\n%s", err, videoURL)
	}

	err = exec.Command(
		"yt-dlp",
		"--audio-format",
		"mp3",
		"-x",
		"-o", filename,
		videoURL,
	).Run()

	if err != nil {
		return "", fmt.Errorf("failed to download audio: %v\n%s", err, filename)
	}

	fi, err := os.Stat(filename)

	if err != nil {
		return "", fmt.Errorf("failed to open audio: %v\n%s", err, filename)
	}

	return fi.Name(), nil
}

func getFilename(videoURL string) (string, error) {
	output, err := exec.Command(
		"yt-dlp",
		"-O", "%(title)s %(id)s",
		videoURL,
	).CombinedOutput()

	if err != nil {
		return "", err
	}

	filename := strings.TrimSpace(string(output))
	filename = filenameRegexp.ReplaceAllString(filename, "")
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = filename + ".mp3"
	p := path.Clean(filename)

	log.Println("Path:", p)

	return p, nil
}
