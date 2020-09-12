package pool

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("couldn't create cookie jar: %w", err)
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
		Jar:     cookieJar,
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

	url := fmt.Sprintf("%s%s", *hostFlag, link)
	fmt.Println("url", url)

	// Request the HTML page.
	res2, err := client.Get(url)
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

	var action string
	doc2.Find("form").Each(func(i int, s *goquery.Selection) {
		action, _ = s.Attr("action")
	})

	if action == "" {
		return errors.New("action not found")
	}
	fmt.Println("action", action)

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	doc2.Find("input").Each(func(i int, s *goquery.Selection) {
		typ, _ := s.Attr("type")
		if typ == "hidden" {
			name, _ := s.Attr("name")
			val, _ := s.Attr("value")
			_ = writer.WriteField(name, val)
		}
	})

	url = fmt.Sprintf("%s%s", *hostFlag, action)
	method := "POST"

	_ = writer.WriteField("Dni", *dniFlag)
	_ = writer.WriteField("Email", *emailFlag)
	_ = writer.WriteField("EmailTemp", *emailFlag)
	_ = writer.WriteField("NumEntradas1", *adultFlag)
	_ = writer.WriteField("NumEntradas2", *youngFlag)
	_ = writer.WriteField("NumEntradas3", *kidFlag)
	err = writer.Close()
	if err != nil {
		fmt.Println(err)
	}

	req, err := http.NewRequest(method, url, payload)
	req.Host = *hostFlag
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res3, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res3.Body.Close()
	if res3.StatusCode != 200 {
		return fmt.Errorf("status code error: %d %s", res3.StatusCode, res3.Status)
	}

	// Load the HTML document
	doc3, err := goquery.NewDocumentFromReader(res3.Body)
	if err != nil {
		return err
	}

	var spans []string
	doc3.Find("#contEntradas").Each(func(i int, card *goquery.Selection) {
		spans = append(spans, card.Text())
	})
	if len(spans) < 2 {
		return fmt.Errorf("codes not found: %d", len(spans))
	}
	fmt.Printf("Código de reserva: %s\n", spans[len(spans)-2])
	fmt.Printf("Código secreto: %s\n", spans[len(spans)-1])
	return nil
}
