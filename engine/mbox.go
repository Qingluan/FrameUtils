package engine

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/mail"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type Mbox struct {
	tableName     string
	fillterheader string
	buf           io.ReadCloser
	count         int
	Mails         []*mail.Message
}

func (mbox *Mbox) emailSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// A single blank line and a "From " separates a message
	// https://en.wikipedia.org/wiki/Mbox#Family
	if i := strings.Index(string(data), "\n\nFrom "); i >= 0 {
		return i + 1, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return
}

func (mbox *Mbox) readEmail(b []byte) {
	// To properly read a mail message, we need to remove any preceeding
	// newlines and additionally remove the "From " line
	const NL = "\n"
	trimmed := strings.TrimLeft(string(b), NL)
	var msgString string
	if strings.Index(trimmed, "From ") == 0 {
		msgString = strings.Join(strings.Split(trimmed, NL)[1:], NL)
	} else {
		msgString = trimmed
	}

	msg, err := mail.ReadMessage(strings.NewReader(msgString))
	if err != nil {
		fmt.Println(err)
		return
	}
	mbox.Mails = append(mbox.Mails, msg)
	// fmt.Println("From:", msg.Header.Get("From"))
}

func (mbox *Mbox) readEmail2(b []byte) *mail.Message {
	// To properly read a mail message, we need to remove any preceeding
	// newlines and additionally remove the "From " line
	const NL = "\n"
	trimmed := strings.TrimLeft(string(b), NL)
	var msgString string
	if strings.Index(trimmed, "From ") == 0 {
		msgString = strings.Join(strings.Split(trimmed, NL)[1:], NL)
	} else {
		msgString = trimmed
	}

	msg, err := mail.ReadMessage(strings.NewReader(msgString))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	mbox.Mails = append(mbox.Mails, msg)
	return msg
	// fmt.Println("From:", msg.Header.Get("From"))
}

func (mbox *Mbox) emailScanner() {
	scanner := bufio.NewScanner(mbox.buf)

	// Allow a maximum of 2^24 bytes per message
	scanner.Buffer([]byte{}, 1<<24)
	scanner.Split(mbox.emailSplit)

	mbox.count = 0
	for scanner.Scan() {
		mbox.count++
		mbox.readEmail(scanner.Bytes())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading input:", err)
	}

	// fmt.Println("Total emails:", count)
}

func (mbox *Mbox) emailScanner2() {
	s := bufio.NewScanner(mbox.buf)

	var (
		msg   []byte
		count int
	)
	for s.Scan() {
		if strings.HasPrefix(s.Text(), "From ") {
			if msg == nil {
				// At the top of the file, there was no previous
				// message to zero out and process
			} else {
				count++
				mbox.readEmail(msg)
				msg = nil
			}
		} else {
			msg = append(msg, []byte("\n")...)
			msg = append(msg, s.Bytes()...)
		}
	}
	mbox.count++
	mbox.readEmail(msg)

	// fmt.Println("Total emails:", count)
}
func (mbox *Mbox) Count() int {
	return mbox.count
}
func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func parseMime(raw string) string {
	dec := new(mime.WordDecoder)
	dec.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		switch strings.ToLower(charset) {
		case "gb2312":
			//例如假字符集。
			//实际使用将与包等集成
			//作为code.google.com/p/go-charset
			content, err := ioutil.ReadAll(input)
			if err != nil {
				return nil, err
			}
			// reader, _ := charset.NewReaderLabel("gb2312", input)
			// strBytes, _ = ioutil.ReadAll(reader)
			// return bytes.NewReader(bytes.ToUpper(content)), nil
			utfbytes, err := GbkToUtf8(content)
			if err != nil {
				return nil, err
			}
			return bytes.NewReader(utfbytes), nil
		default:
			return nil, fmt.Errorf("unhandled charset %q", charset)
		}
	}
	var err error
	for {
		start := strings.Index(raw, "=?")
		// if len(raw)
		if start < 0 {
			break
		}
		end := strings.Index(raw[start+10:], "?=") + start + 10 + 2
		// if end < start {
		// 	break
		// }
		if start >= 0 && end > 0 {
			replaceStr, err := dec.Decode(raw[start:end])
			if err != nil {
				log.Println("142:", start, end, color.New(color.FgRed).Sprint(raw[start:end]), err)
				log.Println("all : ", color.New(color.FgYellow).Sprint(raw))
			}
			raw = strings.Replace(raw, raw[start:end], replaceStr, 1)
		} else {
			break
		}
		// raw, err = dec.Decode(raw)
		if err != nil {
			log.Println("150:", err)
		}
	}
	return raw

	// header, err := dec.Decode(raw)
	// if err != nil {
	// 	log.Println("150:", color.New(color.FgRed).Sprint(err),)
	// }
	// return header

	// if strings.HasPrefix(raw, "=?utf-8") {
	// 	pre := strings.TrimLeft(raw, "=?utf-8")
	// 	if strings.HasPrefix(pre, "?Q?") {

	// 	} else if strings.HasPrefix(pre, "?B?") {
	// 		real, err := base64.StdEncoding.DecodeString(pre[3:])
	// 		if err != nil {
	// 			log.Println(err)
	// 		}
	// 		return string(real)
	// 	} else {

	// 	}
	// } else {
	// 	return raw
	// }

}

func (self *Mbox) Iter(header ...string) <-chan Line {
	ch := make(chan Line)
	if header != nil {
		self.fillterheader = header[0]
	}
	var err error
	self.buf, err = os.Open(self.tableName)
	// obj := bufio.NewScanner(mbox.buf)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ch
	}
	if len(self.Mails) == 0 {
		self.emailScanner2()
	}
	go func() {
		c := 0

		name := filepath.Base(self.tableName)
		for _, msg := range self.Mails {
			header := msg.Header
			frraw := header.Get("From")
			toraw := header.Get("To")
			dateraw := header.Get("Date")
			subraw := header.Get("Subject")
			body, err := ioutil.ReadAll(msg.Body)
			fr, err := mail.ParseAddress(frraw)
			tos, err := mail.ParseAddressList(toraw)
			// to, err := mail.ParseAddress(toraw)
			// realTo := parseMime(to.String())
			date, err := mail.ParseDate(dateraw)
			sub := parseMime(subraw)
			tostrs := []string{}
			for _, to := range tos {
				toreal := to.String()
				if strings.Contains(toreal, "=?") && strings.Contains(toreal, "?=") {
					toreal = parseMime(toreal)
				}
				tostrs = append(tostrs, toreal)
				// fmt.Println(color.New(color.FgBlue).Sprint(toreal))

			}
			line := Line{name, fr.String(), strings.Join(tostrs, " "), date.String(), sub}

			if err == nil {
				line.Push(string(body))
			}
			// line := strings.TrimSpace(self.obj.Text())
			// l := Line(strings.Fields(line))
			// ch <- append(Line{self.tableName
			ch <- line
			c++
		}
		close(ch)
	}()
	return ch
}

func (self *Mbox) GetHead(k string) Line {
	name := filepath.Base(self.tableName)
	return Line{name, "From", "To", "Date", "Subject", "Body"}
}

func (self *Mbox) Close() error {
	if self.buf != nil {
		self.buf.Close()
	}
	return nil
}
func (self *Mbox) header(i ...int) (l Line) {
	headers := self.GetHead("")
	if i != nil {
		l = append(l, headers[i[0]])
		return
	}
	l = headers
	return
}

func (s *Mbox) Tp() string {
	return "mbox"
}

// func main() {
// 	if len(os.Args) != 2 {
// 		log.Fatalln("Usage:", os.Args[0], "<filename>")
// 	}

// 	filename := os.Args[1]
// 	f, err := os.Open(filename)
// 	if err != nil {
// 		log.Fatalln("Unable to open file:", err)
// 	}
// 	defer f.Close()

// 	// emailScanner(f)
// 	emailScanner2(f)

// }
