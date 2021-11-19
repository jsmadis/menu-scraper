package pkg

import (
	"bufio"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/htmlindex"
	"io"
)

// source: https://github.com/PuerkitoBio/goquery/wiki/Tips-and-tricks#handle-non-utf8-html-pages
func detectContentCharset(body io.Reader) string {
	r := bufio.NewReader(body)
	if data, err := r.Peek(1024); err == nil {
		_, name, _ := charset.DetermineEncoding(data, "")
		return name
	}
	return "utf-8"
}

// DecodeHTMLBody returns an decoding reader of the html Body for the specified `charset`
// If `charset` is empty, DecodeHTMLBody tries to guess the encoding from the content
// source: https://github.com/PuerkitoBio/goquery/wiki/Tips-and-tricks#handle-non-utf8-html-pages
// This is necessary because menicka.cz doesn't encode page in utf-8.
func DecodeHTMLBody(body io.Reader) (io.Reader, error) {
	charSet := detectContentCharset(body)
	e, err := htmlindex.Get(charSet)
	if err != nil {
		return nil, err
	}
	if name, _ := htmlindex.Name(e); name != "utf-8" {
		body = e.NewDecoder().Reader(body)
	}
	return body, nil
}
