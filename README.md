# go-rcon

Minecraft RCON client module for connecting to Minecraft server using [RCON](https://wiki.vg/RCON) protocol written in Go.

This project is still under development and requires additional testing before it can be considered production ready.

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
rconClient := rcon.New("localhost", 25575, "password")
```

Use the Send function to request commands be remotely executed on the Minecraft server.

```go
rconClient, err := rconClient.Send("time set day")
```