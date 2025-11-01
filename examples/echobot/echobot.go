package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/Amrakk/zcago"
	"github.com/Amrakk/zcago/api"
	"github.com/Amrakk/zcago/model"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	z := zcago.NewZalo()
	a, err := z.LoginQR(ctx, nil, nil)
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}

	ln := a.Listener()
	if err := ln.Start(ctx, true); err != nil {
		log.Fatalf("listener start error: %v", err)
	}

	defer stop()
	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-ln.Message():
			if !ok {
				return
			}

			var msg string
			switch v := m.(type) {
			case model.UserMessage:
				if v.Data.Content.String != nil {
					msg = *v.Data.Content.String
				}
			case model.GroupMessage:
				if v.Data.Content.String != nil {
					msg = *v.Data.Content.String
				}
			default:
				continue
			}

			log.Printf("%s %s\n", m.GetThreadID(), msg)
			if _, err := a.SendMessage(ctx, m.GetThreadID(), m.GetType(), api.MessageContent{Msg: msg}); err != nil {
				log.Printf("send failed (thread %s): %v", m.GetThreadID(), err)
			}
		}
	}
}
