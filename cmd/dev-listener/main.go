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

func (a *ListenerApp) startChannelListeners(ln listener.Listener) {
	// Connected channel listener
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				fmt.Println("Connected listener: context cancelled")
				return
			case <-ln.Connected():
				fmt.Println("🟢 WebSocket Connected")
			}
		}
	}()

	// Disconnected channel listener
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				fmt.Println("Disconnected listener: context cancelled")
				return
			case closeInfo := <-ln.Disconnected():
				fmt.Printf("🟡 WebSocket Disconnected - Code: %d, Reason: %s, Error: %v\n",
					closeInfo.Code, closeInfo.Reason, closeInfo.Err)
			}
		}
	}()

	// Closed channel listener - this will trigger shutdown
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				fmt.Println("Closed listener: context cancelled")
				return
			case closeInfo := <-ln.Closed():
				fmt.Printf("🔴 WebSocket Closed - Code: %d, Reason: %s, Error: %v\n",
					closeInfo.Code, closeInfo.Reason, closeInfo.Err)
				fmt.Println("Connection closed, stopping listener...")
				// a.cancel() // Trigger shutdown
				// return
			}
		}
	}()

	// Error channel listener
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				fmt.Println("Error listener: context cancelled")
				return
			case err := <-ln.Error():
				fmt.Printf("❌ WebSocket Error: %v\n", err)
			}
		}
	}()

	// Message channel listener
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				fmt.Println("Message listener: context cancelled")
				return
			case msg := <-ln.Message():
				a.handleMessage(msg)
			}
		}
	}()

	// Old Message channel listener
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				fmt.Println("Old Message listener: context cancelled")
				return
			case msg := <-ln.OldMessages():
				a.handleOldMessages(msg)
			}
		}
	}()

	// Reaction channel listener
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				fmt.Println("Reaction listener: context cancelled")
				return
			case reaction := <-ln.Reaction():
				a.handleReaction(reaction)
			}
		}
	}()

	// Old Reaction channel listener
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				fmt.Println("Old Reaction listener: context cancelled")
				return
			case reaction := <-ln.OldReactions():
				a.handleOldReactions(reaction)
			}
		}
	}()

	// Typing channel listener
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				fmt.Println("Typing listener: context cancelled")
				return
			case typing := <-ln.Typing():
				a.handleTyping(typing)
			}
		}
	}()

	// Delivered Message channel listener
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				fmt.Println("Delivered Message listener: context cancelled")
				return
			case delivered := <-ln.DeliveredMessages():
				a.handleDeliveredMessages(delivered)
			}
		}
	}()

	// Seen Message channel listener
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				fmt.Println("Seen Message listener: context cancelled")
				return
			case seen := <-ln.SeenMessages():
				a.handleSeenMessages(seen)
			}
		}
	}()

	// Undo channel listener
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				fmt.Println("Undo listener: context cancelled")
				return
			case undo := <-ln.Undo():
				a.handleUndo(undo)
			}
		}
	}()

	// Group channel listener
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				fmt.Println("Group listener: context cancelled")
				return
			case group := <-ln.Group():
				a.handleGroup(group)
			}
		}
	}()

	// Friend channel listener
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				fmt.Println("Friend listener: context cancelled")
				return
			case friend := <-ln.Friend():
				a.handleFriend(friend)
			}
		}
	}()
	// CipherKey channel listener
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				fmt.Println("CipherKey listener: context cancelled")
				return
			case key := <-ln.CipherKey():
				fmt.Printf("🔑 New Cipher Key received: %s\n", key)
			}
		}
	}()

	fmt.Println("All channel listeners started")
}
