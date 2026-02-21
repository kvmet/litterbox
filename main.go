package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	_ "litterbox/migrations"
)

func loadTemplate(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func main() {
	app := pocketbase.New()

	// loosely check if it was executed using "go run"
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: isGoRun,
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Public submission form
		se.Router.GET("/", func(e *core.RequestEvent) error {
			return handleSubmissionForm(e)
		})

		// Submit handler
		se.Router.POST("/submit", func(e *core.RequestEvent) error {
			return handleSubmitSubmission(e, app)
		})

		// Review page (client-side auth with JS SDK)
		se.Router.GET("/review", func(e *core.RequestEvent) error {
			return handleReviewPage(e, app)
		})

		// Update status API endpoint
		se.Router.POST("/api/submissions/{id}/status", func(e *core.RequestEvent) error {
			return handleUpdateStatusAPI(e, app)
		})

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

func handleSubmissionForm(e *core.RequestEvent) error {
	html, err := loadTemplate("templates/submission_form.html")
	if err != nil {
		e.Response.WriteHeader(http.StatusInternalServerError)
		e.Response.Write([]byte("Failed to load template"))
		return err
	}
	e.Response.Header().Set("Content-Type", "text/html")
	e.Response.WriteHeader(http.StatusOK)
	_, err = e.Response.Write([]byte(html))
	return err
}

func handleSubmitSubmission(e *core.RequestEvent, app *pocketbase.PocketBase) error {
	collection, err := app.FindCollectionByNameOrId("submissions")
	if err != nil {
		e.Response.WriteHeader(http.StatusInternalServerError)
		e.Response.Write([]byte("Collection not found. Please create the 'submissions' collection first."))
		return nil
	}

	record := core.NewRecord(collection)
	record.Set("text", e.Request.PostFormValue("text"))
	record.Set("status", "new")

	if err := app.Save(record); err != nil {
		e.Response.WriteHeader(http.StatusInternalServerError)
		e.Response.Write([]byte("Failed to save submission"))
		return nil
	}

	html, err := loadTemplate("templates/thank_you.html")
	if err != nil {
		e.Response.WriteHeader(http.StatusInternalServerError)
		e.Response.Write([]byte("Failed to load template"))
		return err
	}
	e.Response.Header().Set("Content-Type", "text/html")
	e.Response.WriteHeader(http.StatusOK)
	_, err = e.Response.Write([]byte(html))
	return err
}

func handleReviewPage(e *core.RequestEvent, app *pocketbase.PocketBase) error {
	html, err := loadTemplate("templates/review.html")
	if err != nil {
		e.Response.WriteHeader(http.StatusInternalServerError)
		e.Response.Write([]byte("Failed to load template"))
		return err
	}
	e.Response.Header().Set("Content-Type", "text/html")
	e.Response.WriteHeader(http.StatusOK)
	_, err = e.Response.Write([]byte(html))
	return err
}

func handleUpdateStatusAPI(e *core.RequestEvent, app *pocketbase.PocketBase) error {
	// Auth is checked by PocketBase via Authorization header
	if e.Auth == nil {
		e.Response.WriteHeader(http.StatusUnauthorized)
		e.Response.Write([]byte(`{"error":"Unauthorized"}`))
		return nil
	}

	id := e.Request.PathValue("id")

	// Parse JSON body
	var body struct {
		Status string `json:"status"`
	}
	if err := e.BindBody(&body); err != nil {
		e.Response.WriteHeader(http.StatusBadRequest)
		e.Response.Write([]byte(`{"error":"Invalid request body"}`))
		return nil
	}

	record, err := app.FindRecordById("submissions", id)
	if err != nil {
		e.Response.WriteHeader(http.StatusNotFound)
		e.Response.Write([]byte(`{"error":"Submission not found"}`))
		return nil
	}

	record.Set("status", body.Status)

	if err := app.Save(record); err != nil {
		e.Response.WriteHeader(http.StatusInternalServerError)
		e.Response.Write([]byte(`{"error":"Failed to update submission"}`))
		return nil
	}

	e.Response.Header().Set("Content-Type", "application/json")
	e.Response.WriteHeader(http.StatusOK)
	e.Response.Write([]byte(`{"success":true}`))
	return nil
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
