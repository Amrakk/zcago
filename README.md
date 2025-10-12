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
api, err:= zalo.Login(cred)
```

### Listen for new messages

```go

```

> [IMPORTANT]
> Only one web listener can run per account at a time. If you open Zalo in the browser while the listener is active, the listener will be automatically stopped.

### Send a message

```go

```

---

## Example

See [examples](examples) folder for more details.

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
