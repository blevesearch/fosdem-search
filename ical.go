package main

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const iCalTimeFormat = "20060102T150405"

func parseEvents(path string) chan *Event {
	rv := make(chan *Event)

	go func() {
		defer close(rv)

		var event *Event
		var input io.ReadCloser
		if strings.HasPrefix(path, "http://") ||
			strings.HasPrefix(path, "https://") {
			log.Printf("Loading events from URL: %s", path)
			resp, err := http.Get(path)
			if err != nil {
				log.Println(err)
				return
			}
			if resp.StatusCode != 200 {
				log.Println("bad GET status for "+path, resp.Status)
				return
			}
			lastUpdated = time.Now()
			input = resp.Body
		} else {
			log.Printf("Loading events from file: %s", path)
			file, err := os.Open(path)
			if err != nil {
				log.Println(err)
				return
			}
			input = file
		}
		defer input.Close()

		var location *time.Location
		scanner := bufio.NewScanner(input)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "BEGIN") {
				event = new(Event)
			} else if strings.HasPrefix(line, "END") {
				rv <- event
				location = nil
			} else if strings.HasPrefix(line, "SUMMARY") {
				colon := strings.Index(line, ":")
				if colon > 0 {
					event.Summary = line[colon+1:]
				}
			} else if strings.HasPrefix(line, "DESCRIPTION") {
				colon := strings.Index(line, ":")
				if colon > 0 {
					desc := line[colon+1:]
					desc = strings.TrimSpace(desc)

					if strings.HasPrefix(desc, "<p>") {
						desc = desc[3:]
					}
					if strings.HasSuffix(desc, "</p>") {
						desc = desc[:len(desc)-4]
					}
					if len(desc) > 0 {
						event.Description = desc
					}
				}
			} else if strings.HasPrefix(line, "LOCATION") {
				colon := strings.Index(line, ":")
				if colon > 0 {
					location := line[colon+1:]
					location = strings.TrimSpace(location)
					event.Location = location
				}
			} else if strings.HasPrefix(line, "STATUS") {
				// ignore all CONFIRMED
			} else if strings.HasPrefix(line, "CLASS") {
				// ignore all PUBLIC
			} else if strings.HasPrefix(line, "TZID") {
				colon := strings.Index(line, ":")
				if colon > 0 {
					loc := line[colon+1:]
					loc = strings.TrimSpace(loc)
					loc = strings.Replace(loc, "-", "/", -1)
					var err error
					location, err = time.LoadLocation(loc)
					if err != nil {
						log.Printf("error loading location: %v", err)
					}
				}
			} else if strings.HasPrefix(line, "CATEGORIES") {
				colon := strings.Index(line, ":")
				if colon > 0 {
					cat := line[colon+1:]
					cat = strings.TrimSpace(cat)
					event.Category = cat
				}
			} else if strings.HasPrefix(line, "URL") {
				colon := strings.Index(line, ":")
				if colon > 0 {
					url := line[colon+1:]
					url = strings.TrimSpace(url)
					event.URL = url
				}
			} else if strings.HasPrefix(line, "METHOD") {
				// ignore all PUBLISH
			} else if strings.HasPrefix(line, "UID") {
				colon := strings.Index(line, ":")
				if colon > 0 {
					uid := line[colon+1:]
					uid = strings.TrimSpace(uid)
					event.UID = uid
				}
			} else if strings.HasPrefix(line, "DTSTART") {
				colon := strings.Index(line, ":")
				if colon > 0 {
					start := line[colon+1:]
					start = strings.TrimSpace(start)
					if location == nil {
						location = time.UTC
					}
					startTime, err := time.ParseInLocation(iCalTimeFormat, start, location)
					if err == nil {
						event.Start = startTime
					}
				}
			} else if strings.HasPrefix(line, "DTEND") {
				colon := strings.Index(line, ":")
				if colon > 0 {
					end := line[colon+1:]
					end = strings.TrimSpace(end)
					if location == nil {
						location = time.UTC
					}
					endTime, err := time.ParseInLocation(iCalTimeFormat, end, location)
					if err == nil {
						if !event.Start.IsZero() {
							duration := endTime.Sub(event.Start)
							event.Duration = duration.Minutes()
						}
					}
				}
			} else if strings.HasPrefix(line, "ATTENDEE") {
				attendeeParts := strings.Split(line, ";")
				for _, part := range attendeeParts {
					if strings.HasPrefix(part, "CN") {
						equal := strings.Index(part, "=")
						if equal > 0 {
							cn := part[equal+1:]
							cn = strings.TrimSpace(cn)
							if strings.HasSuffix(cn, "\":invalid:nomail") {
								cn = cn[:len(cn)-len("\":invalid:nomail")]
							}
							if strings.HasPrefix(cn, "\"") {
								cn = cn[1:]
							}
							event.Speaker = cn
						}
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Println(err)
			return
		}
	}()

	return rv
}
