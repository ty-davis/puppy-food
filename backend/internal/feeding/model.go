package feeding

import "time"

type Amount struct {
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

type FeedingItem struct {
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

type Feeding struct {
	ID          string        `json:"id"`
	ScheduledAt time.Time     `json:"scheduledAt"`
	Completed   bool          `json:"completed"`
	Items       []FeedingItem `json:"items"`
	Phase       string        `json:"phase,omitempty"`
}

type FeedingList struct {
	Feedings []Feeding `json:"feedings"`
}

type PublishInput struct {
	StartAt  time.Time  `json:"startAt"`
	EndAt    *time.Time `json:"endAt,omitempty"`
	TimeZone string     `json:"timeZone"`
	Times    []string   `json:"times"`
	Mode     string     `json:"mode"`
	Standard *Standard  `json:"standard,omitempty"`
	Weaning  *Weaning   `json:"weaning,omitempty"`
}

type Standard struct {
	AmountMode string  `json:"amountMode"`
	Quantity   float64 `json:"quantity"`
	Unit       string  `json:"unit"`
}

type Weaning struct {
	OldFood     string  `json:"oldFood"`
	NewFood     string  `json:"newFood"`
	OldQuantity float64 `json:"oldQuantity"`
	OldUnit     string  `json:"oldUnit"`
	NewQuantity float64 `json:"newQuantity"`
	NewUnit     string  `json:"newUnit"`
}

type PublishResult struct {
	ScheduleID int64 `json:"scheduleId"`
}

type CompletionInput struct {
	Completed bool `json:"completed"`
}

type ScheduleRule struct {
	ID             int64
	StartAt        time.Time
	EndAtExclusive *time.Time
	TimeZone       string
	Times          []string
	Mode           string
	Standard       *Standard
	Weaning        *Weaning
	CreatedAt      time.Time
}
