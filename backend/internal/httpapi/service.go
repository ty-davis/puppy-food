package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/swaggest/openapi-go/openapi31"
	"github.com/swaggest/rest/web"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	"puppy-food/backend/internal/feeding"
)

func NewService(store *feeding.Store) *web.Service {
	service := web.NewService(openapi31.NewReflector())
	service.OpenAPISchema().SetTitle("Puppy Feeding Schedule")
	service.OpenAPISchema().SetVersion("1.0.0")

	getFeedings := usecase.NewInteractor(func(ctx context.Context, input struct {
		From string `query:"from" format:"date-time"`
		To   string `query:"to" format:"date-time"`
	}, output *feeding.FeedingList) error {
		from, err := time.Parse(time.RFC3339, input.From)
		if err != nil {
			return invalid("from must be a UTC date-time")
		}
		to, err := time.Parse(time.RFC3339, input.To)
		if err != nil {
			return invalid("to must be a UTC date-time")
		}
		output.Feedings, err = store.Feedings(ctx, from.UTC(), to.UTC())
		if err != nil {
			return invalid(err.Error())
		}
		return nil
	})
	getFeedings.SetTitle("List feedings")
	getFeedings.SetName("listFeedings")
	getFeedings.SetDescription("List scheduled feedings between two UTC timestamps.")
	service.Get("/api/feedings", getFeedings)

	publish := usecase.NewInteractor(func(ctx context.Context, input feeding.PublishInput, output *feeding.PublishResult) error {
		rule, err := feeding.ValidatePublish(input, time.Now())
		if err != nil {
			return invalid(err.Error())
		}
		output.ScheduleID, err = store.Publish(ctx, rule)
		if err != nil {
			return fmt.Errorf("save schedule: %w", err)
		}
		return nil
	})
	publish.SetTitle("Publish schedule")
	publish.SetName("publishSchedule")
	publish.SetDescription("Publish a standard or weaning schedule. A newer schedule replaces overlapping dates.")
	service.Post("/api/schedules", publish)

	setCompletion := usecase.NewInteractor(func(ctx context.Context, input struct {
		ID string `path:"id" minLength:"1" maxLength:"128"`
		feeding.CompletionInput
	}, output *struct {
		ID        string `json:"id"`
		Completed bool   `json:"completed"`
	}) error {
		if err := store.SetCompleted(ctx, input.ID, input.Completed); err != nil {
			return invalid(err.Error())
		}
		output.ID = input.ID
		output.Completed = input.Completed
		return nil
	})
	setCompletion.SetTitle("Complete or remove a feeding")
	setCompletion.SetName("setFeedingCompletion")
	service.Patch("/api/feedings/{id}", setCompletion)

	return service
}

func invalid(message string) error {
	return status.Wrap(errors.New(message), status.InvalidArgument)
}

func SPA(directory string) func(http.Handler) http.Handler {
	absolute, _ := filepath.Abs(directory)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/openapi.json" {
				next.ServeHTTP(w, r)
				return
			}
			candidate := filepath.Join(absolute, filepath.Clean("/"+r.URL.Path))
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				http.ServeFile(w, r, candidate)
				return
			}
			http.ServeFile(w, r, filepath.Join(absolute, "index.html"))
		})
	}
}
