
client = {
  name: 'client',
  kernel: '4.14.32-kernel',
  initrd: 'linc-initramfs',
  cmdline: 'console=ttyS1',
  defaultnic: 'e1000',
  memory: { capacity: GB(1) }, // currently necessary
  defaultdisktype: { dev: 'sd', bus: 'sata' }
}

server = {
  name: 'server',
  image: 'fedora-27',
  cpu: { cores: 2 },
  memory: { capacity: GB(2) },
  mounts: [
    // where code resides
    { source: env.PWD+'/../../', point: '/tmp/code' },
    // where the kernel and initramfs reside
    { source: env.PWD+'/images/', point: '/var/img/' },
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

