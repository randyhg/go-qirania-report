/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/spf13/cobra"
	"go-qirania/config"
	"go-qirania/utils/fwRedis"
	"go-qirania/utils/milog"
	"os"
	"os/signal"
)

// readCmd represents the read command
var readCmd = &cobra.Command{
	Use:   "read",
	Short: "",
	Long:  "",
	Run:   telegramStart,
}

func init() {
	rootCmd.AddCommand(readCmd)
	config.Init()
	fwRedis.RedisInit()
}

func telegramStart(cmd *cobra.Command, args []string) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithMiddlewares(MiddlewareCheck),
		bot.WithDefaultHandler(DefaultHandler),
	}
	b, err := bot.New(config.Conf.BotToken, opts...)
	if err != nil {
		panic(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, StartHandler)

	fmt.Println("================================ Telegram Bot Started ================================")
	b.Start(ctx)
}

func MiddlewareCheck(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message != nil {
			milog.Infof("%d say: %s", update.Message.From.ID, update.Message.Text)

			// check
			if !isUser(update.Message.From.ID) {
				NotAuthUserHandler(ctx, b, update)
			} else {
				next(ctx, b, update)
			}
		}
	}
}
