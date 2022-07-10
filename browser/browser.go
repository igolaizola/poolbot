package browser

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// go run main.go -day 20 -turn T2 -email xx@gmail.com -dni 44445555X -adult 2 -young 0 -kid 1

// Book does a reservation
func Book(hostFlag, dayFlag, turnFlag, emailFlag, dniFlag, adultFlag, youngFlag, kidFlag *string) error {
	if *hostFlag == "" {
		return errors.New("host not provided")
	}
	if *emailFlag == "" {
		return errors.New("email not provided")
	}
	if *dniFlag == "" {
		return errors.New("dni not provided")
	}
	if *dayFlag == "" {
		return errors.New("day not provided")
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Request the HTML page.
	res, err := client.Get(fmt.Sprintf("%s/piscinas", *hostFlag))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	var link string
	var day *goquery.Selection
	doc.Find(".col-sm-12.col-md-6.col-lg-4").Each(func(i int, s *goquery.Selection) {
		s.Find("h6").Each(func(ii int, ss *goquery.Selection) {
			text := strings.Trim(ss.Text(), " ")
			split := strings.Split(text, " ")
			if len(split) == 3 && split[1] == *dayFlag {
				day = s
			}
		})
	})

	if day == nil {
		return errors.New("day not found")
	}

	day.Find(".card-body").Each(func(i int, s *goquery.Selection) {
		s.Find(".text-muted").Each(func(ii int, ss *goquery.Selection) {
			text := strings.Trim(ss.Text(), " ")
			split := strings.Split(text, ":")
			if split[0] == *turnFlag {
				b := s.Find("button").First()
				if b != nil {
					link, _ = b.Attr("onclick")
				}
			}
		})
	})

	if link == "" {
		return errors.New("link not found")
	}

	link = strings.Replace(link, "location.href='", "", 1)
	link = strings.Replace(link, "';", "", 1)

	u := fmt.Sprintf("%s%s", *hostFlag, link)

	// Request the HTML page.
	res2, err := client.Get(u)
	if err != nil {
		return err
	}
	defer res2.Body.Close()
	if res2.StatusCode != 200 {
		return fmt.Errorf("status code error: %d %s", res2.StatusCode, res2.Status)
	}

	// Load the HTML document
	doc2, err := goquery.NewDocumentFromReader(res2.Body)
	if err != nil {
		return err
	}
	h, _ := doc2.Html()
	ioutil.WriteFile("doc2.html", []byte(h), 0644)

	var action string
	doc2.Find("form").Each(func(i int, s *goquery.Selection) {
		action, _ = s.Attr("action")
	})

	if action == "" {
		return errors.New("action not found")
	}
	log.Println("action", action)

	var token, turn, date string
	doc2.Find("input").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		val, _ := s.Attr("value")
		switch name {
		case "__RequestVerificationToken":
			token = val
		case "Turno":
			turn = val
		case "Fecha":
			date = val
		}
	})

	var cookies []string
	for _, cookie := range res2.Cookies() {
		cookies = append(cookies, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}
	cookie := strings.Join(cookies, "; ")

	if err := book(*hostFlag, action, token, cookie, date, turn, *emailFlag, *dniFlag, *adultFlag, *youngFlag, *kidFlag); err != nil {
		return err
	}

	return nil
}

func book(host, action, token, cookie, date, turn, email, dni, adult, young, kid string) error {
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", false))...)
	defer cancel()

	// create chrome instance
	ctx, cancel = chromedp.NewContext(
		ctx,
		//chromedp.WithDebugf(log.Printf),
	)
	defer cancel()

	// login
	err := chromedp.Run(ctx,
		chromedp.Navigate(host+action),
	)
	time.Sleep(1 * time.Second)
	chromedp.Run(ctx,
		chromedp.SetValue(`input[name="Dni"]`, dni),
		chromedp.SetValue("#Email", email),
		chromedp.SetValue("#EmailTemp", email),
		chromedp.SetValue(`select[name="NumEntradas1"]`, adult),
		chromedp.SetValue(`select[name="NumEntradas2"]`, young),
		chromedp.SetValue(`select[name="NumEntradas3"]`, kid),
	)
	time.Sleep(1 * time.Second)
	chromedp.Run(ctx,
		chromedp.Click(`button[type="submit"]`),
	)
	if err != nil {
		return err
	}
	time.Sleep(60 * time.Second)
	return nil
}
