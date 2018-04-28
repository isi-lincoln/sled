Testing:

go build bolt-update.go
git submodule update --init


If individually testing:

sudo -E rvn build
sudo -E rvn deploy
sudo -E rvn pingwait server
sudo -E rvn configure

go build telnet-client.go
./telnet-client

ssh -o StrictHostKeyChecking=no -i /var/rvn/ssh/rvn rvn@/tmp/code/test/integration/bolt-update
