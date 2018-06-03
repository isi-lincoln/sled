
client = {
  name: 'client',
  kernel: '4.15.11-kernel',
  initrd: 'initramfs',
  cmdline: 'console=ttyS1',
  defaultnic: 'e1000',
  memory: { capacity: GB(4) }, // currently necessary
  defaultdisktype: { dev: 'sd', bus: 'sata' }
}

server = {
  name: 'server',
  image: 'fedora-27',
  cpu: { cores: 2 },
  memory: { capacity: GB(16) },
  mounts: [
    // where code resides
    { source: env.PWD+'/../../', point: '/tmp/code' },
    // where the kernel and initramfs reside
    { source: env.PWD+'images/', point: '/var/img/' },
  ]
}

topo = {
  name: 'sled-basic',
  nodes: [client, server],
  switches: [],
  links: [
    Link('client', 0, 'server', 0),
  ]
}

