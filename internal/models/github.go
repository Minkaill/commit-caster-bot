package models

import "time"

// GitHubWebhookPayload представляет данные от GitHub webhook
type GitHubWebhookPayload struct {
	Ref        string     `json:"ref"`
	Before     string     `json:"before"`
	After      string     `json:"after"`
	Repository Repository `json:"repository"`
	Pusher     Pusher     `json:"pusher"`
	Commits    []Commit   `json:"commits"`
	HeadCommit Commit     `json:"head_commit"`
}

type Repository struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	HTMLURL     string `json:"html_url"`
	Description string `json:"description"`
}

type Pusher struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Commit struct {
	ID        string    `json:"id"`
	TreeID    string    `json:"tree_id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	URL       string    `json:"url"`
	Author    Author    `json:"author"`
	Added     []string  `json:"added"`
	Removed   []string  `json:"removed"`
	Modified  []string  `json:"modified"`
}

type Author struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}
