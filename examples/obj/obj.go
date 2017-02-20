// Object orientated API example.

package main

import (
	"flag"
	"fmt"
	"github.com/smoogle/gosmart"
	"golang.org/x/net/context"
	"log"
)

var (
	flagClient    = flag.String("client", "", "OAuth Client ID")
	flagSecret    = flag.String("secret", "", "OAuth Secret")
)

func main() {
	flag.Parse()

	// No date on log messages
	log.SetFlags(0)

	ctx := context.Background()
	cfg := gosmart.Config{
		ClientID: *flagClient,
		Secret: *flagSecret,
	}
	st, err := gosmart.Connect(ctx, cfg)
	if err != nil {
		log.Fatalln(err)
	}

	for _, dev := range st.Devices {
		fmt.Printf("\nDevice ID:      %s\n", dev.ID)
		fmt.Printf("  Name:         %s\n", dev.Name)
		fmt.Printf("  Display Name: %s\n", dev.DisplayName)
		if len(dev.Attributes) > 0 {
			fmt.Printf("  Attributes:\n")
			for k, v := range dev.Attributes {
				fmt.Printf("    %v: %v\n", k, v)
			}
		}
		if len(dev.Commands) > 0 {
			fmt.Printf("  Commands:\n")
			for _, cmd := range dev.Commands {
				fmt.Printf("    %s\n", cmd)
			}
		}
		fmt.Println()
	}

	fmt.Println()
	fmt.Printf("Turning all devices on...\n")
	for _, dev := range st.Devices {
		err := dev.Call("on")
		if err != nil {
			fmt.Printf("[%v] %s: %v\n", dev.ID, dev.Name, err)
		} else {
			fmt.Printf("[%v] %s: OK\n", dev.ID, dev.Name)
		}
	}
}
