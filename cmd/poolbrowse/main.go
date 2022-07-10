package main

import (
	"context"
	"flag"
	"fmt"
	"log"
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
	flag.Parse()

	ctx := context.Background()
	if *timeoutFlag > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeoutFlag)
		defer cancel()
	}
	var prev string
	for {
		err := browser.Book(hostFlag, dayFlag, turnFlag, emailFlag, dniFlag, adultFlag, youngFlag, kidFlag)
		if err == nil {
			fmt.Println("Reservado!")
			return
		}
		if err.Error() != prev {
			log.Println(err)
		}
		prev = err.Error()
		select {
		case <-time.After(5 * time.Second):
			continue
		case <-ctx.Done():
		}
		fmt.Println("Timeout")
		return
	}
}
