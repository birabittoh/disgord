package music

import (
	"strings"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/kkdai/youtube/v2"
)

func getVideo(args []string) (*youtube.Video, error) {
	video, err := yt.GetVideo(args[0])
	if err == nil {
		return video, nil
	}

	id, err := gl.Search(strings.Join(args, " "))
	if err != nil || id == "" {
		return nil, err
	}

	return yt.GetVideo(id)
}
