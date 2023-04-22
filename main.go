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

var youtubeLinkRegexp = regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?youtube.com\/watch\?v=[\w-]+&?(?:\S+)?|https?:\/\/(?:www\.)?youtu.be\/[\w-]+`)

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

		youtubeLinks := findYoutubeLinks(update.Message.Text)

		if len(youtubeLinks) > 0 {
			log.Println("Downloading:", update.Message.Text)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Downloading ..."))
			for _, link := range youtubeLinks {
				audioFile, err := downloadAudio(link)
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
			}
		} else {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Send yt link here"))
			log.Println("Sent:", update.Message.Chat.ID, "Default")
		}
	}
}

func findYoutubeLinks(text string) []string {
	return youtubeLinkRegexp.FindAllString(text, -1)
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
		"--restrict-filenames",
		"--no-warnings",
		"--print", "filename",
		"-o", "%(title)s.mp3",
		videoURL,
	).CombinedOutput()

	if err != nil {
		return "", err
	}

	p := path.Clean(strings.TrimSpace(string(output)))

	log.Println("Path:", p)

	return p, nil
}
