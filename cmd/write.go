/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go-qirania/config"
	"go-qirania/utils/fwRedis"
	"go-qirania/utils/milog"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"time"

	"github.com/spf13/cobra"
)

// writeCmd represents the write command
var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "",
	Long:  "",
	Run:   process,
}

const (
	queueKey    = "process_queue"
	queueKeyErr = "process_queue_err"
	pubSubKey   = "status_channel"
	notAuthUser = "not_auth_user_msg"
)

type ProcessQueue struct {
	Name       string `json:"name"`
	Jenis      string `json:"jenis"`
	Keterangan string `json:"keterangan"`
	Berat      int64  `json:"berat"`
	Waktu      string `json:"waktu"`
	Harga      int64  `json:"harga"`
}

func init() {
	rootCmd.AddCommand(writeCmd)
	config.Init()
	fwRedis.RedisInit()
}

func process(cmd *cobra.Command, args []string) {
	client, err := GetHttpClient(config.Conf.CredentialPath)
	if err != nil {
		milog.Fatal("Unable to get http client: %s", err.Error())
	}

	// cron for create next month spreadsheet
	go InitTimer(client)

	milog.Debug("processing queue:", queueKey)

	ctx := context.Background()
	fmt.Println("================================ Job Process Started ================================")
	for {
		q, err := fwRedis.RedisQueue().RPop(ctx, queueKey).Result()

		if errors.Is(err, redis.Nil) {
			milog.Debug(queueKey, ": no queue")
			i := time.Duration(config.Conf.DelayWhenNoJobInSeconds)
			time.Sleep(i * time.Second)
			continue
		}

		if err != nil {
			milog.Error(queueKey, "Error getting queue", err)
			i := time.Duration(config.Conf.DelayWhenErrorInSeconds)
			time.Sleep(i * time.Second)
			continue
		}

		queue := parseQueue(q)

		srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			errorHandler(queue, ctx)
			milog.Fatalf("Unable to retrieve Sheets client: %s", err.Error())
		}

		values := [][]interface{}{
			{queue.Name, queue.Jenis, queue.Keterangan, queue.Berat, queue.Waktu, queue.Harga},
		}

		var vr sheets.ValueRange
		vr.Values = values
		writeRange := fmt.Sprintf("%s!%s", time.Now().Month().String(), config.Conf.CellRange)

		_, err = srv.Spreadsheets.Values.Append(config.Conf.SpreadSheetId, writeRange, &vr).ValueInputOption("RAW").Do()
		if err != nil {
			errorHandler(queue, ctx)
			milog.Fatalf("Unable to append data to sheet: %v", err)
		}
		msg := fmt.Sprintf("Data %s berhasil di input!", queue.Name)
		err = fwRedis.RedisQueue().Publish(ctx, pubSubKey, msg).Err()
		if err != nil {
			milog.Error(err)
			break
		}
		milog.Infof("%s data appended succesfully!", queue.Name)
		i := time.Duration(config.Conf.DelayWhenJobDoneInSeconds)

		time.Sleep(i * time.Second)
	}
}

func errorHandler(queue ProcessQueue, ctx context.Context) {
	by, _ := json.Marshal(&queue)
	redisErr := fwRedis.RedisQueue().LPush(ctx, queueKeyErr, string(by)).Err()
	if redisErr != nil {
		milog.Error("Job : has error during record to error queue", redisErr.Error())
		milog.Debug("Manual trigger queue run: \nLPUSH "+queueKeyErr+"'"+string(by)+"'", "\nerr ===", redisErr.Error())
	}

	msg := fmt.Sprintf("Data %s gagal di input!", queue.Name)
	redisErr = fwRedis.RedisQueue().Publish(ctx, pubSubKey, msg).Err()
	if redisErr != nil {
		milog.Error(redisErr)
	}
}

func parseQueue(queue string) ProcessQueue {
	procQueue := ProcessQueue{}
	_ = json.Unmarshal([]byte(queue), &procQueue)
	return procQueue
}
