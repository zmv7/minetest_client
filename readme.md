
Minetest console client, written in go

State: **WIP**

# About

Console client for minetest

Features:
* Log into remote servers
* Ping servers
* Download media
* Listen to various server events

# Usage

Requirements:
* golang >= 1.17

```bash
go build
```

```
# ./minetest_client --help
Usage of ./minetest_client:
  -help
    	Shows the help
  -host string
    	The hostname (default "127.0.0.1")
  -media
    	Download media
  -password string
    	The password (default "enter")
  -ping
    	Just ping the given host:port and exit
  -port int
    	The portname (default 30000)
   -skip int
      Skip some SRP sending to prevent 'access denied'
  -stalk
    	Stalk mode: don't really join, just listen
  -username string
    	The username (default "test")
```
* Note: use -skip <number>(usually 1-2) if server sending HELLO(and clients sends SRP bytes A) more than 1 time
# License

MIT
