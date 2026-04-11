module github.com/birabittoh/disgord

go 1.25.1

require (
	github.com/birabittoh/miri v1.5.3
	github.com/bwmarrin/discordgo v0.29.0
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/joho/godotenv v1.5.1
	github.com/lmittmann/tint v1.1.3
	github.com/pion/opus v0.0.0-20260408170506-085f78c96784
)

require (
	github.com/cloudflare/circl v1.6.3 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
)

replace github.com/bwmarrin/discordgo => github.com/yeongaori/discordgo-fork v0.0.0-20260308044327-f9e3cff6c311
