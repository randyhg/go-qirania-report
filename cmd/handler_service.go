package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go-qirania/utils/fwRedis"
	"go-qirania/utils/milog"
	"strconv"
	"strings"
	"time"
)

const defaultMsg = `Silahkan kirim pesan dengan format:
<code>nama_pelanggan, jenis_paket, keterangan, berat, harga</code>

Contoh:
<b>Uci, Setrika, Karpet 1, 3, 21000</b>`

var AuthUsers = []int64{1, 2, 3, 753193133, 6812309738} // 753193133

func composeQueue(text string, now time.Time) (*ProcessQueue, error) {
	queueSlice := strings.Split(text, ", ")
	if len(queueSlice) <= 1 {
		return nil, errors.New(fmt.Sprintf("<b>Format pesan salah!</b>\n\n%s", defaultMsg))
	}

	// jenis paket
	jenis := strings.TrimSpace(queueSlice[1])
	if jenis != "Cuci" && jenis != "Setrika" {
		jenis = "Cuci & Setrika"
	}

	// weight
	weight, err := strconv.ParseInt(queueSlice[3], 10, 64)
	if err != nil {
		return nil, err
	}

	// price
	price, err := strconv.ParseInt(queueSlice[4], 10, 64)
	if err != nil {
		return nil, err
	}

	// nama_pelanggan,jenis_paket,keterangan,berat,harga
	queue := &ProcessQueue{
		Name:       strings.TrimSpace(queueSlice[0]),
		Jenis:      jenis,
		Keterangan: queueSlice[2],
		Berat:      weight,
		Harga:      price,
		Waktu:      now.Format("2006-01-02"),
	}

	return queue, nil
}

func getStatus(ctx context.Context, b *bot.Bot, update *models.Update, m *models.Message) error {
	pubsub := fwRedis.RedisQueue().Subscribe(ctx, pubSubKey)

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			milog.Error(err)
			return err
		}

		b.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    m.Chat.ID,
			MessageID: m.ID,
		})

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   msg.Payload,
		})
		break
	}
	pubsub.Unsubscribe(ctx, pubSubKey)
	return nil
}

func sendErrorMessage(err error, ctx context.Context, b *bot.Bot, update *models.Update, m *models.Message, msg string) {
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    m.Chat.ID,
		MessageID: m.ID,
	})

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      msg,
		ParseMode: models.ParseModeHTML,
	})
	fmt.Printf("error while processing: %s\n", err.Error())
	fmt.Printf("stop processing %s request >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n\n", update.Message.Chat.Username)
}

func isUser(userId int64) bool {
	isAuthUser := false

	for _, authUser := range AuthUsers {
		if userId == authUser {
			isAuthUser = true
			break
		}
	}
	return isAuthUser
}

func saveNotAuthUserMsg(ctx context.Context, userId int64, userName, msg string) {
	msg = fmt.Sprintf("%d - %s says: %s", userId, userName, msg)
	if err := fwRedis.RedisQueue().LPush(ctx, notAuthUser, msg).Err(); err != nil {
		milog.Error(err)
	}
}
