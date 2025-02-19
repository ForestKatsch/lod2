package webhook

import (
	"lod2/cplane/redeploy"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-github/v50/github"
)

const githubWebhookSecretEnv = "GITHUB_WEBHOOK_SECRET"

var githubWebhookSecret string

// init loads the secret from the environment
func init() {
	githubWebhookSecret = os.Getenv(githubWebhookSecretEnv)
}

func GitHubWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if githubWebhookSecret == "" {
		http.Error(w, "Could not validate secret", http.StatusInternalServerError)
		log.Printf("cannot handle GitHub webhook without environment variable %s", githubWebhookSecretEnv)
		return
	}

	// Validate and parse the payload
	payload, err := github.ValidatePayload(r, []byte(githubWebhookSecret))
	if err != nil {
		http.Error(w, "Invalid payload signature", http.StatusForbidden)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		http.Error(w, "Could not parse webhook", http.StatusBadRequest)
		return
	}

	switch event := event.(type) {
	case *github.PushEvent:
		var ref = *event.Ref
		log.Printf("received a push event for ref '%s'", ref)

		if ref == "refs/heads/main" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Webhook received, redeploying"))
			log.Println("Redeploying...")
			redeploy.Redeploy()
			return
		}

		log.Printf("ignoring push event for non-main ref '%s'", ref)
	default:
		http.Error(w, "Unhandled event type", http.StatusNotImplemented)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook received"))
}

func GithubWebhookRouter() http.Handler {
	r := chi.NewRouter()

	r.Post("/", GitHubWebhookHandler)

	return r
}
