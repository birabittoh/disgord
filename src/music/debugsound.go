package music

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"os/exec"
	"strconv"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/bwmarrin/discordgo"
)

func generateWAV(freq float64, durationSec float64, sampleRate int, channels int) []byte {
	numSamples := int(float64(sampleRate) * durationSec)
	dataSize := numSamples * channels * 2

	buf := new(bytes.Buffer)
	buf.Grow(44 + dataSize)

	// RIFF header
	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, uint32(36+dataSize))
	buf.WriteString("WAVE")

	// fmt chunk
	buf.WriteString("fmt ")
	binary.Write(buf, binary.LittleEndian, uint32(16))
	binary.Write(buf, binary.LittleEndian, uint16(1))
	binary.Write(buf, binary.LittleEndian, uint16(channels))
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate))
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate*channels*2))
	binary.Write(buf, binary.LittleEndian, uint16(channels*2))
	binary.Write(buf, binary.LittleEndian, uint16(16))

	// data chunk
	buf.WriteString("data")
	binary.Write(buf, binary.LittleEndian, uint32(dataSize))

	for i := range numSamples {
		t := float64(i) / float64(sampleRate)
		sample := int16(math.Sin(2*math.Pi*freq*t) * 0.5 * 32767)
		for range channels {
			binary.Write(buf, binary.LittleEndian, sample)
		}
	}

	return buf.Bytes()
}

func newAudioFromReader(input io.Reader, vc *discordgo.VoiceConnection, ms *MusicService) (*Audio, error) {
	a := &Audio{
		playing:    true,
		Done:       make(chan error),
		outputChan: make(chan []byte, 450),
		ms:         ms,
	}

	bitrate := gl.AudioBitrate
	if bitrate < 1 || bitrate > 512 {
		bitrate = 64
	}

	ffmpegArgs := []string{
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
		return nil, err
	}
	a.ffmpegStream, _ = a.ffmpegCmd.StdoutPipe()

	if err := a.ffmpegCmd.Start(); err != nil {
		return nil, err
	}

	go func() {
		io.Copy(ffmpegStdin, input)
		ffmpegStdin.Close()
	}()

	go a.reader()
	go a.play_sound(vc)

	return a, nil
}

func (ms *MusicService) HandleDebugSound(args string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	r, _, vc := ms.us.GetVoiceChannelID(m.Member, m.GuildID, m.Author.ID)
	if r != "" {
		return ms.us.EmbedMessage(r)
	}

	voice, err := ms.GetVoiceConnection(vc, m.GuildID)
	if err != nil {
		return ms.us.EmbedMessage(gl.MsgError)
	}

	wav := generateWAV(440.0, 3.0, gl.AudioFrameRate, gl.AudioChannels)

	a, err := newAudioFromReader(bytes.NewReader(wav), voice, ms)
	if err != nil {
		ms.Logger.Error("could not create debug audio", "error", err)
		return ms.us.EmbedMessage(gl.MsgError)
	}

	a.onFinish = func() {
		voice.Disconnect(ms.us.Ctx)
	}
	a.Monitor()

	return ms.us.EmbedMessage("Playing debug tone (440Hz, 3s).")
}
