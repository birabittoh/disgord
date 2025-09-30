package music

import (
	"bufio"
	"encoding/binary"
	"io"
	"os/exec"
	"strconv"

	"github.com/birabittoh/miri"
	"github.com/bwmarrin/discordgo"
	"layeh.com/gopus"
)

const (
	AudioChannels    int    = 2
	AudioFrameRate   int    = 48000
	AudioFrameSize   int    = 960
	AudioBitrate     int    = 128
	AudioApplication string = "voip"
	MaxBytes         int    = (AudioFrameSize * AudioChannels) * 2
)

type Audio struct {
	playing      bool
	Done         chan error
	opusEncoder  *gopus.Encoder
	encodeChan   chan []int16
	outputChan   chan []byte
	ffmpegStream io.ReadCloser
	ffmpegCmd    *exec.Cmd

	ms *MusicService
}

func NewAudio(track *miri.SongResult, vc *discordgo.VoiceConnection, ms *MusicService) (a *Audio, err error) {
	a = &Audio{
		playing:    true,
		Done:       make(chan error),
		encodeChan: make(chan []int16, 450),
		outputChan: make(chan []byte, 450),
		ms:         ms,
	}

	a.opusEncoder, err = gopus.NewEncoder(AudioFrameRate, AudioChannels, gopus.Voip)
	if err != nil {
		ms.Logger.Error("NewEncoder Error:", err)
		return
	}

	bitrate := AudioBitrate
	if AudioBitrate < 1 || AudioBitrate > 512 {
		bitrate = 64
	}

	a.opusEncoder.SetBitrate(bitrate * 1000)
	a.opusEncoder.SetApplication(gopus.Voip)

	a.downloader(track)
	go a.reader()
	go a.encoder()

	go a.play_sound(vc)
	return
}

func (a *Audio) downloader(track *miri.SongResult) {
	a.ffmpegCmd = exec.Command("ffmpeg", "-i", "pipe:0", "-f", "s16le", "-acodec", "pcm_s16le", "-ar", "48000", "-ac", "2", "pipe:1")
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
		a.ms.Client.StreamTrackByID(a.ms.Ctx, strconv.Itoa(track.ID), ffmpegStdin)
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
		buf := make([]int16, AudioFrameSize*AudioChannels)

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
		opus, err := a.opusEncoder.Encode(pcm, AudioFrameSize, MaxBytes)
		if err != nil {
			a.ms.Logger.Error("Encoding Error:", err)
			return
		}

		// write opus data to OutputChan
		a.outputChan <- opus
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

func (a *Audio) Monitor(onFinish func()) {
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

		if onFinish != nil {
			onFinish()
		}
	}()
}
