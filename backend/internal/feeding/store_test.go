package feeding

import (
	"context"
	"testing"
	"time"
)

func TestNewScheduleReplacesOnlyItsRangeAndClearsCompletion(t *testing.T) {
	store, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	now := time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)
	start := time.Date(2027, time.January, 10, 0, 0, 0, 0, time.UTC)
	base, err := ValidatePublish(PublishInput{
		StartAt: start, TimeZone: "UTC", Times: []string{"08:00", "20:00"}, Mode: "standard",
		Standard: &Standard{AmountMode: "daily", Quantity: 2, Unit: "c"},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	baseID, err := store.Publish(ctx, base)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SetCompleted(ctx, "feeding-1-2027-01-10-0", true); err != nil {
		t.Fatal(err)
	}

	newStart := time.Date(2027, time.January, 11, 0, 0, 0, 0, time.UTC)
	newEnd := newStart
	replacement, err := ValidatePublish(PublishInput{
		StartAt: newStart, EndAt: &newEnd, TimeZone: "UTC", Times: []string{"09:00"}, Mode: "standard",
		Standard: &Standard{AmountMode: "perMeal", Quantity: 0.5, Unit: "c"},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	replacementID, err := store.Publish(ctx, replacement)
	if err != nil {
		t.Fatal(err)
	}
	if replacementID <= baseID {
		t.Fatalf("replacement ID %d should follow base ID %d", replacementID, baseID)
	}

	feedings, err := store.Feedings(ctx, start, time.Date(2027, time.January, 12, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(feedings) != 3 {
		t.Fatalf("got %d feedings, want two retained feedings and one replacement feeding", len(feedings))
	}
	if !feedings[0].Completed || feedings[0].ScheduledAt.Hour() != 8 {
		t.Fatalf("older day was not retained: %#v", feedings[0])
	}
	if feedings[2].ScheduledAt.Hour() != 9 || feedings[2].Completed || feedings[2].Items[0].Quantity != 0.5 {
		t.Fatalf("replacement day was not regenerated: %#v", feedings[2])
	}
}

func TestOpenEndedScheduleGeneratesFarFutureDays(t *testing.T) {
	store, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	start := time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)
	rule, err := ValidatePublish(PublishInput{
		StartAt: start, TimeZone: "UTC", Times: []string{"08:00"}, Mode: "standard",
		Standard: &Standard{AmountMode: "daily", Quantity: 1, Unit: "c"},
	}, time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Publish(context.Background(), rule); err != nil {
		t.Fatal(err)
	}
	from := time.Date(2036, time.June, 1, 0, 0, 0, 0, time.UTC)
	feedings, err := store.Feedings(context.Background(), from, from.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(feedings) != 1 || feedings[0].ScheduledAt.Year() != 2036 {
		t.Fatalf("open-ended schedule did not generate a far-future feeding: %#v", feedings)
	}
}
