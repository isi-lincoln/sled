package main

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/ceftb/sled"
)

func main() {

	cmdSet := &cobra.Command{
		Use:   "set [command]",
		Short: "set a command for a node",
		Long:  "set a command for a node",
		Args:  cobra.MinimumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("set: select subcommand")
		},
	}
	var server string
	cmdSet.Flags().StringVarP(&server, "server", "s", "localhost", "sled server address")

	cmdWipe := &cobra.Command{
		Use:   "wipe [mac] [disk device]",
		Short: "set the wipe command for a device",
		Long:  "set the wipe command for a device",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			wipe(server, args[0], args[1])
		},
	}

	cmdWrite := &cobra.Command{
		Use:   "write [mac] [image] [kernel] [initrd] [disk device]",
		Short: "set the write command for a device",
		Long: "set the write command for a device, " +
			"the image,kernel,initrd must be provided out of band to /var/img/" +
			"on the sled server",
		Args: cobra.MinimumNArgs(5),
		Run: func(cmd *cobra.Command, args []string) {
			write(server, args[0], args[1], args[2], args[3], args[4])
		},
	}

	cmdKexec := &cobra.Command{
		Use:   "kexec [mac] [kernel] [append] [initrd]",
		Short: "set the kexec command for a device",
		Long:  "set the kexec command for a device",
		Args:  cobra.MinimumNArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			kexec(server, args[0], args[1], args[2], args[3])
		},
	}

	var rootCmd = &cobra.Command{Use: "sledctl"}
	rootCmd.AddCommand(cmdSet)
	cmdSet.AddCommand(cmdWipe)
	cmdSet.AddCommand(cmdWrite)
	cmdSet.AddCommand(cmdKexec)

	rootCmd.Execute()

}

func wipe(server, mac, disk string) {
	conn, cli := initClient(server)
	defer conn.Close()

	_, err := cli.Update(context.TODO(), &sled.UpdateRequest{
		Mac:        mac,
		CommandSet: &sled.CommandSet{Wipe: &sled.Wipe{disk}},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func write(server, mac, image, kernel, initrd, device string) {
	conn, cli := initClient(server)
	defer conn.Close()

	_, err := cli.Update(context.TODO(), &sled.UpdateRequest{
		Mac: mac,
		CommandSet: &sled.CommandSet{Write: &sled.Write{
			ImageName:  image,
			KernelName: kernel,
			InitrdName: initrd,
			Device:     device,
		}},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func kexec(server, mac, kernel, append, initrd string) {
	conn, cli := initClient(server)
	defer conn.Close()

	_, err := cli.Update(context.TODO(), &sled.UpdateRequest{
		Mac: mac,
		CommandSet: &sled.CommandSet{Kexec: &sled.Kexec{
			Kernel: kernel,
			Append: append,
			Initrd: initrd,
		}},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func initClient(server string) (*grpc.ClientConn, sled.SledClient) {
	conn, err := grpc.Dial(server+":6000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect to sled server - %v", err)
	}

	return conn, sled.NewSledClient(conn)
}
