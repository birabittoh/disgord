package music

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/myks"
	"github.com/kkdai/youtube/v2"
)

const (
	defaultCacheDuration = 6 * time.Hour
)

var (
	expireRegex = regexp.MustCompile(`(?i)expire=(\d+)`)
	ks          = myks.New[youtube.Video](time.Hour)
)

func getFormat(video youtube.Video) *youtube.Format {
	formats := video.Formats.Type("audio")
	for i, format := range formats {
		if format.URL != "" {
			return &formats[i]
		}
	}

	return nil
}

func parseExpiration(url string) time.Duration {
	expireString := expireRegex.FindStringSubmatch(url)
	if len(expireString) < 2 {
		return defaultCacheDuration
	}

	expireTimestamp, err := strconv.ParseInt(expireString[1], 10, 64)
	if err != nil {
		return defaultCacheDuration
	}

	return time.Until(time.Unix(expireTimestamp, 0))
}

func getFromYT(videoID string) (video *youtube.Video, err error) {
	url := "https://youtu.be/" + videoID

	const maxRetries = 3
	const maxBytesToCheck = 1024
	var duration time.Duration

	for i := 0; i < maxRetries; i++ {
		logger.Info("Requesting video", url, "attempt", i+1)
		video, err = yt.GetVideo(url)
		if err != nil || video == nil {
			logger.Error("Error fetching video info:", err)
			continue
		}

		format := getFormat(*video)
		if format == nil {
			logger.Errorf("no audio formats available for video %s", videoID)
			continue
		}

		duration = parseExpiration(format.URL)

		resp, err := http.Get(format.URL)
		if err != nil {
			logger.Error("Error fetching video URL:", err)
			continue
		}
		defer resp.Body.Close()

		if resp.ContentLength <= 0 {
			logger.Error("Invalid video link, no content length...")
			continue
		}

		buffer := make([]byte, maxBytesToCheck)
		n, err := resp.Body.Read(buffer)
		if err != nil {
			logger.Error("Error reading video content:", err)
			continue
		}

		if n > 0 {
			logger.Info("Valid video link found.")
			ks.Set(videoID, *video, duration)
			return video, nil
		}

		logger.Error("Invalid video link, content is empty...")
		time.Sleep(1 * time.Second)
	}

	err = fmt.Errorf("failed to fetch valid video after %d attempts", maxRetries)
	return nil, err
}

func getFromCache(videoID string) (video *youtube.Video, err error) {
	video, err = ks.Get(videoID)
	if err != nil {
		return
	}

	if video == nil {
		err = errors.New("video should not be nil")
		return
	}

	return
}

func getVideo(args []string) (*youtube.Video, error) {
	videoID, err := youtube.ExtractVideoID(args[0])
	if err != nil {
		searchQuery := strings.Join(args, " ")
		videoID, err = gl.Search(searchQuery)
		if err != nil || videoID == "" {
			return nil, err
		}
	}

	video, err := getFromCache(videoID)
	if err != nil {
		return getFromYT(videoID)
	}

	return video, nil
}
