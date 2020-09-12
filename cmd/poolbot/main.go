package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	pool "github.com/igolaizola/poolbot"
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
	flag.Parse()

	for {
		err := pool.Book(hostFlag, dayFlag, turnFlag, emailFlag, dniFlag, adultFlag, youngFlag, kidFlag)
		if err == nil {
			fmt.Println("Reservado!")
			return
		}
		log.Println(err)
		<-time.After(5 * time.Second)
	}
}
