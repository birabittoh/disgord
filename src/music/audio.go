package music

import (
	"bufio"
	"encoding/binary"
	"io"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/miri"
	"github.com/bwmarrin/discordgo"
	"layeh.com/gopus"
)

type Audio struct {
	playing      atomic.Bool
	Done         chan error
	opusEncoder  *gopus.Encoder
	encodeChan   chan []int16
	outputChan   chan []byte
	ffmpegStream io.ReadCloser
	ffmpegCmd    *exec.Cmd

	mu       sync.Mutex
	onFinish func()
	waitOnce sync.Once

	ms *MusicService
}

func NewAudio(track *miri.SongResult, vc *discordgo.VoiceConnection, ms *MusicService, seekTo int) (a *Audio, err error) {
	a = &Audio{
		Done:       make(chan error, 1), // Buffered to avoid blocking
		encodeChan: make(chan []int16, 450),
		outputChan: make(chan []byte, 450),
		ms:         ms,
	}
	a.playing.Store(true)

	a.opusEncoder, err = gopus.NewEncoder(gl.AudioFrameRate, gl.AudioChannels, gopus.Voip)
	if err != nil {
		ms.Logger.Error("NewEncoder Error:", err)
		return
	}

	bitrate := gl.AudioBitrate
	if gl.AudioBitrate < 1 || gl.AudioBitrate > 512 {
		bitrate = 64
	}

	a.opusEncoder.SetBitrate(bitrate * 1000)
	a.opusEncoder.SetApplication(gopus.Voip)

	a.downloader(track, seekTo, vc.GuildID)
	go a.reader()
	go a.encoder()
	go a.play_sound(vc)
	return
}

func (a *Audio) downloader(track *miri.SongResult, seekTo int, guildID string) {
	ffmpegArgs := []string{
		"-ss", strconv.Itoa(seekTo),
		"-i", "pipe:0",
		"-f", "s16le",
		"-acodec", "pcm_s16le",
		"-ar", strconv.Itoa(gl.AudioFrameRate),
		"-ac", strconv.Itoa(gl.AudioChannels),
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
	var err error
	defer func() {
		close(a.encodeChan)
	}()

	stdin := bufio.NewReaderSize(a.ffmpegStream, 16384)

	for {
		buf := make([]int16, gl.AudioFrameSize*gl.AudioChannels)

		err = binary.Read(stdin, binary.LittleEndian, &buf)
		if err == io.EOF {
			err = nil
			// Okay! There's nothing left, time to quit.
			return
		}

		if err == io.ErrUnexpectedEOF {
			// Well there's just a tiny bit left, lets encode it, then quit.
			// EncodeChan <- buf
			err = nil
			return
		}

		if err != nil {
			// Oh no, something went wrong!
			a.ms.Logger.Error("error reading from stdin,", err)
			return
		}

		// write pcm data to the EncodeChan
		a.encodeChan <- buf
	}
}

func (a *Audio) encoder() {
	defer func() {
		close(a.outputChan)
	}()

	for {
		pcm, ok := <-a.encodeChan
		if !ok {
			// if chan closed, exit
			return
		}

		// try encoding pcm frame with Opus
		opus, err := a.opusEncoder.Encode(pcm, gl.AudioFrameSize, gl.MaxBytes)
		if err != nil {
			a.ms.Logger.Error("Encoding Error:", err)
			return
		}

		// write opus data to OutputChan
		a.outputChan <- opus
	}
}

func (a *Audio) play_sound(vc *discordgo.VoiceConnection) (err error) {
	defer func() {
		if r := recover(); r != nil {
			a.ms.Logger.Error("Recovered from panic in play_sound:", r)
			err = nil
		}
		select {
		case a.Done <- err:
		default:
		}
	}()

	for a.playing.Load() {
		opus, ok := <-a.outputChan
		if !ok {
			a.playing.Store(false)
			break
		}
		if vc != nil && vc.OpusSend != nil {
			// Try to send, but recover if channel is closed
			func() {
				defer func() {
					if r := recover(); r != nil {
						a.ms.Logger.Println("OpusSend channel closed, stopping playback")
						a.playing.Store(false)
					}
				}()
				vc.OpusSend <- opus
			}()
		}
	}

	return nil
}

func (a *Audio) Stop() {
	if !a.playing.Swap(false) {
		return
	}

	// Close the ffmpeg stream
	if a.ffmpegStream != nil {
		a.ffmpegStream.Close()
	}

	// Kill the ffmpeg process if it's still running
	if a.ffmpegCmd != nil && a.ffmpegCmd.Process != nil {
		a.ffmpegCmd.Process.Kill()
		a.waitOnce.Do(func() {
			a.ffmpegCmd.Wait() // Clean up zombie process
		})
	}
}

func (a *Audio) Monitor() {
	go func() {
		err := <-a.Done
		if err != nil {
			a.ms.Logger.Errorf("Playback error: %v", err)
		}

		// Ensure cleanup happens even after normal playback
		if a.ffmpegStream != nil {
			a.ffmpegStream.Close()
		}

		if a.ffmpegCmd != nil && a.ffmpegCmd.Process != nil {
			a.waitOnce.Do(func() {
				a.ffmpegCmd.Wait() // Wait for the process to finish
			})
		}

		a.mu.Lock()
		onFinish := a.onFinish
		a.mu.Unlock()

		if onFinish != nil {
			onFinish()
		}
	}()
}

func (a *Audio) IsPlaying() bool {
	return a.playing.Load()
}

func (a *Audio) SetOnFinish(f func()) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.onFinish = f
}
