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

	}

	fmt.Println()
	fmt.Printf("Turning all devices on...\n")
	for _, dev := range st.Devices {
		err := dev.Call("setLevel", 100)
		if err != nil {
			fmt.Printf("[%v] %s: %v\n", dev.ID, dev.Name, err)
		} else {
			fmt.Printf("[%v] %s: OK\n", dev.ID, dev.Name)
		}
	}
}
