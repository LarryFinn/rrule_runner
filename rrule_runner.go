package main

import (
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/teambition/rrule-go"
	"strconv"
	"strings"
	"time"
)

var test = cron.New(cron.WithParser(RRuleParser{}))

type RRuleParser struct {

}

type RRuleSchedule struct {
	rule *rrule.RRule

}

func (sched RRuleSchedule) Next(nextTime time.Time) time.Time {
	return sched.rule.After(nextTime, true)
}

func numberListToSlice(numList string) (result []int, err error){
	pieces := strings.Split(numList, ",")
	result = make([]int, len(pieces))
	for index, piece := range pieces {
		num, err := strconv.Atoi(piece)
		if err != nil {
			return nil, err
		}
		result[index] = num
	}
	return result, err
}

func stringToWeekday(weekday string) (rrule.Weekday, error) {
	switch weekday {
	case "MO": return rrule.MO, nil
	case "TU": return rrule.TU, nil
	case "WE": return rrule.WE, nil
	case "TH": return rrule.TH, nil
	case "FR": return rrule.FR, nil
	case "SA": return rrule.SA, nil
	case "SU": return rrule.SU, nil
	default:
		return rrule.Weekday{}, errors.New(fmt.Sprintf("Invalid weekday %s", weekday))
	}
}

func weekdayListToSlice(weekdayList string) (result []rrule.Weekday, err error) {
	pieces := strings.Split(weekdayList, ",")
	result = make([]rrule.Weekday, len(pieces))
	for index, piece := range pieces {
		weekday, err :=  stringToWeekday(piece)
		if err != nil {
			return nil, err
		}
		result[index] = weekday
	}
	return result, err
}

func (rr RRuleParser) Parse(spec string) (cron.Schedule, error) {
	dateLayout := "20060102T030405"
	opts := rrule.ROption{}
	lines := strings.Split(spec, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "DTSTART") {
			dtStartStripped := strings.Replace(line, "DTSTART;", "", 1)
			if strings.Contains(dtStartStripped, ":") {
				tzAndDatePieces := strings.SplitN(dtStartStripped, ":", 2)
				tzIdAndKey := strings.SplitN(tzAndDatePieces[0], "=", 2)
				tzId := tzIdAndKey[1]
				location, err := time.LoadLocation(tzId)
				if err != nil {
					return nil, err
				}
				date, err := time.ParseInLocation(dateLayout, tzAndDatePieces[1], location)
				if err != nil {
					return nil, err
				}
				opts.Dtstart = date
			}

		} else if strings.HasPrefix(line, "RRULE:") {
			rruleLineStripped := strings.Replace(line, "RRULE:", "", 1)
			pieces := strings.Split(rruleLineStripped, ";")
			for _, piece := range pieces {
				keyAndValue := strings.SplitN(piece, "=", 2)
				key := keyAndValue[0]
				value := keyAndValue[1]
				switch key {
				case "FREQ":
					switch value {
					case "YEARLY": opts.Freq = rrule.YEARLY
					case "MONTHLY": opts.Freq = rrule.MONTHLY
					case "WEEKLY": opts.Freq = rrule.WEEKLY
					case "DAILY": opts.Freq = rrule.DAILY
					case "HOURLY": opts.Freq = rrule.HOURLY
					case "MINUTELY": opts.Freq = rrule.MINUTELY
					case "SECONDLY": opts.Freq = rrule.SECONDLY
					}
				case "BYHOUR":
					byHour, err := numberListToSlice(value)
					if err != nil {
						return nil, err
					}
					opts.Byhour = byHour
				case "BYMINUTE":
					byMinute, err := numberListToSlice(value)
					if err != nil {
						return nil, err
					}
					opts.Byminute = byMinute
				case "BYMONTH":
					byMonth, err := numberListToSlice(value)
					if err != nil {
						return nil, err
					}
					opts.Bymonth = byMonth
				case "BYMONTHDAY":
					byMonthDay, err := numberListToSlice(value)
					if err != nil {
						return nil, err
					}
					opts.Bymonthday = byMonthDay
				case "BYSECOND":
					bySecond, err := numberListToSlice(value)
					if err != nil {
						return nil, err
					}
					opts.Bysecond = bySecond
				case "BYSETPOS":
					bySetPos, err := numberListToSlice(value)
					if err != nil {
						return nil, err
					}
					opts.Bysetpos = bySetPos
				case "BYDAY":
					byDay, err := weekdayListToSlice(value)
					if err != nil {
						return nil, err
					}
					opts.Byweekday = byDay
				case "BYWEEKNO":
					byWeekNo, err := numberListToSlice(value)
					if err != nil {
						return nil, err
					}
					opts.Byweekno = byWeekNo
				case "BYYEARDAY":
					byYearDay, err := numberListToSlice(value)
					if err != nil {
						return nil, err
					}
					opts.Byyearday = byYearDay
				case "WKST":
					weekStart, err := stringToWeekday(value)
					if err != nil {
						return nil, err
					}
					opts.Wkst = weekStart
				case "INTERVAL":
					interval, err := strconv.Atoi(value)
					if err != nil {
						return nil, err
					}
					opts.Interval = interval
				case "UNTIL":
					date, err := time.Parse(dateLayout, value)
					if err != nil {
						return nil, err
					}
					opts.Until = date
				case "COUNT":
					count, err := strconv.Atoi(value)
					if err != nil {
						return nil, err
					}
					opts.Count = count
				}
			}

		} else {
			return nil, errors.New(fmt.Sprintf("Cannot parse RRule line %s", line))
		}
	}


	rule, err := rrule.NewRRule(opts)
	if err != nil {
		return nil, err
	}
	return RRuleSchedule{rule}, nil
}

func main() {
	sched, err := RRuleParser{}.Parse("DTSTART;TZID=America/New_York:20210905T090000\nRRULE:FREQ=DAILY;COUNT=30;INTERVAL=1;WKST=SU;BYDAY=MO,TU;BYHOUR=15;BYMINUTE=10;BYSECOND=0")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(sched.Next(time.Now()))
	}
	cronny := cron.New(cron.WithParser(RRuleParser{}))
	id, err := cronny.AddFunc("RRULE:FREQ=DAILY;COUNT=30;INTERVAL=1;WKST=SU;BYDAY=MO,TU;BYHOUR=15;BYMINUTE=08;BYSECOND=0", func() {
		fmt.Println("WE DID IT")
		fmt.Println(time.Now())
	})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(id)
		cronny.Start()
		dur, _ := time.ParseDuration("1 second")
		for {

			time.Sleep(dur)
		}
	}
}