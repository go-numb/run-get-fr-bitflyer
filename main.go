package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"cloud.google.com/go/firestore"
	zerolog "github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-numb/go-bitflyer"
	"github.com/go-numb/go-bitflyer/public"

	_ "github.com/joho/godotenv/autoload"
)

var PORT string

func init() {
	// OSによって環境変数を設定
	// 環境別の処理
	if runtime.GOOS == "linux" {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		PORT = fmt.Sprintf("0.0.0.0:%s", os.Getenv("PORT"))
		log.Debug().Msgf("Linuxでの処理, PORT: %s", PORT)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		PORT = fmt.Sprintf("localhost:%s", os.Getenv("PORT"))
		log.Debug().Msgf("その他のOSでの処理, PORT: %s", PORT)
	}
}

func main() {
	// api http server
	api := http.NewServeMux()

	api.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	api.HandleFunc("/api/get-ticker", get)

	// start http server
	http.ListenAndServe(PORT, api)
}

func get(w http.ResponseWriter, r *http.Request) {
	projectId := os.Getenv("PROJECTID")

	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	} else if r.Header.Get("ProjectId") != projectId {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	productCode := r.URL.Query().Get("product_code")
	if productCode == "" {
		productCode = "FX_BTC_JPY"
	}

	client := bitflyer.New("", "")

	ticker, _, err := client.Ticker(&public.Ticker{
		ProductCode: productCode,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get Ticker")
	}

	fr, _, err := client.Fr(&public.Fr{
		ProductCode: productCode,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get Fr")
	}

	log.Debug().Msgf("Ticker: %+v, fr: %+v", ticker, fr)

	// save to firestore
	app, err := firestore.NewClient(context.Background(), projectId)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create firestore client")
	}

	_, err = app.Collection("funding_rate").
		Doc(fmt.Sprintf("%s_%s", ticker.ProductCode, ticker.Timestamp)).
		Set(
			context.Background(),
			map[string]any{
				"product_code":                 "FX_BTC_JPY",
				"ticker":                       ticker,
				"current_funding_rate":         fr.CurrentFundingRate,
				"next_funding_rate_settledate": fr.NextFundingRateSettledate,
				"timestamp":                    ticker.Timestamp,
				"created_at":                   time.Now(),
			},
		)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to save firestore")
	}

	w.Write([]byte("success"))
}
