package music

import (
	"errors"
	"regexp"
	"strings"

	"github.com/birabittoh/rabbitpipe"
)

var (
	videoRegexpList = []*regexp.Regexp{ // from github.com/kkdai/youtube
		regexp.MustCompile(`(?:v|embed|shorts|watch\?v)(?:=|/)([^"&?/=%]{11})`),
		regexp.MustCompile(`(?:=|/)([^"&?/=%]{11})`),
		regexp.MustCompile(`([^"&?/=%]{11})`),
	}
)

func extractVideoID(videoID string) (string, error) {
	if strings.Contains(videoID, "youtu") || strings.ContainsAny(videoID, "\"?&/<%=") {
		for _, re := range videoRegexpList {
			if isMatch := re.MatchString(videoID); isMatch {
				subs := re.FindStringSubmatch(videoID)
				videoID = subs[1]
			}
		}
	}

	if strings.ContainsAny(videoID, "?&/<%=") {
		return "", errors.New("invalid characters in videoID")
	}

	if len(videoID) < 10 {
		return "", errors.New("videoID is too short")
	}

	return videoID, nil
}

func getFormat(video rabbitpipe.Video) *rabbitpipe.AdaptiveFormat {
	formats := video.AdaptiveFormats
	for i, format := range formats {
		if format.URL != "" && format.AudioChannels > 0 {
			return &formats[i]
		}
	}

	return nil
}

func getFromYT(videoID string) (video *rabbitpipe.Video, err error) {
	video, err = yt.GetVideo(videoID)
	if err != nil || video == nil {
		logger.Error("Error fetching video info:", err)
		return nil, errors.New("error fetching video info")
	}

	format := getFormat(*video)
	if format == nil {
		logger.Errorf("no audio formats available for video %s", videoID)
		return nil, errors.New("no audio formats available")
	}

	return
}

func search(query string) (videoID string, err error) {
	results, err := yt.Search(query)
	if err == nil && results != nil {
		for _, result := range *results {
			if result.Type == "video" && !result.LiveNow && !result.IsUpcoming && !result.Premium {
				logger.Printf("Video found by API.")

				return result.VideoID, nil
			}
		}
		err = errors.New("search did not return any valid videos")
	}

	return "", err
}

func getVideo(args []string) (*rabbitpipe.Video, error) {
	videoID, err := extractVideoID(args[0])
	if err != nil {
		videoID, err = search(strings.Join(args, " "))
		if err != nil || videoID == "" {
			return nil, err
		}
	}

	return getFromYT(videoID)
}
