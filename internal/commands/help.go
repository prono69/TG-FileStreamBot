package commands

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
)

func (m *command) LoadHelp(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("help")
	defer log.Sugar().Info("Loaded")
	dispatcher.AddHandler(handlers.NewCommand("help", help))
}

func help(ctx *ext.Context, u *ext.Update) error {
	chatId := u.EffectiveChat().GetID()

	peer := ctx.PeerStorage.GetPeerById(chatId)
	if peer != nil && peer.Type != int(storage.TypeUser) {
		return dispatcher.EndGroups
	}

	if len(config.ValueOf.AllowedUsers) != 0 && !utils.Contains(config.ValueOf.AllowedUsers, chatId) {
		ctx.Reply(u, ext.ReplyTextString("You are not allowed to use this bot."), nil)
		return dispatcher.EndGroups
	}

	msg := `🤖 **File Stream Bot Help**

This bot allows you to generate **direct streamable/download links** for Telegram files.

📌 **How to use:**
1. Send any file to the bot
2. Bot will generate a streaming link
3. Open link in browser or share it

⚡ **Commands:**

/start - Start the bot  
/help - Show this help message  
/usage - Show server & bot stats  

🚀 Fast • Simple • Streamable`

	ctx.Reply(u, ext.ReplyTextString(msg), nil)
	return dispatcher.EndGroups
}