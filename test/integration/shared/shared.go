package shared

import "fmt"

var ClientIface string = "eth1"
var ServerIface string = "eth1"

var ClientIP string = "10.0.0.2"
var ServerIP string = "10.0.0.1"

var ClientMAC string = "00:00:00:00:00:01"

var LocalBoltPath string = "bolt/manual-update"
var BoltDBPath string = fmt.Sprintf("/tmp/code/test/integration/%s", LocalBoltPath)
var SleddPath string = "/tmp/code/build/sledd"
