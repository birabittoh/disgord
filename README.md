# disgord
A simple Discord bot in Go. 

## Instructions
Create your own application in the [Discord Developer Portal](https://discord.com/developers/applications).
In the "Bot" tab, click on "Add Bot", then generate a new token.

Now, do:
```
cp .env.example .env
```

And use your editor of choice to fill in all required variables.



Your can use the following link to add your bot to a server (replace `<applicationID>` first):
```
https://discord.com/api/oauth2/authorize?client_id=<applicationID>&permissions=17825792&scope=bot
```

## OAuth2 permissions
If you want to change the bot permissions, you can visit [this page](https://discordapi.com/permissions.html#17825792) and generate a custom link.
