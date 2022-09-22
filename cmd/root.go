package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const newFileEventType = "io.jaconi.prepper.file.new"

type NewFileEvent struct {
	Path string `json:"path"`
}

var rootCmd = &cobra.Command{
	Use:   "lzbench",
	Short: "lzbench is an in-memory benchmark of open-source LZ77/LZSS/LZMA compressors",
	Long: `lzbench is an in-memory benchmark of open-source LZ77/LZSS/LZMA compressors.
It joins all compressors into a single exe. At the beginning an input file is
read to memory. Then all compressors are used to compress and decompress the
file and decompressed file is verified. This approach has a big advantage of
using the same compiler with the same optimizations for all compressors. The
disadvantage is that it requires source code of each compressor (therefore Slug
or lzturbo are not included).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log, err := zap.NewProduction()
		if err != nil {
			return fmt.Errorf("log configuration failed: %w", err)
		}

		client, err := cloudevents.NewClientHTTP()
		if err != nil {
			return fmt.Errorf("failed to create new CloudEvent client: %w", err)
		}

		hostname, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("failed to determine hostname: %w", err)
		}

		receiver := Receiver{
			Type:   newFileEventType,
			Source: fmt.Sprintf("jaconi.io/prepper/%s", hostname),
			Logger: log,
		}

		log.Info("starting receiver on :8080")
		if err := client.StartReceiver(cmd.Context(), receiver.ReceiveAndReply); err != nil {
			return fmt.Errorf("failed to start receiver: %w", err)
		}

		return nil
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(viper.AutomaticEnv)

	viper.BindPFlags(rootCmd.Flags())
}

type Receiver struct {
	Type   string
	Source string
	Logger *zap.Logger
}

// ReceiveAndReply is invoked whenever we receive an event.
func (r Receiver) ReceiveAndReply(ctx context.Context, event cloudevents.Event) cloudevents.Result {
	// Do not acknowledge events with wrong type.
	if event.Type() != r.Type {
		r.Logger.Warn("wrong event type", zap.String("expected", r.Type), zap.String("actual", event.Type()))
		return cloudevents.NewReceipt(false, "wrong event type: expected %s but was %s", r.Type, event.Type())
	}

	// Do not acknowledge events with wrong source.
	if event.Source() != r.Source {
		r.Logger.Warn("wrong event source", zap.String("expected", r.Source), zap.String("actual", event.Source()))
		return cloudevents.NewReceipt(false, "wrong event source: expected %s but was %s", r.Source, event.Source())
	}

	var evt NewFileEvent
	if err := event.DataAs(&evt); err != nil {
		r.Logger.Error("data conversion failed", zap.Error(err))
		return cloudevents.NewHTTPResult(http.StatusBadRequest, "failed to convert data: %s", err)
	}

	// TODO: Actually do something.
	println(evt.Path)

	return cloudevents.ResultACK
}
