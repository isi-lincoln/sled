
client = {
  name: 'client',
  kernel: 'bzImage-4.15.8:x86_64',
  initrd: 'sled-0.1.0:x86_64',
  cmdline: 'console=ttyS1',
  defaultnic: 'e1000',
  defaultdisktype: { dev: 'sd', bus: 'sata' }
}

server = {
  name: 'server',
  image: 'fedora-27',
  cpu: { cores: 2 },
  memory: { capacity: GB(2) },
  mounts: [{ source: env.PWD+'/../..', point: '/tmp/code' }]
}

topo = {
  name: 'sled-basic',
  nodes: [client, server],
  switches: [],
  links: [
    Link('client', 0, 'server', 0),
  ]
}

