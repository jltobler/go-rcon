# go-rcon

Minecraft RCON client module for connecting to Minecraft server using [RCON](https://wiki.vg/RCON) protocol written in Go.

`rcon.Client` features a concurrent-safe `Send` function that creates a separate RCON connection for each command enabling asynchronous execution if desired. 
The ability to reuse a single connection is also available by creating `rcon.Conn` directly and sending commands via `SendCommand` function until the connection is closed.

Fragmented responses over 4096 bytes are also handled and parsed into single response.

*This project is still under development and requires additional testing before it can be considered production ready.*

## Getting Started

### Installing

`go get` *will always pull the latest tagged release from the main branch.*

```sh
go get github.com/jltobler/go-rcon
```

### Usage

Import the package into your project.

```go
import "github.com/jltobler/go-rcon"
```

Construct a new RCON client which can be used to access the send function.

```go
rconClient := rcon.NewClient("rcon://localhost:25575", "password")
```

Use the Send function to request commands be remotely executed on the Minecraft server.

```go
rconClient, err := rconClient.Send("time set day")
```

Use the `CharSet` type to specify the required character set for the RCON connection. This constraints sent and received data to the specified character set.

```go
rconClient := rcon.NewClient("rcon://localhost:25575", "password", WithOptions(rcon.CharSetASCII))
```
For older Minecraft versions use `CharSetASCII`. This is also the default character set when no options are specified.

For newer Minecraft versions or those that sends Minecraft Color codes such as `§3` use `CharSetLatin1` or for servers understanding UTF8 you can use `CharSetUTF8`.

## Development

### Prerequisites

- Go (defined in `go.mod`)
- [`golangci-lint`](https://golangci-lint.run/)

### Available Commands

A `Makefile` is provided to simplify common development operations:

- **Run tests:**
  ```sh
  make test
  ```
- **Run tests with coverage:**
  ```sh
  make test-coverage
  ```
- **Run linter:**
  ```sh
  make lint
  ```
- **Format code:**
  ```sh
  make fmt
  ```
- **Tidy Go modules:**
  ```sh
  make tidy
  ```
- **Clean build artifacts:**
  ```sh
  make clean
  ```
