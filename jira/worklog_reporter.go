package jira

import "LogS/shared"

func (w *WorklogService) GenerateDailyReports(period shared.Period, dailyHours map[string]float64, dailyTickets map[string][]string) []shared.DailyReport {
	var reports []shared.DailyReport

	for d := period.Start; !d.After(period.End); d = d.AddDate(0, 0, 1) {
		date := d.Format(shared.DateFormat)
		day := d.Weekday().String()[:shared.DayNameLength]

		report := shared.DailyReport{
			Date:    date,
			Day:     day,
			Hours:   dailyHours[date],
			Tickets: dailyTickets[date],
		}

		if w.IsWeekend(d) {
			report.Status = shared.DayWeekend
		} else if w.IsLeaveDay(d) {
			report.Status = shared.DayLeave
		} else if report.Hours >= shared.RequiredHoursPerDay {
			report.Status = shared.DayCompliant
		} else if report.Hours > shared.ZeroHours {
			report.Status = shared.DayPartial
		} else {
			report.Status = shared.DayMissing
		}

		reports = append(reports, report)
	}

	return reports
}

func (w *WorklogService) GenerateSummary(period shared.Period, reports []shared.DailyReport, ticketLogs map[string]*shared.TicketWorklog) shared.Summary {
	summary := shared.Summary{
		StartDate:    period.Start.Format(shared.DateFormat),
		EndDate:      period.End.Format(shared.DateFormat),
		TotalTickets: len(ticketLogs),
	}

	for _, report := range reports {
		if report.Status == shared.DayWeekend || report.Status == shared.DayLeave {
			continue
		}

		summary.Workdays++
		summary.LoggedHours += report.Hours

		switch report.Status {
		case shared.DayCompliant:
			summary.CompliantDays = append(summary.CompliantDays, report.Date)
			summary.LoggedDays++
		case shared.DayPartial:
			summary.PartialDays = append(summary.PartialDays, report.Date)
			summary.LoggedDays++
		case shared.DayMissing:
			summary.MissingDays = append(summary.MissingDays, report.Date)
		}
	}

	summary.RequiredHours = float64(summary.Workdays) * shared.RequiredHoursPerDay

	return summary
}
