package feeding

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	store := &Store{db: db}
	if _, err := db.Exec(`
		PRAGMA foreign_keys = ON;
		CREATE TABLE IF NOT EXISTS schedule_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			start_at TEXT NOT NULL,
			end_at_exclusive TEXT,
			time_zone TEXT NOT NULL,
			definition TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS feeding_completions (
			feeding_id TEXT PRIMARY KEY,
			completed_at TEXT NOT NULL
		);
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize database: %w", err)
	}
	return store, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Publish(ctx context.Context, rule ScheduleRule) (int64, error) {
	definition, err := json.Marshal(struct {
		Times    []string  `json:"times"`
		Mode     string    `json:"mode"`
		Standard *Standard `json:"standard,omitempty"`
		Weaning  *Weaning  `json:"weaning,omitempty"`
	}{rule.Times, rule.Mode, rule.Standard, rule.Weaning})
	if err != nil {
		return 0, err
	}
	var end any
	if rule.EndAtExclusive != nil {
		end = rule.EndAtExclusive.UTC().Format(time.RFC3339Nano)
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO schedule_rules (start_at, end_at_exclusive, time_zone, definition, created_at) VALUES (?, ?, ?, ?, ?)`,
		rule.StartAt.UTC().Format(time.RFC3339Nano), end, rule.TimeZone, string(definition), time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) Feedings(ctx context.Context, from, to time.Time) ([]Feeding, error) {
	if !to.After(from) {
		return nil, errors.New("to must be after from")
	}
	if to.Sub(from) > 93*24*time.Hour {
		return nil, errors.New("date range cannot exceed 93 days")
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, start_at, end_at_exclusive, time_zone, definition, created_at FROM schedule_rules ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]ScheduleRule, 0)
	for rows.Next() {
		var rule ScheduleRule
		var startText, endText sql.NullString
		var definition, createdText string
		if err := rows.Scan(&rule.ID, &startText, &endText, &rule.TimeZone, &definition, &createdText); err != nil {
			return nil, err
		}
		if err := parseStoredTime(startText.String, &rule.StartAt); err != nil {
			return nil, err
		}
		if endText.Valid {
			var end time.Time
			if err := parseStoredTime(endText.String, &end); err != nil {
				return nil, err
			}
			rule.EndAtExclusive = &end
		}
		if err := parseStoredTime(createdText, &rule.CreatedAt); err != nil {
			return nil, err
		}
		var definitionData struct {
			Times    []string  `json:"times"`
			Mode     string    `json:"mode"`
			Standard *Standard `json:"standard"`
			Weaning  *Weaning  `json:"weaning"`
		}
		if err := json.Unmarshal([]byte(definition), &definitionData); err != nil {
			return nil, err
		}
		rule.Times, rule.Mode = definitionData.Times, definitionData.Mode
		rule.Standard, rule.Weaning = definitionData.Standard, definitionData.Weaning
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	generated := make([]Feeding, 0)
	for _, rule := range rules {
		loc, err := time.LoadLocation(rule.TimeZone)
		if err != nil {
			return nil, err
		}
		first := localMidnight(from.Add(-24*time.Hour).In(loc), loc)
		last := localMidnight(to.Add(24*time.Hour).In(loc), loc)
		for day := first; !day.After(last); day = day.AddDate(0, 0, 1) {
			dayStart := day.UTC()
			if dayStart.Before(rule.StartAt) || (rule.EndAtExclusive != nil && !dayStart.Before(*rule.EndAtExclusive)) {
				continue
			}
			winner := newestRuleAt(rules, dayStart)
			if winner == nil || winner.ID != rule.ID {
				continue
			}
			dayFeedings, err := BuildDay(rule, day, rule.ID)
			if err != nil {
				return nil, err
			}
			for _, feeding := range dayFeedings {
				if !feeding.ScheduledAt.Before(from) && feeding.ScheduledAt.Before(to) {
					generated = append(generated, feeding)
				}
			}
		}
	}

	completions := make(map[string]bool, len(generated))
	for _, feeding := range generated {
		var completedAt string
		err := s.db.QueryRowContext(ctx, `SELECT completed_at FROM feeding_completions WHERE feeding_id = ?`, feeding.ID).Scan(&completedAt)
		if err == nil {
			completions[feeding.ID] = true
		} else if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	for i := range generated {
		generated[i].Completed = completions[generated[i].ID]
	}
	sort.Slice(generated, func(i, j int) bool { return generated[i].ScheduledAt.Before(generated[j].ScheduledAt) })
	return generated, nil
}

func (s *Store) SetCompleted(ctx context.Context, id string, completed bool) error {
	if len(id) > 128 {
		return errors.New("invalid feeding id")
	}
	if completed {
		_, err := s.db.ExecContext(ctx, `INSERT INTO feeding_completions (feeding_id, completed_at) VALUES (?, ?) ON CONFLICT(feeding_id) DO UPDATE SET completed_at = excluded.completed_at`, id, time.Now().UTC().Format(time.RFC3339Nano))
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM feeding_completions WHERE feeding_id = ?`, id)
	return err
}

func newestRuleAt(rules []ScheduleRule, instant time.Time) *ScheduleRule {
	var winner *ScheduleRule
	for i := range rules {
		rule := &rules[i]
		if instant.Before(rule.StartAt) || (rule.EndAtExclusive != nil && !instant.Before(*rule.EndAtExclusive)) {
			continue
		}
		if winner == nil || rule.ID > winner.ID {
			winner = rule
		}
	}
	return winner
}

func parseStoredTime(text string, into *time.Time) error {
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return err
	}
	*into = parsed.UTC()
	return nil
}
