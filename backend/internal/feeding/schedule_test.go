package feeding

import (
	"testing"
	"time"
)

func TestWeaningPhaseAssignsRemainderToEarlierPhases(t *testing.T) {
	tests := []struct {
		days int
		want []int
	}{
		{days: 1, want: []int{0}},
		{days: 2, want: []int{0, 1}},
		{days: 4, want: []int{0, 0, 1, 2}},
		{days: 5, want: []int{0, 0, 1, 1, 2}},
		{days: 10, want: []int{0, 0, 0, 0, 1, 1, 1, 2, 2, 2}},
	}
	for _, test := range tests {
		for day, expected := range test.want {
			if got := weaningPhase(day, test.days); got != expected {
				t.Fatalf("weaningPhase(%d, %d) = %d, want %d", day, test.days, got, expected)
			}
		}
	}
}

func TestBuildDaySplitsDailyWeaningAmountsAcrossMeals(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2027, time.April, 1, 0, 0, 0, 0, loc)
	end := time.Date(2027, time.April, 5, 0, 0, 0, 0, loc)
	rule, err := ValidatePublish(PublishInput{
		StartAt:  start.UTC(),
		EndAt:    &end,
		TimeZone: loc.String(),
		Times:    []string{"07:00", "13:00", "19:00"},
		Mode:     "weaning",
		Weaning: &Weaning{
			OldFood: "Old", NewFood: "New",
			OldQuantity: 1, OldUnit: "c",
			NewQuantity: 60, NewUnit: "g",
		},
	}, time.Date(2027, time.March, 1, 12, 0, 0, 0, loc))
	if err != nil {
		t.Fatal(err)
	}

	dayOne, err := BuildDay(rule, start, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(dayOne) != 3 || dayOne[0].Phase != "75:25" {
		t.Fatalf("unexpected first phase: %#v", dayOne)
	}
	if dayOne[0].Items[0].Quantity != 0.25 || dayOne[0].Items[1].Quantity != 5 {
		t.Fatalf("unexpected first feeding amounts: %#v", dayOne[0].Items)
	}
	if dayOne[0].Items[0].Unit != "c" || dayOne[0].Items[1].Unit != "g" {
		t.Fatalf("food units were not preserved: %#v", dayOne[0].Items)
	}

	lastDay, err := BuildDay(rule, start.AddDate(0, 0, 4), 4)
	if err != nil {
		t.Fatal(err)
	}
	if lastDay[0].Phase != "25:75" || lastDay[0].Items[0].Quantity != 0.083 || lastDay[0].Items[1].Quantity != 15 {
		t.Fatalf("unexpected final phase amounts: %#v", lastDay[0])
	}
}

func TestValidatePublishRequiresTodayOrLaterAndWeaningEnd(t *testing.T) {
	now := time.Date(2027, time.January, 10, 12, 0, 0, 0, time.UTC)
	input := PublishInput{
		StartAt:  time.Date(2027, time.January, 9, 0, 0, 0, 0, time.UTC),
		TimeZone: "UTC",
		Times:    []string{"07:00"},
		Mode:     "standard",
		Standard: &Standard{AmountMode: "perMeal", Quantity: 1, Unit: "c"},
	}
	if _, err := ValidatePublish(input, now); err == nil {
		t.Fatal("expected past start date to be rejected")
	}

	input.StartAt = time.Date(2027, time.January, 10, 0, 0, 0, 0, time.UTC)
	input.Mode = "weaning"
	input.Standard = nil
	input.Weaning = &Weaning{OldFood: "Old", NewFood: "New", OldQuantity: 1, OldUnit: "c", NewQuantity: 1, NewUnit: "c"}
	if _, err := ValidatePublish(input, now); err == nil {
		t.Fatal("expected weaning without end date to be rejected")
	}
}
