package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/Amrakk/zcago"
	"github.com/Amrakk/zcago/listener"
	"github.com/Amrakk/zcago/model"
)

type ListenerApp struct {
	zalo     zcago.Zalo
	credPath string
	isDebug  bool
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	app := &ListenerApp{
		zalo:     zcago.NewZalo(zcago.WithLogLevel(2), zcago.WithSelfListen(true)),
		credPath: filepath.Join(rootDir(), "cmd", "credentials.json"),
		isDebug:  true,
	}

	ctx, cancel := context.WithCancel(context.Background())
	app.ctx = ctx
	app.cancel = cancel

	go func() {
		<-sigCh
		fmt.Println("\n🛑 Shutdown signal received, stopping listener...")
		cancel()
	}()

	if err := app.run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func (a *ListenerApp) run() error {
	api, err := a.authenticate()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Println("Starting listener...")

	ln := api.Listener()
	if ln == nil {
		return fmt.Errorf("failed to get listener")
	}

	a.startChannelListeners(ln)

	if err := ln.Start(a.ctx, true); err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	if err := ln.RequestOldReactions(a.ctx, model.ThreadTypeGroup, nil); err != nil {
		return fmt.Errorf("failed to request old reactions: %w", err)
	}
	if err := ln.RequestOldMessages(a.ctx, model.ThreadTypeGroup, nil); err != nil {
		return fmt.Errorf("failed to request old messages: %w", err)
	}

	fmt.Println("Listener started. Press Ctrl+C to stop...")

	a.wg.Wait()

	fmt.Println("All listeners stopped. Exiting.")
	return nil
}

func spawn[T any](a *ListenerApp, name string, ch <-chan T, handle func(T)) {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				fmt.Printf("%s listener: context cancelled\n", name)
				return
			case v, ok := <-ch:
				if !ok {
					fmt.Printf("%s listener: channel closed\n", name)
					return
				}
				handle(v)
			}
		}
	}()
}

func (a *ListenerApp) startChannelListeners(ln listener.Listener) {
	// Note the prefix: spawn(a, ...)
	spawn(a, "Connected", ln.Connected(), func(_ struct{}) {
		fmt.Println("🟢 WebSocket Connected")
	})

	spawn(a, "Disconnected", ln.Disconnected(), func(ci listener.CloseInfo) {
		fmt.Printf("🟡 WebSocket Disconnected - Code: %d, Reason: %s, Error: %v\n",
			ci.Code, ci.Reason, ci.Err)
	})

	spawn(a, "Closed", ln.Closed(), func(ci listener.CloseInfo) {
		fmt.Printf("🔴 WebSocket Closed - Code: %d, Reason: %s, Error: %v\n",
			ci.Code, ci.Reason, ci.Err)
		fmt.Println("Connection closed, stopping listener...")
		// a.cancel() // Uncomment to trigger shutdown
	})

	spawn(a, "Error", ln.Error(), func(err error) {
		fmt.Printf("❌ WebSocket Error: %v\n", err)
	})

	spawn(a, "Message", ln.Message(), func(m model.Message) {
		a.handleMessage(m)
	})

	spawn(a, "Old Message", ln.OldMessages(), func(m model.OldMessages) {
		a.handleOldMessages(m)
	})

	spawn(a, "Reaction", ln.Reaction(), func(r model.Reaction) {
		a.handleReaction(r)
	})

	spawn(a, "Old Reaction", ln.OldReactions(), func(r model.OldReactions) {
		a.handleOldReactions(r)
	})

	spawn(a, "Typing", ln.Typing(), func(t model.Typing) {
		a.handleTyping(t)
	})

	spawn(a, "Delivered Message", ln.DeliveredMessages(), func(d []model.DeliveredMessage) {
		a.handleDeliveredMessages(d)
	})

	spawn(a, "Seen Message", ln.SeenMessages(), func(s []model.SeenMessage) {
		a.handleSeenMessages(s)
	})

	spawn(a, "Undo", ln.Undo(), func(u model.Undo) {
		a.handleUndo(u)
	})

	spawn(a, "Group", ln.Group(), func(g model.GroupEvent) {
		a.handleGroup(g)
	})

	spawn(a, "Friend", ln.Friend(), func(f model.FriendEvent) {
		a.handleFriend(f)
	})

	spawn(a, "CipherKey", ln.CipherKey(), func(key string) {
		fmt.Printf("🔑 New Cipher Key received: %s\n", key)
	})

	fmt.Println("All channel listeners started")
}
