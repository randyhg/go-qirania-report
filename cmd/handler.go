package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go-qirania/utils/fwRedis"
	"time"
)

func DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Printf("start processing %s request >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n", update.Message.Chat.Username)
	now := time.Now()

	// prompt processing
	msg := fmt.Sprintf("Processing.....")
	m, _ := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   msg,
	})

	queue, err := composeQueue(update.Message.Text, now)
	if err != nil {
		sendErrorMessage(err, ctx, b, update, m, err.Error())
		return
	}

	by, _ := json.Marshal(&queue)

	// send to redis for processing
	if err = fwRedis.RedisQueue().LPush(ctx, queueKey, string(by)).Err(); err != nil {
		sendErrorMessage(err, ctx, b, update, m, "Failed to Append data!")
		return
	}

	// get the status from redis
	if err = getStatus(ctx, b, update, m); err != nil {
		sendErrorMessage(err, ctx, b, update, m, "Failed to Append data!")
		return
	}
	fmt.Printf("stop processing %s request >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n\n", update.Message.Chat.Username)
}

func StartHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      defaultMsg,
		ParseMode: models.ParseModeHTML,
	})
}

func NotAuthUserHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      "You're not registered user!",
		ParseMode: models.ParseModeHTML,
	})
	saveNotAuthUserMsg(ctx, update.Message.From.ID, update.Message.From.Username, update.Message.Text)
}
