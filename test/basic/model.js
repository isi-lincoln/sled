
client = {
  name: 'client',
  kernel: '4.15.11-kernel',
  initrd: 'initramfs',
  cmdline: 'console=ttyS1',
  defaultnic: 'e1000',
  defaultdisktype: { dev: 'sd', bus: 'sata' }
}

server = {
  name: 'server',
  image: 'fedora-27',
  cpu: { cores: 2 },
  memory: { capacity: GB(8) },
  //mounts: [{ source: env.PWD+'/../..', point: '/tmp/code' }]
  mounts: [
    { source: '/home/lthurlow/go/src/github.com/ceftb/sled/', point: '/tmp/code' },
    // test writing images (uncompressed qcow
    // qemu-img convert test.qcow2 -O raw test.img
    { source: '/home/lthurlow/sled/img/', point: '/var/img/' },
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

