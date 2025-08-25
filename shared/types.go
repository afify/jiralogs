package shared

import "time"

// Constants for Jira and worklog operations
const (
	RequiredHoursPerDay = 8.0
	MaxResults          = 500
	HTTPClientError     = 400
	SecondsPerHour      = 3600
	ExitCodeError       = 1
	DayNameLength       = 3
	FilePermissions     = 0666
	DateFormat          = "2006-01-02"
	WorkStartTime       = "T09:00:00.000+0300"
	WorklogDateFormat   = "2006-01-02T15:04:05.000-0700"
	ZeroHours           = 0.0
	ZeroTotal           = 0
)

type User struct {
	AccountID string `json:"accountId"`
}

type Issue struct {
	Key    string `json:"key"`
	Fields struct {
		Summary string `json:"summary"`
	} `json:"fields"`
}

type Worklog struct {
	ID     string `json:"id"`
	Author struct {
		AccountID    string `json:"accountId"`
		DisplayName  string `json:"displayName"`
		EmailAddress string `json:"emailAddress"`
	} `json:"author"`
	TimeSpentSeconds int    `json:"timeSpentSeconds"`
	TimeSpent        string `json:"timeSpent"`
	Started          string `json:"started"`
	Created          string `json:"created"`
	Updated          string `json:"updated"`
	Comment          string `json:"comment"`
}

type DayLog struct {
	Date    string
	Hours   float64
	Comment string
}

type TicketWorklog struct {
	Key     string
	Summary string
	Logs    []Worklog
	Total   float64
}

type DayStatus int

const (
	DayCompliant DayStatus = iota
	DayPartial
	DayMissing
	DayWeekend
)

type DailyReport struct {
	Date    string
	Day     string
	Hours   float64
	Status  DayStatus
	Tickets []string
}

type Summary struct {
	StartDate     string
	EndDate       string
	Workdays      int
	LoggedDays    int
	RequiredHours float64
	LoggedHours   float64
	CompliantDays []string
	PartialDays   []string
	MissingDays   []string
	TotalTickets  int
}

type Period struct {
	Start time.Time
	End   time.Time
}

type ProjectData struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	ID   string `json:"id"`
}

type JiraUser struct {
	AccountID    string `json:"accountId"`
	EmailAddress string `json:"emailAddress"`
	DisplayName  string `json:"displayName"`
}

type TicketLogData struct {
	Summary string
	Total   float64
	LastLog string
	Logs    []DayLog
}

type SummaryData struct {
	StartDate     string
	EndDate       string
	Workdays      int
	LoggedDays    int
	RequiredHours float64
	LoggedHours   float64
	CompliantDays []string
	PartialDays   []string
	MissingDays   []string
	TotalTickets  int
}
