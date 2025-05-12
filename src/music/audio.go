package music

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"os/exec"

	"github.com/bwmarrin/discordgo"
	"layeh.com/gopus"
)

type Audio struct {
	playing bool
	Done    chan error

	opusEncoder  *gopus.Encoder
	encodeChan   chan []int16
	outputChan   chan []byte
	ffmpegStream io.ReadCloser
}

var (
	AudioChannels    int    = 2
	AudioFrameRate   int    = 48000
	AudioFrameSize   int    = 960
	AudioBitrate     int    = 128
	AudioApplication string = "voip"
	MaxBytes         int    = (AudioFrameSize * AudioChannels) * 2
)

func NewAudio(url string, vc *discordgo.VoiceConnection) (a *Audio, err error) {
	a = &Audio{
		playing:    true,
		Done:       make(chan error),
		encodeChan: make(chan []int16, 450),
		outputChan: make(chan []byte, 450),
	}

	a.opusEncoder, err = gopus.NewEncoder(AudioFrameRate, AudioChannels, gopus.Voip)
	if err != nil {
		log.Println("NewEncoder Error:", err)
		return
	}

	if AudioBitrate < 1 || AudioBitrate > 512 {
		AudioBitrate = 64
	}

	a.opusEncoder.SetBitrate(AudioBitrate * 1000)
	a.opusEncoder.SetApplication(gopus.Voip)

	a.downloader(url)
	go a.reader()
	go a.encoder()

	go a.play_sound(vc)
	return
}

func (a *Audio) downloader(url string) {
	cmd := exec.Command("yt-dlp", "-f", "251", "-o", "-", url)
	ffmpeg_cmd := exec.Command("ffmpeg", "-i", "pipe:0", "-f", "s16le", "-acodec", "pcm_s16le", "pipe:1")
	ffmpeg_cmd.Stdin, _ = cmd.StdoutPipe()
	a.ffmpegStream, _ = ffmpeg_cmd.StdoutPipe()

	// Start yt-dlp command
	if err := cmd.Start(); err != nil {
		log.Println("Error starting yt-dlp command:", err)
		return
	}

	// Start ffmpeg command
	if err := ffmpeg_cmd.Start(); err != nil {
		log.Println("Error starting ffmpeg command:", err)
		return
	}

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
			//EncodeChan <- buf
			err = nil
			return
		}

		if err != nil {
			// Oh no, something went wrong!
			log.Println("error reading from stdin,", err)
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
			log.Println("Encoding Error:", err)
			return
		}

		// write opus data to OutputChan
		a.outputChan <- opus
	}
}

func (a *Audio) play_sound(vc *discordgo.VoiceConnection) (err error) {
	a.playing = true
	for a.playing {
		opus, ok := <-a.outputChan
		if !ok {
			a.playing = false
		}
		vc.OpusSend <- opus
	}

	a.Done <- err

	return nil
}

/*
	func (a *Audio) Pause() {
		if a.paused {
			return
		}

		a.paused = true
	}

	func (a *Audio) Resume() {
		if !a.paused {
			return
		}

		a.paused = false
	}
*/

func (a *Audio) Stop() {
	a.playing = false
}

func (a *Audio) Monitor(onFinish func()) {
	go func() {
		if err := <-a.Done; err != nil {
			log.Printf("Playback error: %v", err)
		}

		if onFinish != nil {
			onFinish()
		}
	}()
}
