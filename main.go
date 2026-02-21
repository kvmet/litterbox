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
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Submit an Idea</title>
    <style>
        body { font-family: sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        textarea { width: 100%; min-height: 200px; padding: 10px; font-size: 16px; box-sizing: border-box; }
        button { padding: 10px 20px; font-size: 16px; margin-top: 10px; cursor: pointer; }
    </style>
</head>
<body>
    <h1>Submit Your Idea</h1>
    <form method="POST" action="/submit">
        <textarea name="text" placeholder="Enter your idea here..." required></textarea>
        <br>
        <button type="submit">Submit</button>
    </form>
</body>
</html>
`
	e.Response.Header().Set("Content-Type", "text/html")
	e.Response.WriteHeader(http.StatusOK)
	_, err := e.Response.Write([]byte(html))
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

	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Thank You</title>
    <style>
        body { font-family: sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; text-align: center; }
        a { color: #0066cc; }
    </style>
</head>
<body>
    <h1>Thank You!</h1>
    <p>Your idea has been submitted successfully.</p>
    <a href="/">Submit another idea</a>
</body>
</html>
`
	e.Response.Header().Set("Content-Type", "text/html")
	e.Response.WriteHeader(http.StatusOK)
	_, err = e.Response.Write([]byte(html))
	return err
}

func handleReviewPage(e *core.RequestEvent, app *pocketbase.PocketBase) error {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Review Submissions</title>
    <script src="https://cdn.jsdelivr.net/npm/pocketbase@0.21.1/dist/pocketbase.umd.js"></script>
    <style>
        body { font-family: sans-serif; max-width: 1200px; margin: 20px auto; padding: 20px; }
        .login-form { max-width: 400px; margin: 50px auto; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
        .login-form input { width: 100%; padding: 10px; margin: 10px 0; box-sizing: border-box; }
        .login-form button { width: 100%; padding: 10px; background: #2196F3; color: white; border: none; cursor: pointer; }
        .error { color: #f44336; margin: 10px 0; }
        #review-content { display: none; }
        .group { margin-bottom: 30px; border: 1px solid #ddd; padding: 15px; border-radius: 5px; }
        .group h2 { margin-top: 0; text-transform: uppercase; }
        .submission { background: #f9f9f9; padding: 15px; margin: 10px 0; border-radius: 5px; }
        .submission-text { white-space: pre-wrap; margin: 10px 0; }
        .meta { font-size: 12px; color: #666; }
        button { padding: 5px 10px; margin-right: 5px; cursor: pointer; }
        .btn-approved { background: #4CAF50; color: white; border: none; }
        .btn-hidden { background: #FF9800; color: white; border: none; }
        .btn-deleted { background: #f44336; color: white; border: none; }
        .btn-new { background: #2196F3; color: white; border: none; }
        .btn-done { background: #9C27B0; color: white; border: none; }
        .group h2 { user-select: none; }
    </style>
</head>
<body>
    <div id="login-form" class="login-form">
        <h2>Login to Review</h2>
        <div id="error" class="error"></div>
        <form onsubmit="login(); return false;">
            <input type="email" id="email" placeholder="Email" required>
            <input type="password" id="password" placeholder="Password" required>
            <button type="submit">Login</button>
        </form>
    </div>

    <div id="review-content">
        <h1>Review Submissions</h1>
        <p><button onclick="logout()">Logout</button></p>
        <div id="submissions"></div>
    </div>

    <script>
        const pb = new PocketBase(window.location.origin);

        async function login() {
            const email = document.getElementById('email').value;
            const password = document.getElementById('password').value;
            const errorDiv = document.getElementById('error');

            try {
                await pb.collection('_superusers').authWithPassword(email, password);
                errorDiv.textContent = '';
                showReviewContent();
            } catch (err) {
                errorDiv.textContent = 'Login failed: ' + (err.message || 'Invalid credentials');
            }
        }

        async function logout() {
            pb.authStore.clear();
            document.getElementById('login-form').style.display = 'block';
            document.getElementById('review-content').style.display = 'none';
        }

        async function updateStatus(id, status) {
            try {
                await pb.send('/api/submissions/' + id + '/status', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': pb.authStore.token
                    },
                    body: JSON.stringify({ status: status })
                });
                loadSubmissions();
            } catch (err) {
                alert('Failed to update: ' + err.message);
            }
        }

        async function loadSubmissions() {
            try {
                const records = await pb.collection('submissions').getFullList({
                    sort: '-created',
                });

                const grouped = {
                    new: [],
                    approved: [],
                    hidden: [],
                    deleted: [],
                    done: []
                };

                records.forEach(record => {
                    if (grouped[record.status]) {
                        grouped[record.status].push(record);
                    }
                });

                let html = '';
                ['approved', 'new', 'done', 'hidden', 'deleted'].forEach(status => {
                    const groupId = 'group-' + status;
                    html += '<div class="group">';
                    html += '<h2 onclick="toggleGroup(\'' + groupId + '\')" style="cursor: pointer;">';
                    html += '<span id="' + groupId + '-icon">▼</span> ';
                    html += status.toUpperCase() + ' (' + grouped[status].length + ')';
                    html += '</h2>';
                    html += '<div id="' + groupId + '">';

                    if (grouped[status].length === 0) {
                        html += '<p>No submissions</p>';
                    } else {
                        grouped[status].forEach(record => {
                            const text = escapeHtml(record.text);
                            html += '<div class="submission">';
                            html += '<div class="submission-text">' + text + '</div>';
                            html += '<div class="meta">Created: ' + record.created + '</div>';
                            html += '<div>';
                            html += '<button class="btn-approved" onclick="updateStatus(\'' + record.id + '\', \'approved\')">Approved</button>';
                            html += '<button class="btn-new" onclick="updateStatus(\'' + record.id + '\', \'new\')">New</button>';
                            html += '<button class="btn-done" onclick="updateStatus(\'' + record.id + '\', \'done\')">Done</button>';
                            html += '<button class="btn-hidden" onclick="updateStatus(\'' + record.id + '\', \'hidden\')">Hidden</button>';
                            html += '<button class="btn-deleted" onclick="updateStatus(\'' + record.id + '\', \'deleted\')">Deleted</button>';
                            html += '</div></div>';
                        });
                    }
                    html += '</div></div>';
                });

                document.getElementById('submissions').innerHTML = html;
            } catch (err) {
                console.error('Failed to load submissions:', err);
            }
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        function toggleGroup(groupId) {
            const group = document.getElementById(groupId);
            const icon = document.getElementById(groupId + '-icon');
            if (group.style.display === 'none') {
                group.style.display = 'block';
                icon.textContent = '▼';
            } else {
                group.style.display = 'none';
                icon.textContent = '▶';
            }
        }
</text>


        async function showReviewContent() {
            document.getElementById('login-form').style.display = 'none';
            document.getElementById('review-content').style.display = 'block';
            await loadSubmissions();
        }

        // Check if already logged in
        if (pb.authStore.isValid) {
            showReviewContent();
        }
    </script>
</body>
</html>
`
	e.Response.Header().Set("Content-Type", "text/html")
	e.Response.WriteHeader(http.StatusOK)
	_, err := e.Response.Write([]byte(html))
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
