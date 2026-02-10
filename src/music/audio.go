package music

import (
	"bytes"
	"io"
	"os/exec"
	"strconv"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/miri"
	"github.com/bwmarrin/discordgo"
	"github.com/pion/opus/pkg/oggreader"
)

type Audio struct {
	playing      bool
	Done         chan error
	outputChan   chan []byte
	ffmpegStream io.ReadCloser
	ffmpegCmd    *exec.Cmd
	onFinish     func()

	ms *MusicService
}

func NewAudio(track *miri.SongResult, vc *discordgo.VoiceConnection, ms *MusicService, seekTo int) (a *Audio, err error) {
	a = &Audio{
		playing:    true,
		Done:       make(chan error),
		outputChan: make(chan []byte, 450),
		ms:         ms,
	}

	bitrate := gl.AudioBitrate
	if gl.AudioBitrate < 1 || gl.AudioBitrate > 512 {
		bitrate = 64
	}

	a.downloader(track, seekTo, vc.GuildID, bitrate)
	go a.reader()
	go a.play_sound(vc)
	return
}

func (a *Audio) downloader(track *miri.SongResult, seekTo int, guildID string, bitrate int) {
	ffmpegArgs := []string{
		"-ss", strconv.Itoa(seekTo),
		"-i", "pipe:0",
		"-c:a", "libopus",
		"-b:a", strconv.Itoa(bitrate) + "k",
		"-ar", strconv.Itoa(gl.AudioFrameRate),
		"-ac", strconv.Itoa(gl.AudioChannels),
		"-frame_duration", "20",
		"-application", "voip",
		"-f", "ogg",
		"pipe:1",
	}

	a.ffmpegCmd = exec.Command("ffmpeg", ffmpegArgs...)
	ffmpegStdin, err := a.ffmpegCmd.StdinPipe()
	if err != nil {
		a.ms.Logger.Error("Error creating ffmpeg stdin pipe:", err)
		return
	}
	a.ffmpegStream, _ = a.ffmpegCmd.StdoutPipe()

	if err := a.ffmpegCmd.Start(); err != nil {
		a.ms.Logger.Error("Error starting ffmpeg command:", err)
		return
	}

	// Stream track directly into ffmpeg's Stdin
	go func() {
		q := a.ms.GetQueue(guildID)
		if q == nil {
			a.ms.Logger.Error("Queue not found for guild:", guildID)
			ffmpegStdin.Close()
			return
		}

		// Stream the track into ffmpeg's stdin
		err := q.client.StreamTrackByID(a.ms.us.Ctx, track.ID, ffmpegStdin)
		if err != nil {
			a.ms.Logger.Error("Error streaming track to ffmpeg stdin:", err)
		}
		ffmpegStdin.Close()
	}()
}

func (a *Audio) reader() {
	defer func() {
		close(a.outputChan)
	}()

	ogg, _, err := oggreader.NewWith(a.ffmpegStream)
	if err != nil {
		a.ms.Logger.Error("Error creating ogg reader:", err)
		return
	}

	var packet []byte
	packetCount := 0
	for {
		segments, _, err := ogg.ParseNextPage()
		if err == io.EOF {
			return
		}

		if err != nil {
			a.ms.Logger.Error("error reading from ogg stream,", err)
			return
		}

		for _, segment := range segments {
			packet = append(packet, segment...)
			if len(segment) < 255 {
				if packetCount < 2 {
					packetCount++
					packet = nil
					continue
				}
				if len(packet) > 0 {
					a.outputChan <- packet
					packet = nil
				}
			}
		}
	}
}

func (a *Audio) play_sound(vc *discordgo.VoiceConnection) (err error) {
	a.playing = true
	defer func() {
		if r := recover(); r != nil {
			a.ms.Logger.Error("Recovered from panic in play_sound:", r)
			err = nil
		}
		a.Done <- err
	}()

	for a.playing {
		opus, ok := <-a.outputChan
		if !ok {
			a.playing = false
			break
		}
		if vc != nil && vc.OpusSend != nil {
			// Try to send, but recover if channel is closed
			func() {
				defer func() {
					if r := recover(); r != nil {
						a.ms.Logger.Println("OpusSend channel closed, stopping playback")
						a.playing = false
					}
				}()
				vc.OpusSend <- opus
			}()
		}
	}

	return nil
}

func (a *Audio) Stop() {
	a.playing = false

	// Close the ffmpeg stream
	if a.ffmpegStream != nil {
		a.ffmpegStream.Close()
	}

	// Kill the ffmpeg process if it's still running
	if a.ffmpegCmd != nil && a.ffmpegCmd.Process != nil {
		a.ffmpegCmd.Process.Kill()
		a.ffmpegCmd.Wait() // Clean up zombie process
	}
}

func (a *Audio) Monitor() {
	go func() {
		if err := <-a.Done; err != nil {
			a.ms.Logger.Errorf("Playback error: %v", err)
		}

		// Ensure cleanup happens even after normal playback
		if a.ffmpegStream != nil {
			a.ffmpegStream.Close()
		}

		if a.ffmpegCmd != nil && a.ffmpegCmd.Process != nil {
			a.ffmpegCmd.Wait() // Wait for the process to finish
		}

		if a.onFinish != nil {
			a.onFinish()
		}
	}()
}
