package feeding

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

func ValidatePublish(input PublishInput, now time.Time) (ScheduleRule, error) {
	loc, err := time.LoadLocation(input.TimeZone)
	if err != nil {
		return ScheduleRule{}, fmt.Errorf("invalid time zone")
	}
	if input.StartAt.IsZero() {
		return ScheduleRule{}, errors.New("start date is required")
	}
	startLocal := input.StartAt.In(loc)
	startDay := localMidnight(startLocal, loc)
	today := localMidnight(now.In(loc), loc)
	if startDay.Before(today) {
		return ScheduleRule{}, errors.New("start date must be today or later")
	}

	var endExclusive *time.Time
	if input.EndAt != nil {
		endLocal := input.EndAt.In(loc)
		endDay := localMidnight(endLocal, loc)
		if endDay.Before(startDay) {
			return ScheduleRule{}, errors.New("end date must not be before the start date")
		}
		exclusive := endDay.AddDate(0, 0, 1).UTC()
		endExclusive = &exclusive
	}

	if len(input.Times) < 1 || len(input.Times) > 12 {
		return ScheduleRule{}, errors.New("choose between 1 and 12 feedings per day")
	}
	times := append([]string(nil), input.Times...)
	lastMinutes := -1
	for _, clock := range times {
		parsed, err := time.Parse("15:04", clock)
		if err != nil {
			return ScheduleRule{}, errors.New("feeding times must use HH:mm format")
		}
		minutes := parsed.Hour()*60 + parsed.Minute()
		if minutes <= lastMinutes {
			return ScheduleRule{}, errors.New("feeding times must be in increasing order")
		}
		lastMinutes = minutes
	}

	rule := ScheduleRule{
		StartAt:        startDay.UTC(),
		EndAtExclusive: endExclusive,
		TimeZone:       input.TimeZone,
		Times:          times,
		Mode:           input.Mode,
	}
	switch input.Mode {
	case "standard":
		if input.Standard == nil || input.Weaning != nil {
			return ScheduleRule{}, errors.New("standard food amounts are required")
		}
		standard := *input.Standard
		if standard.AmountMode != "daily" && standard.AmountMode != "perMeal" {
			return ScheduleRule{}, errors.New("amount mode must be daily or perMeal")
		}
		if !validAmount(standard.Quantity) || strings.TrimSpace(standard.Unit) == "" {
			return ScheduleRule{}, errors.New("food quantity and unit are required")
		}
		standard.Unit = strings.TrimSpace(standard.Unit)
		rule.Standard = &standard
	case "weaning":
		if input.Weaning == nil || input.Standard != nil {
			return ScheduleRule{}, errors.New("weaning food amounts are required")
		}
		if endExclusive == nil {
			return ScheduleRule{}, errors.New("weaning requires an end date")
		}
		weaning := *input.Weaning
		weaning.OldFood = strings.TrimSpace(weaning.OldFood)
		weaning.NewFood = strings.TrimSpace(weaning.NewFood)
		weaning.OldUnit = strings.TrimSpace(weaning.OldUnit)
		weaning.NewUnit = strings.TrimSpace(weaning.NewUnit)
		if weaning.OldFood == "" || weaning.NewFood == "" ||
			!validAmount(weaning.OldQuantity) || !validAmount(weaning.NewQuantity) ||
			weaning.OldUnit == "" || weaning.NewUnit == "" {
			return ScheduleRule{}, errors.New("food names, quantities, and units are required")
		}
		rule.Weaning = &weaning
	default:
		return ScheduleRule{}, errors.New("mode must be standard or weaning")
	}

	if rule.EndAtExclusive != nil {
		end := rule.EndAtExclusive.In(loc)
		// Calendar-day arithmetic is used below; this bound prevents accidental huge inserts.
		calendarDays := 0
		for day := startDay; day.Before(end); day = day.AddDate(0, 0, 1) {
			calendarDays++
			if calendarDays > 3660 {
				return ScheduleRule{}, errors.New("date range cannot exceed 10 years")
			}
		}
	}

	return rule, nil
}

func BuildDay(rule ScheduleRule, day time.Time, scheduleID int64) ([]Feeding, error) {
	loc, err := time.LoadLocation(rule.TimeZone)
	if err != nil {
		return nil, err
	}
	localDay := localMidnight(day.In(loc), loc)
	phase, oldPercent, newPercent := 0, 0.0, 0.0
	if rule.Mode == "weaning" {
		if rule.EndAtExclusive == nil || rule.Weaning == nil {
			return nil, errors.New("invalid weaning rule")
		}
		start := rule.StartAt.In(loc)
		end := rule.EndAtExclusive.In(loc)
		totalDays := calendarDayDifference(start, end, loc)
		dayIndex := calendarDayDifference(start, localDay, loc)
		phase = weaningPhase(dayIndex, totalDays)
		switch phase {
		case 0:
			oldPercent, newPercent = .75, .25
		case 1:
			oldPercent, newPercent = .50, .50
		case 2:
			oldPercent, newPercent = .25, .75
		default:
			return nil, errors.New("weaning day outside configured range")
		}
	}

	feedings := make([]Feeding, 0, len(rule.Times))
	for i, clock := range rule.Times {
		parsed, _ := time.Parse("15:04", clock)
		scheduled := time.Date(localDay.Year(), localDay.Month(), localDay.Day(), parsed.Hour(), parsed.Minute(), 0, 0, loc)
		feeding := Feeding{
			ID:          fmt.Sprintf("feeding-%d-%s-%d", scheduleID, localDay.Format("2006-01-02"), i),
			ScheduledAt: scheduled.UTC(),
			Items:       make([]FeedingItem, 0, 2),
		}
		switch rule.Mode {
		case "standard":
			quantity := rule.Standard.Quantity
			if rule.Standard.AmountMode == "daily" {
				quantity /= float64(len(rule.Times))
			}
			feeding.Items = append(feeding.Items, FeedingItem{Quantity: rounded(quantity), Unit: rule.Standard.Unit})
		case "weaning":
			feeding.Phase = []string{"75:25", "50:50", "25:75"}[phase]
			feeding.Items = append(feeding.Items,
				FeedingItem{Name: rule.Weaning.OldFood, Quantity: rounded(rule.Weaning.OldQuantity * oldPercent / float64(len(rule.Times))), Unit: rule.Weaning.OldUnit},
				FeedingItem{Name: rule.Weaning.NewFood, Quantity: rounded(rule.Weaning.NewQuantity * newPercent / float64(len(rule.Times))), Unit: rule.Weaning.NewUnit},
			)
		}
		feedings = append(feedings, feeding)
	}
	sort.Slice(feedings, func(i, j int) bool { return feedings[i].ScheduledAt.Before(feedings[j].ScheduledAt) })
	return feedings, nil
}

func weaningPhase(dayIndex, totalDays int) int {
	base, remainder := totalDays/3, totalDays%3
	lengths := [3]int{base, base, base}
	for i := 0; i < remainder; i++ {
		lengths[i]++
	}
	cumulative := 0
	for phase, length := range lengths {
		cumulative += length
		if dayIndex < cumulative {
			return phase
		}
	}
	return -1
}

func validAmount(quantity float64) bool {
	return quantity > 0 && !math.IsNaN(quantity) && !math.IsInf(quantity, 0)
}

func rounded(value float64) float64 {
	return math.Round(value*1000) / 1000
}

func localMidnight(value time.Time, loc *time.Location) time.Time {
	year, month, day := value.In(loc).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}

func calendarDayDifference(start, end time.Time, loc *time.Location) int {
	start = localMidnight(start, loc)
	end = localMidnight(end, loc)
	days := 0
	for current := start; current.Before(end); current = current.AddDate(0, 0, 1) {
		days++
	}
	return days
}
