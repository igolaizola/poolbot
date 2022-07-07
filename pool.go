package pool

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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

	res3, err := book(*hostFlag, action, token, cookie, date, turn, *emailFlag, *dniFlag, *adultFlag, *youngFlag, *kidFlag)
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
	h, _ = doc3.Html()
	ioutil.WriteFile("doc3.html", []byte(h), 0644)

	title := doc3.Find("title").First()
	if title != nil && strings.Contains(title.Text(), "bloqueada") {
		return errors.New("request blocked")
	}

	var spans []string
	doc3.Find("#contEntradas").Each(func(i int, card *goquery.Selection) {
		spans = append(spans, card.Text())
	})
	if len(spans) >= 2 {
		fmt.Printf("Código de reserva: %s\n", spans[len(spans)-2])
		fmt.Printf("Código secreto: %s\n", spans[len(spans)-1])
	}

	return nil
}

func book(host, action, token, cookie, date, turn, email, dni, adult, young, kid string) (*http.Response, error) {
	params := url.Values{}
	params.Add("__RequestVerificationToken", token)
	params.Add("Fecha", date)
	params.Add("Turno", turn)
	params.Add("Descriturno", `T1: MAÑANA`)
	params.Add("PrecioEntrada1", `2,2`)
	params.Add("PrecioEntrada2", `1,8`)
	params.Add("PrecioEntrada3", `0`)
	params.Add("TotalEntradas", ``)
	params.Add("ImporteEntradas", ``)
	params.Add("NumeroMaximoReservas", `6`)
	params.Add("Hora_modif_mn", `13:00:00`)
	params.Add("Hora_modif_tarde", `18:15:00`)
	params.Add("Descriturno_mn", `T1: MAÑANA`)
	params.Add("Descriturno_tarde", `T2: TARDE`)
	params.Add("Dni", dni)
	params.Add("Email", email)
	params.Add("EmailTemp", email)
	params.Add("NumEntradas1", adult)
	params.Add("NumEntradas2", young)
	params.Add("NumEntradas3", kid)
	body := strings.NewReader(params.Encode())

	u := host + action
	req, err := http.NewRequest("POST", host+action, body)
	if err != nil {
		return nil, err
	}
	parsed, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authority", parsed.Host)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Accept-Language", "es-ES,es;q=0.9")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Origin", host)
	req.Header.Set("Referer", u)
	req.Header.Set("Sec-Ch-Ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"102\", \"Google Chrome\";v=\"102\"")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", "\"Windows\"")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Safari/537.36")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
