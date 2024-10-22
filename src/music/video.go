package music

import (
	"strings"

	gl "github.com/BiRabittoh/disgord/src/globals"
	"github.com/kkdai/youtube/v2"
)

func getVideo(args []string) (*youtube.Video, error) {
	video, err := yt.GetVideo(args[0])
	if err == nil {
		return video, nil
	}

	if !strings.HasPrefix(err.Error(), "extractVideoID failed") {
		return nil, err
	}

	id, err := gl.Search(args)
	if err != nil || id == "" {
		return nil, err
	}

	return yt.GetVideo(id)
}
