# disgord
A simple Discord bot in Go. 

## Instructions
Create your own application in the [Discord Developer Portal](https://discord.com/developers/applications).
In the "Bot" tab, click on "Add Bot", then generate a new token.

Now you can copy `config.example.json` into `config.json` and insert the following:
* `applicationID`: the ID of your bot
* `token`: pretty self-explanatory
* `arlCookie`: get this from your Deezer cookies when logged in
* `secretKey`: you must find this on your own :P

Your can use the following link to add your bot to a server (replace `<applicationID>` first):
```
https://discord.com/api/oauth2/authorize?client_id=<applicationID>&permissions=17825792&scope=bot
```

## OAuth2 permissions
If you want to change the bot permissions, you can visit [this page](https://discordapi.com/permissions.html#17825792) and generate a custom link.
