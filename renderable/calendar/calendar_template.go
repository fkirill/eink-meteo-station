package calendar

import (
	"errors"
	"fkirill.org/eink-meteo-station/config"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"
)

type calendarDataDay struct {
	Day           int
	Visible       bool
	CurrentDay    bool
	PublicHoliday bool
	SchoolHoliday bool
	Weekend       bool
	Important     bool
	DayLegend     string
}

type calendarDataRow struct {
	WeekNum int
	Days    []calendarDataDay //always 7 elements
}

type dayHeader struct {
	Text    string
	Weekend bool
}

type calendarData struct {
	Month      string
	Year       int
	CurrentDay string
	DayHeaders []dayHeader // always 7 elements
	Rows       []calendarDataRow
	Legend     template.HTML
}

var weekdayNames = []dayHeader{
	{"Mon", false},
	{"Tue", false},
	{"Wed", false},
	{"Thu", false},
	{"Fri", false},
	{"Sat", true},
	{"Sun", true},
}

func createCalendarData(year int, month time.Month, currentDay int, specialDays []*config.SpecialDayOrInterval) (*calendarData, error) {
	if year < 1900 || year > 2100 {
		return nil, errors.New("wrong Year")
	}
	if month < time.January || month > time.December {
		return nil, errors.New("wrong Month")
	}
	if currentDay < 1 || currentDay > 31 {
		return nil, errors.New("wrong CurrentDay")
	}
	// first day of Month
	date := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	dailySpecialDays := dailySpecialDays(year, month, specialDays)
	importantDays := importantDays(dailySpecialDays)
	schoolHolidays := schoolHolidays(dailySpecialDays)
	publicHolidays := publicHolidays(dailySpecialDays)
	first := true
	calendarRow := 0
	_, weekNumber := date.ISOWeek()
	currentRow := calendarDataRow{
		WeekNum: weekNumber,
		Days:    make([]calendarDataDay, 7),
	}
	rows := []calendarDataRow{currentRow}
	for {
		if date.Weekday() == time.Monday && !first {
			calendarRow++
			_, weekNumber := date.ISOWeek()
			currentRow = calendarDataRow{
				WeekNum: weekNumber,
				Days:    make([]calendarDataDay, 7),
			}
			rows = append(rows, currentRow)
		}
		first = false
		day := date.Day()
		weekDay := date.Weekday()
		currentRow.Days[startsFromMonday(weekDay)] = calendarDataDay{
			Day:           day,
			Visible:       true,
			CurrentDay:    day == currentDay,
			PublicHoliday: publicHolidays[day],
			SchoolHoliday: schoolHolidays[day],
			Important:     importantDays[day],
			Weekend:       weekDay == time.Saturday || weekDay == time.Sunday,
		}
		date = date.AddDate(0, 0, 1)
		if date.Month() != month {
			break
		}
	}

	currentDayStr := ""
	if currentDay < 10 {
		currentDayStr = "0"
	}
	currentDayStr += strconv.Itoa(currentDay)

	legend := calendarLegend(year, month, specialDays)
	return &calendarData{
		Month:      month.String(),
		Year:       year,
		Rows:       rows,
		CurrentDay: currentDayStr,
		DayHeaders: weekdayNames,
		Legend:     template.HTML(legend),
	}, nil
}

func startsFromMonday(weekday time.Weekday) int {
	switch weekday {
	case time.Monday:
		return 0
	case time.Tuesday:
		return 1
	case time.Wednesday:
		return 2
	case time.Thursday:
		return 3
	case time.Friday:
		return 4
	case time.Saturday:
		return 5
	case time.Sunday:
		return 6
	}
	panic("wrong day of week")
}

var currentMonthContentTemplateText = `
<div class="calendar currentMonthCalendar">
  <div class="currentDayLarge">{{.CurrentDay}}</div>
  <div class="monthName">{{.Month}}</div>
  <table class="calendarTable">
    <thead>
      <tr>
        <td class="weekNumHeader">wk</td>
{{range .DayHeaders}}
        <td class="weekDayHeader{{if .Weekend}} weekDayHeaderWeekend{{end}}">{{.Text}}</td>
{{end}}
      </tr>
    </thead>
    <tbody>
{{range .Rows}}
    <tr>
      <td class="weekNumRow">{{.WeekNum}}</td>
  {{range .Days}}
      <td class="dayCell{{if .CurrentDay}} currentDay{{end}}{{if .Weekend}} weekend{{end}}{{if .PublicHoliday}} publicHoliday{{end}}{{if .SchoolHoliday}} schoolHoliday{{end}}{{if .Important}} important{{end}}">{{if .Visible}}{{.Day}}{{else}}&nbsp;{{end}}</td>
  {{end}}
    </tr>
{{end}}
    </tbody>
  </table>
</div>
<div class="calendarLegend">{{.Legend}}</div>
`

var currentMonthHtmlTemplateText = `
<html>
<head>
  <link rel="stylesheet" href="fonts.css"/>
  <link rel="stylesheet" href="calendar.css"/>
</head>
<body style="margin: 0">
` + currentMonthContentTemplateText + `
</body>
</html>
`

func renderCurrentMonth(year int, month time.Month, currentDay int, specialDays []*config.SpecialDayOrInterval, template *template.Template) (string, error) {
	data, err := createCalendarData(year, month, currentDay, specialDays)
	if err != nil {
		return "", err
	}
	sb := strings.Builder{}
	err = template.Execute(&sb, data)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

func dailySpecialDays(year int, month time.Month, specialDays []*config.SpecialDayOrInterval) []*config.SpecialDayOrInterval {
	// covers days for all months (31 day max), we don't use 0-th element, so all normal month day numbers apply
	res := make([]*config.SpecialDayOrInterval, 32, 32)
	day := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	for {
		var sdp *config.SpecialDayOrInterval = nil
		for _, sd := range specialDays {
			if sd.Type == "once_off" && sd.StartDateYear == year && sd.StartDateMonth == int(month) && sd.StartDateDay == day.Day() {
				sdp = sd
				break
			}
			if sd.Type == "annual" && sd.StartDateMonth == int(month) && sd.StartDateDay == day.Day() {
				sdp = sd
				break
			}
			if sd.Type == "interval" {
				startDate := time.Date(sd.StartDateYear, time.Month(sd.StartDateMonth), sd.StartDateDay, 0, 0, 0, 0, time.UTC)
				endDate := time.Date(sd.EndDateYear, time.Month(sd.EndDateMonth), sd.EndDateDay, 0, 0, 0, 0, time.UTC)
				if startDate == day || endDate == day || (startDate.Before(day) && endDate.After(day)) {
					sdp = sd
					break
				}
			}
		}
		res[day.Day()] = sdp
		day = day.AddDate(0, 0, 1)
		if day.Month() != month {
			break
		}
	}
	return res
}

func importantDays(dailySpecialDays []*config.SpecialDayOrInterval) []bool {
	res := make([]bool, len(dailySpecialDays), len(dailySpecialDays))
	for i := range dailySpecialDays {
		res[i] = dailySpecialDays[i] != nil && !dailySpecialDays[i].IsSchoolHoliday && !dailySpecialDays[i].IsPublicHoliday
	}
	return res
}

func schoolHolidays(dailySpecialDays []*config.SpecialDayOrInterval) []bool {
	res := make([]bool, len(dailySpecialDays), len(dailySpecialDays))
	for i := range dailySpecialDays {
		res[i] = dailySpecialDays[i] != nil && dailySpecialDays[i].IsSchoolHoliday
	}
	return res
}

func publicHolidays(dailySpecialDays []*config.SpecialDayOrInterval) []bool {
	res := make([]bool, len(dailySpecialDays), len(dailySpecialDays))
	for i := range dailySpecialDays {
		res[i] = dailySpecialDays[i] != nil && dailySpecialDays[i].IsPublicHoliday
	}
	return res
}

func calendarLegend(year int, month time.Month, specialDays []*config.SpecialDayOrInterval) string {
	res := ""
	day := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	processedIntervals := []int{}
	for {
		var sdp *config.SpecialDayOrInterval = nil
		for _, sd := range specialDays {
			if sd.Type == "once_off" && sd.StartDateYear == year && sd.StartDateMonth == int(month) && sd.StartDateDay == day.Day() {
				sdp = sd
				break
			}
			if sd.Type == "annual" && sd.StartDateMonth == int(month) && sd.StartDateDay == day.Day() {
				sdp = sd
				break
			}
			if sd.Type == "interval" {
				startDate := time.Date(sd.StartDateYear, time.Month(sd.StartDateMonth), sd.StartDateDay, 0, 0, 0, 0, time.UTC)
				endDate := time.Date(sd.StartDateYear, time.Month(sd.StartDateMonth), sd.StartDateDay, 0, 0, 0, 0, time.UTC)
				if startDate == day || endDate == day || (startDate.Before(day) && endDate.After(day)) {
					alreadyProcessed := false
					for _, intervalIndex := range processedIntervals {
						if sd.Index == intervalIndex {
							alreadyProcessed = true
							break
						}
					}
					if !alreadyProcessed {
						sdp = sd
						processedIntervals = append(processedIntervals, sd.Index)
					}
					break
				}
			}
		}
		if sdp != nil {
			if res != "" {
				res += "; "
			}
			if sdp.Type != "interval" {
				res += fmt.Sprintf("%d&nbsp;%s", day.Day(), sdp.DisplayText)
			} else {
				startDateStr := ""
				if sdp.StartDateMonth != int(month) {
					startDateStr = fmt.Sprintf("%d&nbsp;%s", sdp.StartDateDay, time.Month(sdp.StartDateMonth).String()[0:2])
				} else {
					startDateStr = strconv.Itoa(sdp.StartDateDay)
				}
				endDateStr := ""
				if sdp.EndDateMonth != int(month) {
					endDateStr = fmt.Sprintf("%d&nbsp;%s", sdp.EndDateDay, time.Month(sdp.EndDateMonth).String()[0:2])
				} else {
					endDateStr = strconv.Itoa(sdp.EndDateDay)
				}
				res += fmt.Sprintf("%s&nbsp;-&nbsp;%s&nbsp;%s", startDateStr, endDateStr, sdp.DisplayText)
			}
		}
		day = day.AddDate(0, 0, 1)
		if day.Month() != month {
			break
		}
	}
	return res
}
