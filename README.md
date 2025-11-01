# ZCAGO

> [NOTE]
> This is an unofficial Zalo API for personal account. It work by simulating the browser to interact with Zalo Web.

> [WARNING]
> Using this API could get your account locked or banned. We are not responsible for any issues that may happen. Use it at your own risk.

---

## Table of Contents

-   [Installation](#installation)
-   [Basic Usages](#basic-usages)
    -   [Login](#login)
    -   [Listen for new messages](#listen-for-new-messages)
    -   [Send a message](#send-a-message)
-   [Example](#example)
-   [License](#license)

## Installation

```bash
go get github.com/Amrakk/zcago.git
```

---

## Documentation

See [API Documentation](https://tdung.gitbook.io/zca-js) for more details.

---

## Basic Usages

### Login

```go
import "github.com/amrakk/zcago"

cred := &zcago.Credentials{
    IMEI: "imei"
    Cookie: []
    UserAgent: "user-agent"
}
zalo := zcago.NewZalo()
api, err := zalo.Login(cred)
```

### Listen for new messages

```go
import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/Amrakk/zcago"
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

			switch v := m.(type) {
			case model.UserMessage:
				// received message in direct message
			case model.GroupMessage:
				// received message in group message
			default:
				continue
			}
		}
	}
}
```

> [IMPORTANT]
> Only one web listener can run per account at a time. If you open Zalo in the browser while the listener is active, the listener will be automatically stopped.

### Send a message

```go
import (
	"context"
	"log"

	"github.com/Amrakk/zcago"
	"github.com/Amrakk/zcago/api"
	"github.com/Amrakk/zcago/model"
)

func main() {
	ctx := context.Background()

	z := zcago.NewZalo()
	a, err := z.LoginQR(ctx, nil, nil)
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}

	userID := "123456789"
	message := api.MessageContent{Msg: "Hello from ZCAGO!"}

	response, err := a.SendMessage(ctx, userID, model.ThreadTypeUser, message)
	if err != nil {
		log.Fatalf("failed to send message: %v", err)
	}

	log.Printf("Message sent successfully! Message ID: %s", response.Message.MsgID)
}
```

---

## Example

See [examples](examples) folder for more details.

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
