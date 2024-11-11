package music

import (
	"github.com/FoxeiZ/dca"
	"github.com/bwmarrin/discordgo"
)

type Audio struct {
	session *dca.EncodeSession
	stream  *dca.StreamingSession
	paused  bool
	Done    chan error
}

var audioEncodeOptions = &dca.EncodeOptions{
	Channels:         2,
	FrameRate:        48000,
	FrameDuration:    20,
	Bitrate:          96,
	Application:      dca.AudioApplicationLowDelay,
	CompressionLevel: 10,
	PacketLoss:       1,
	BufferedFrames:   100,
	VBR:              true,
	StartTime:        0,
	VolumeFloat:      1,
	RawOutput:        true,
}

func NewAudio(url string, vc *discordgo.VoiceConnection) (as *Audio, err error) {
	as = &Audio{
		session: nil,
		stream:  nil,
		paused:  false,
		Done:    make(chan error),
	}

	as.session, err = dca.EncodeFile(url, audioEncodeOptions)
	if err != nil {
		return
	}

	as.stream = dca.NewStream(as.session, vc, as.Done)
	return
}

func (a *Audio) Pause() {
	if a.stream == nil || a.paused {
		return
	}

	a.stream.SetPaused(true)
	a.paused = true
}

func (a *Audio) Resume() {
	if a.stream == nil || !a.paused {
		return
	}

	a.stream.SetPaused(false)
	a.paused = false
}

func (a *Audio) Stop() {
	if a.stream != nil {
		a.stream.FinishNow()
		a.stream = nil
	}

	if a.session != nil {
		a.session.Stop()
		a.session.Cleanup()
		a.session = nil
	}
}

func (a *Audio) Finished() (bool, error) {
	return a.stream.Finished()
}

func (a *Audio) Monitor(onFinish func()) {
	go func() {
		for err := range a.Done {
			if err != nil {
				logger.Errorf("Playback error: %v", err)
				break
			}
		}

		a.Stop()

		if onFinish != nil {
			onFinish()
		}
	}()
}
