package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func rolldice(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "roll")
	defer span.End()

	ll := logger.With("req_attr_player", r.PathValue("player"))
	ll.Info("Rolling the dice without info context")
	ll.InfoContext(ctx, "Rolling the dice with info context")

	roll := 1 + rand.Intn(6)

	ll = ll.With("roll-value", roll)

	ctx, sleepSpan := tracer.Start(ctx, "operation sleep")

	ctx, span = tracer.Start(ctx, "sleep by roll value")
	if roll > 3 {
		span.AddEvent("roll is greater than - this is a a long sleep!")
	} else {
		span.AddEvent("roll is less than 3 - I'll be up quickly")
	}
	span.SetAttributes(attribute.Int("roll-value", roll))
	time.Sleep(time.Millisecond * 10 * time.Duration(roll))
	span.End()

	ctx, span = tracer.Start(ctx, "I can do a generic sleep")
	time.Sleep(time.Millisecond * 100)
	span.End()

	sleepSpan.End()

	downstreamResp, err := callCardService(ctx, ll, roll)
	if err != nil {
		ll.ErrorContext(ctx, "error calling card service", "error", err)
		http.Error(w, "error calling card service", http.StatusInternalServerError)
		return
	}

	// ------------------Handler Logic-----------------
	var msg string
	if player := r.PathValue("player"); player != "" {
		msg = fmt.Sprintf("%s is rolling the dice", player)
	} else {
		msg = "Anonymous player is rolling the dice"
	}
	// ------------------------------------------------
	logger.InfoContext(ctx, msg, "result", roll)

	if downstreamResp != "" {
		msg = fmt.Sprintf("%s-%s", msg, downstreamResp)
	}

	rollValueAttr := attribute.Int("roll", roll)
	span.SetAttributes(rollValueAttr)
	rollCnt.Add(ctx, 1, metric.WithAttributes(rollValueAttr))

	// ------------------Response Logic-----------------
	resp := msg + "\n"
	if _, err := io.WriteString(w, resp); err != nil {
		log.Printf("Write failed: %v\n", err)
	}
	// ------------------------------------------------
}

func callCardService(ctx context.Context, ll *slog.Logger, roll int) (string, error) {
	var downstreamResp string
	if roll > 3 {
		ll.WarnContext(ctx, "roll is greater than 3 - calling card service")
		resp, err := otelhttp.Get(ctx, "http://service-card:8080/pickacard")
		if err != nil {
			ll.ErrorContext(ctx, "error calling card service", "error", err)
			return "", err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			ll.ErrorContext(ctx, "error reading card service response", "error", err)
			return "", err
		}

		downstreamResp = string(body)

		ll.InfoContext(ctx, "card service response", "response", downstreamResp)
	}
	return downstreamResp, nil
}
