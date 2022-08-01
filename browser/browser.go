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
func Book(hostFlag, dayFlag, turnFlag, emailFlag, dniFlag, adultFlag, youngFlag, kidFlag *string, showFlag *bool) error {
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

	id, secret, err := book(*hostFlag, action, *emailFlag, *dniFlag, *adultFlag, *youngFlag, *kidFlag, *showFlag)
	if err != nil {
		return err
	}
	log.Println("dni", *dniFlag, "id", id, "secret", secret)

	return nil
}

func book(host, action, email, dni, adult, young, kid string, show bool) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if show {
		ctx, cancel = chromedp.NewExecAllocator(ctx, append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", false))...)
	} else {
		ctx, cancel = chromedp.NewContext(ctx)
	}
	defer cancel()

	// create chrome instance
	ctx, cancel = chromedp.NewContext(
		ctx,
		//chromedp.WithDebugf(log.Printf),
	)
	defer cancel()

	// login
	if err := chromedp.Run(ctx,
		chromedp.Navigate(host+action),
	); err != nil {
		return "", "", err
	}
	time.Sleep(500 * time.Millisecond)
	if err := chromedp.Run(ctx,
		chromedp.SetValue(`input[name="Dni"]`, dni),
		chromedp.SetValue("#Email", email),
		chromedp.SetValue("#EmailTemp", email),
		chromedp.SetValue(`select[name="NumEntradas1"]`, adult),
		chromedp.SetValue(`select[name="NumEntradas2"]`, young),
		chromedp.SetValue(`select[name="NumEntradas3"]`, kid),
	); err != nil {
		return "", "", err
	}
	time.Sleep(500 * time.Millisecond)
	if err := chromedp.Run(ctx,
		chromedp.Click(`button[type="submit"]`),
	); err != nil {
		return "", "", err
	}
	time.Sleep(1000 * time.Millisecond)

	var card string
	if err := chromedp.Run(ctx,
		chromedp.Text(`#cardEspecial[class="col d-md-block"]`, &card),
	); err != nil {
		return "", "", err
	}
	fmt.Println(card)

	kvs := map[string]string{}
	lines := strings.Split(card, "\n")
	for _, l := range lines {
		kv := strings.SplitN(l, ":", 2)
		if len(kv) < 2 {
			continue
		}
		k := strings.Trim(kv[0], " ")
		v := strings.Trim(kv[1], " ")
		kvs[k] = v
	}

	id, ok := kvs["Código de reserva"]
	if !ok {
		return "", "", errors.New("id not found")
	}
	secret, ok := kvs["Código secreto"]
	if !ok {
		return "", "", errors.New("secret not found")
	}
	return id, secret, nil
}
