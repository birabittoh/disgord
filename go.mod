module github.com/birabittoh/disgord

go 1.25.1

require (
	github.com/birabittoh/miri v1.5.2
	github.com/birabittoh/mylo v0.0.2
	github.com/bwmarrin/discordgo v0.29.0
	github.com/joho/godotenv v1.5.1
	github.com/pion/opus v0.0.0-20260122090349-7342caad2cf7
)

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/crypto v0.42.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
)

replace github.com/bwmarrin/discordgo => github.com/ozraru/discordgo v0.26.2-0.20250118163132-f992cd170161
