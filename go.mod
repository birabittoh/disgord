module github.com/birabittoh/disgord

go 1.25.1

require (
	github.com/birabittoh/miri v1.5.5
	github.com/bwmarrin/discordgo v0.29.0
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/joho/godotenv v1.5.1
	github.com/lmittmann/tint v1.1.3
	github.com/pion/opus v0.1.0
	github.com/playwright-community/playwright-go v0.6000.0
)

require (
	github.com/cloudflare/circl v1.6.4 // indirect
	github.com/deckarep/golang-set/v2 v2.9.0 // indirect
	github.com/go-jose/go-jose/v3 v3.0.5 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	go.mongodb.org/mongo-driver v1.17.9 // indirect
	golang.org/x/crypto v0.53.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
)

replace github.com/bwmarrin/discordgo => github.com/yeongaori/discordgo-fork v0.0.0-20260308044327-f9e3cff6c311
