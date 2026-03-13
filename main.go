package main

import (
	"log"
	"os"

	"github.com/Milansaji/Grit/grit"
	"github.com/joho/godotenv"
)

var (
	ServiceAccountKeyPath string
	ProjectID             string
	FirebaseAPIKey        string
	JWTSecret             string
	Port                  uint
)

type Post struct {
	Title  string `json:"title"`
	Body   string `json:"body"`
	Author string `json:"author"`
	Status string `json:"status"` // "draft" | "published"
}

type Comment struct {
	PostID  string `json:"post_id"`
	Author  string `json:"author"`
	Content string `json:"content"`
}

func LoadEnv() {

	_ = godotenv.Load(".env")

	ServiceAccountKeyPath = os.Getenv("SERVICE_ACCOUNT_KEY_PATH")
	ProjectID = os.Getenv("PROJECT_ID")
	FirebaseAPIKey = os.Getenv("FIREBASE_API_KEY")
	JWTSecret = os.Getenv("JWT_SECRET_KEY")
}

func main() {

	LoadEnv()

	// Register models
	grit.RegisterModel("posts", &Post{})
	grit.RegisterModel("comments", &Comment{})

	// Init Firebase
	grit.InitFirebase(ServiceAccountKeyPath, ProjectID, FirebaseAPIKey)

	r := grit.NewRouter()
	protect := grit.FirebaseProtected(JWTSecret)
	adminOnly := grit.RequirePermission(JWTSecret, "admin:all")
	userRead := grit.RequirePermission(JWTSecret, "user:read")

	// Auth
	r.Post("/auth/signup", grit.FirebaseSignupWithEmail(JWTSecret))
	r.Post("/auth/signin", grit.FirebaseSigninWithEmail(JWTSecret))
	r.Post("/auth/signout", protect(grit.FirebaseSignoutHandler))
	r.Get("/auth/me", protect(grit.FirebaseMeHandler))

	// Posts CRUD
	r.Post("/posts/create", adminOnly(protect(grit.FirestoreC("posts"))))
	r.Get("/posts", protect(grit.FirestoreR("posts")))
	r.Get("/post", userRead(protect(grit.FirestoreGetByID("posts"))))
	r.Put("/post", adminOnly(protect(grit.FirestoreU("posts"))))
	r.Patch("/post", adminOnly(protect(grit.FirestoreU("posts"))))
	r.Delete("/post", adminOnly(protect(grit.FirestoreD("posts"))))

	// Comments CRUD
	r.Post("/comments", protect(grit.FirestoreC("comments")))
	r.Get("/comments", protect(grit.FirestoreR("comments")))
	r.Get("/comment", protect(grit.FirestoreGetByID("comments")))
	r.Put("/comment", protect(grit.FirestoreU("comments")))
	r.Delete("/comment", protect(grit.FirestoreD("comments")))

	// Query routes
	r.Get("/posts/by-author", protect(grit.FirestoreWhere("posts", "author", "==")))
	r.Get("/posts/published", protect(grit.FirestoreWhere("posts", "status", "==")))
	r.Get("/comments/by-post", protect(grit.FirestoreWhere("comments", "post_id", "==")))

	// Health
	r.Get("/health", grit.HealthHandler)

	if err := r.Start("3001"); err != nil {
		log.Fatal(err)
	}
}
