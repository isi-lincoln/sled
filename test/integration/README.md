## Integration testing

This suite is meant to test the end-to-end components for sled.

Using a very basic single client, server model.  The server and client
are configured using ssh and telnet respectively.  The server's bolt db
is updated with the specified image for the client.  The client begins
running sledc to grab the image from the server.  The image is specific
for testing, and must include the `rvn` user and ssh keys, as it relies
on pingwait for work.


### Testing

Run `make` to build the binary to update the server's bolt db entries.
Make will also download all the necessary images to use into the images/
directory.  Next, run `make install` this will require sudo and will place
the necessary images in the correct raven directories in `/var/rvn/*`.

From this `integration` directory, make sure to be root and run `go test`.
Root is necessary in order to run rvn.  `model.js` and
`bolt/manual-update.go` use the images/ directory for the kernel,
intramfs, and client image to load (ubuntu 1604).

The default timeout is 10 minutes, make sure that if a custom timeout is
used that it does not go below 5 minutes, as this is the timeout used by
pingwait, to enforce that the client drive was wiped, image downloaded,
and properly kexec'd.

```
INFO[0000] Testing: Starting Raven configuration sledc - sledd 
INFO[0000] Building Raven Topology                      
INFO[0009] Deploying Raven Topology                     
INFO[0023] Waiting on Raven Topology                    
INFO[0034] Configuring Raven Topology                   
INFO[0043] Client: Setting MAC Address                  
INFO[0043] Client: Setting interface UP                 
INFO[0043] Client: Setting IP Address                   
INFO[0043] Server: Setting interface UP                 
INFO[0045] Server: Setting IP Address                   
INFO[0045] Server: Creating BoltDB Entry                
INFO[0046] Server: Starting Sledd                       
INFO[0048] Client: Running Sledc                        
INFO[0048] Waiting for Client to finish Kexec...        
INFO[0238] Tearing down Raven Topology                  
INFO[0238] shutting down sled-basic_client              
INFO[0238] shutting down sled-basic_server              
PASS
ok      github.com/ceftb/sled/test/integration  248.978s
```
