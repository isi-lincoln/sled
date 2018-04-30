Testing:

Requirements:
1. sledd must exist in the top level directory
make build/sledd
2. bolt update script must exist
cd bolt
go build bolt-update.go

// this may be outdated soon.
git submodule update --init


// to test mannually (client, or server)

sudo -E rvn build
sudo -E rvn deploy
sudo -E rvn pingwait server
sudo -E rvn configure

// client code, need to change package to main first
cd client
go build telnet-client.go
./telnet-client

// server code, need to change package to main first
cd server
go build ssh-server.go
./ssh-server
