package ui

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/birabittoh/disgord/src/bot"
	"github.com/birabittoh/disgord/src/config"
	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/joho/godotenv"
)

func getCommitID() string {
	file, err := os.ReadFile(filepath.Join(".git", "HEAD"))
	if err == nil {
		ref := string(file)
		if len(ref) > 5 && ref[:5] == "ref: " {
			refFile, err := os.ReadFile(filepath.Join(".git", ref[5:len(ref)-1]))
			if err == nil {
				return string(refFile[:7])
			}
		} else if len(ref) >= 7 {
			return ref[:7]
		}
	}
	return "unknown"
}

func Main() (err error) {
	godotenv.Load()

	cfg, err := config.New()
	if err != nil {
		return errors.New("could not load config: " + err.Error())
	}

	bs, err := bot.NewBotService(cfg)
	if err != nil {
		return errors.New("could not create bot service: " + err.Error())
	}

	if gl.CommitID == "" {
		gl.CommitID = getCommitID()
	}

	ui, err := NewUIService(bs)
	if err != nil {
		return errors.New("could not create ui service: " + err.Error())
	}

	return ui.Start()
}
