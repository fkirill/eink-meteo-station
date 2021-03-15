package calendar

import (
	"errors"
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

func createCalendarData(year int, month time.Month, currentDay int) (*calendarData, error) {
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
			PublicHoliday: false,
			SchoolHoliday: false,
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

	return &calendarData{
		Month:      month.String(),
		Year:       year,
		Rows:       rows,
		CurrentDay: currentDayStr,
		DayHeaders: weekdayNames,
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

var currentMonthHtmlTemplate *template.Template

var templatesInitialized = false

func initTemplates() error {
	if templatesInitialized {
		return nil
	}
	err := error(nil)
	currentMonthHtmlTemplate, err = template.New("currentMonthHtml").Parse(currentMonthHtmlTemplateText)
	if err != nil {
		return err
	}
	templatesInitialized = true
	return nil
}

func renderCurrentMonth(year int, month time.Month, currentDay int, template *template.Template) (string, error) {
	data, err := createCalendarData(year, month, currentDay)
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

func RenderCurrentMonthHtml(year int, month time.Month, currentDay int) (string, error) {
	err := initTemplates()
	if err != nil {
		return "", err
	}
	return renderCurrentMonth(year, month, currentDay, currentMonthHtmlTemplate)
}
