package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/igolaizola/poolbot/browser"
)

func main() {
	hostFlag := flag.String("host", "", "")
	dayFlag := flag.String("day", "", "")
	turnFlag := flag.String("turn", "T2", "")

	emailFlag := flag.String("email", "", "")
	dniFlag := flag.String("dni", "", "")
	adultFlag := flag.String("adult", "2", "")
	youngFlag := flag.String("young", "0", "")
	kidFlag := flag.String("kid", "1", "")
	timeoutFlag := flag.Duration("timeout", 0, "")
	showFlag := flag.Bool("show", false, "")

	flag.Parse()

	ctx := context.Background()
	if *timeoutFlag > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeoutFlag)
		defer cancel()
	}
	var prev string
	for {
		err := browser.Book(hostFlag, dayFlag, turnFlag, emailFlag, dniFlag, adultFlag, youngFlag, kidFlag, showFlag)
		switch {
		case err == nil:
			fmt.Println("Reservado!")
			return
		case strings.Contains(err.Error(), "stream error: stream ID"):
		default:
			if err.Error() != prev {
				log.Println(err)
			}
			prev = err.Error()
		}

		select {
		case <-time.After(5 * time.Second):
			continue
		case <-ctx.Done():
		}
		fmt.Println("Timeout")
		return
	}
}
