# GophKeeper

## Run

0. Start database `make db-start`. Starts database server in docker (postgresql)
1. Run server `go run cmd/server/main.go -c=.var/conf/server.cfg` to start server
2. Run client `go run cmd/client/main.go -c=.var/conf/client.cfg` to start client app in TUI mode

## Instructions

1. Register user in server `./client -c=".var/conf/client.cfg" register`.
    - Enter login & password on request.
1. Create new item `./client -c=".var/conf/client.cfg" item store`.
    - Enter Keychain pass (your private password, NOT password from server).
    - Enter requested values (Label, Comment, select Type, Secret and other meta data).
1. Change existing item `./client -c=".var/conf/client.cfg" item --label="<LABEL>" store`.
    - Select item if found more than one.
    - Enter requested values (Label, Comment, select Type, Secret and other meta data).
1. Show item `./client -c=".var/conf/client.cfg" item --label="<LABEL" show`
    - Select item if found more than one.
1. List items `./client -c=".var/conf/client.cfg" item --label="<LABEL" list`
1. Use `--offline` flag to enable offline mode
