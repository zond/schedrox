package event

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type recurrenceChecker func(matches []string, start, t time.Time) bool

type rruleGenerator func(matches []string, until time.Time) string

type recurrenceType struct {
	pattern   *regexp.Regexp
	checker   recurrenceChecker
	generator rruleGenerator
}

type recurrenceParser struct {
	matches   []string
	checker   recurrenceChecker
	generator rruleGenerator
}

func (self *recurrenceParser) matchesTime(start, t time.Time) bool {
	return self.checker(self.matches, start, t)
}
func (self *recurrenceParser) rrule(until time.Time) string {
	return self.generator(self.matches, until)
}

type recurrenceTypes []recurrenceType

func (self recurrenceTypes) find(r string) *recurrenceParser {
	for _, typ := range self {
		if matches := typ.pattern.FindStringSubmatch(r); matches != nil {
			return &recurrenceParser{
				matches:   matches,
				checker:   typ.checker,
				generator: typ.generator,
			}
		}
	}
	return nil
}

var theRecurrenceTypes = recurrenceTypes{
	// Every n days
	// DAY:n
	recurrenceType{
		pattern: regexp.MustCompile("^DAY:([0-9]+)$"),
		checker: func(matches []string, start, t time.Time) bool {
			var err error
			var dayN int
			if dayN, err = strconv.Atoi(matches[1]); err != nil {
				panic(err)
			}
			diff := int(t.Sub(start) / (time.Hour * 24))
			if diff%dayN == 0 {
				return true
			}
			return false
		},
		generator: func(matches []string, until time.Time) string {
			buf := new(bytes.Buffer)
			fmt.Fprintf(buf, "FREQ=DAILY;UNTIL=%v;INTERVAL=%v", until.Format("20060102T150405Z"), matches[1])
			return string(buf.Bytes())
		},
	},
	// Every n months on the x,y,z days of the month
	// DOM:x,y,z/n
	recurrenceType{
		pattern: regexp.MustCompile("^DOM:(([0-9]{1,2})(,[0-9]{1,2})*)/([0-9]+)$"),
		generator: func(matches []string, until time.Time) string {
			buf := new(bytes.Buffer)
			fmt.Fprintf(buf, "FREQ=MONTHLY;UNTIL=%v;INTERVAL=%v;BYMONTHDAY=%v", until.Format("20060102T150405Z"), matches[4], matches[1])
			return string(buf.Bytes())
		},
		checker: func(matches []string, start, t time.Time) bool {
			var err error
			var everyM int
			if everyM, err = strconv.Atoi(matches[4]); err != nil {
				panic(err)
			}
			diff := int(t.Month() - start.Month())
			if diff < 0 {
				diff = -diff
			}
			if diff%everyM != 0 {
				return false
			}
			var dayN int
			for _, day := range strings.Split(matches[1], ",") {
				if dayN, err = strconv.Atoi(day); err != nil {
					panic(err)
				}
				if dayN == int(t.Day()) {
					return true
				}
			}
			return false
		},
	},
	// Every n weeks, on x,y,z days of the week.
	// DOW:x,y,z/n
	recurrenceType{
		pattern: regexp.MustCompile("^DOW:(([0-6])(,[0-6])*)/([0-9]+)$"),
		generator: func(matches []string, until time.Time) string {
			buf := new(bytes.Buffer)
			var days []string
			dayMap := map[string]string{
				"1": "MO",
				"2": "TU",
				"3": "WE",
				"4": "TH",
				"5": "FR",
				"6": "SA",
				"7": "SU",
			}
			for _, dayN := range strings.Split(matches[1], ",") {
				days = append(days, dayMap[dayN])
			}
			fmt.Fprintf(buf, "FREQ=WEEKLY;UNTIL=%v;INTERVAL=%v;WKST=MO;BYDAY=%v", until.Format("20060102T150405Z"), matches[4], strings.Join(days, ","))
			return string(buf.Bytes())
		},
		checker: func(matches []string, start, t time.Time) bool {
			var err error
			var everyW int
			if everyW, err = strconv.Atoi(matches[4]); err != nil {
				panic(err)
			}
			_, startWeek := start.ISOWeek()
			_, tWeek := t.ISOWeek()
			diff := tWeek - startWeek
			if diff < 0 {
				diff = -diff
			}
			if diff%everyW != 0 {
				return false
			}
			var dayN int
			for _, day := range strings.Split(matches[1], ",") {
				if dayN, err = strconv.Atoi(day); err != nil {
					panic(err)
				}
				if dayN == int(t.Weekday()) {
					return true
				}
			}
			return false
		},
	},
}
