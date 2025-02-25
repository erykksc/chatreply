# ChatReply

## Example usage
Run the command and pipe input into its _stdin_

_Tip: By default the messages are splitted by newline character_  
_(In the example the '\\' is used to escape '!' symbol)_
```bash
echo "React to this message\!
 This is a second message" | go run chatreply.go
```

Bot starts to watching for reaction on the messages (indicated by the eyes emoji)

![Watching for reaction on discord](./readme-assets/discord-watching-message.png)

Upon reacting to a message the eyes emoji dissapears  
_Tip: Both emoji reactions and text responses are supported_

![Reacting to discord message](./readme-assets/discord-reaction.png)

And original messages including the replies are outputed to _stdout_ in real time (once the program sees the reply)

![Cli output with reaction](./readme-assets/cli-output.png)

## Installation
You can install it using go

```bash
go install github.com/erykksc/chatreply
```

_Tip: Later on you can use this tool with "chatreply" command instead of "go run chatreply.go"_

## Configuration
### From .dotfiles
You need to specify the providers and their configuration in
`$XDG_CONFIG_HOME/chatreply/conf.toml`

The toml fields and syntax:

```toml
ActiveProvider = "discord"

[Discord]
UserID = "<YOUR-USER-ID>"
Token = "<YOUR-TOKEN>"

[Telegram]
ChatID = "<YOUR-CHAT-ID>"
Token = "<YOUR-API-TOKEN>"
```

### Specify path
You can also specify the path of the .toml config file as an argument
```shell
chatreply -f "./config-file.toml"
```

## Multimedia support

By default the tool will try to parse lines as file paths. For example:
```
echo "./images/orange.png" | chatreply
```

will make __chatreply__ try to open "./images/orange.png" file and send it as an attachment to the chat.  
On the output it will still be just text.

__Tip: You can disable this behaviour using a "-text-only" flag__

### Example
Discord view

![Discord with an open image](./readme-assets/image-discord.png)

CLI view

![Discord image cli output](./readme-assets/image-discord-output.png)


## Flags and arguments
```
❯ chatreply -h
Usage of chatreply:
  -f string
        Filepath of the config .toml file (default "$XDG_CONFIG_HOME/chatreply/conf.toml")
  -msg-sep string
        Separator between messages (default "\n")
  -out-sep string
        Separator between output messages (default "\n")
  -replies-count int
        Number of replies to wait per message for before exiting, -1 will wait indefinitely, allowing multiple replies per message (default 1)
  -s string
        Separator between message and emoji (default ":")
  -skip-replies
        Do not wait for replies, just send the messages
  -text-only
        Make all messages text only, disable trying to parse messages as multimedia
  -v    Sets logging level to Debug
  -watch-emoji string
        Emoji used to indicate the program is watching the message for a reply (default "👀")
```
