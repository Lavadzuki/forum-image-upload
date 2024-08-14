package main

import (
	"context"
	"fmt"
	"forum/app/config"
	"forum/app/handlers"
	"forum/app/repository"
	"forum/app/service/post"
	"forum/app/service/session"
	"forum/app/service/user/auth"
	"forum/app/service/user/user"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const (
	uploadPath   = "./uploads"
	uploadedFile = "uploaded_image"
)

func main() {
	err := os.MkdirAll(uploadPath, os.ModePerm)
	if err != nil {
		log.Println(err)
		return
	}

	http.HandleFunc("/test", uploadPage)
	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/uploads/", serveImage)
	fmt.Println("Server starting on http://localhost:8080")
	http.ListenAndServe(":8080", nil)

	cfg, err := config.InitConfig("./config/config.json")
	if err != nil {
		log.Fatalln(err)
		return
	}

	if cfg.ServerAddress != cfg.Port {
		fmt.Sprintf(cfg.ServerAddress+":", cfg.Port)
	}

	db, err := repository.NewDB(cfg.Database)
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer db.Close()

	repo := repository.NewRepo(db)
	authservice := auth.NewAuthService(repo)
	userservice := user.NewUserService(repo)

	sessionService := session.NewSessionService(repo)

	postService := post.NewPostService(repo)

	app := handlers.NewAppService(authservice, sessionService, postService, userservice, cfg)
	server := app.Run(cfg.Http)

	go app.ClearSession()

	go func() {
		log.Printf("server started at http://localhost%s", cfg.Port)
		err := server.ListenAndServe()
		if err != nil {
			log.Printf("listen %s ", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("shutting down servers ...")
	err = server.Shutdown(ctx)
	if err != nil {
		log.Printf("server shut down:%s", err)
	}
	log.Println("server stopped")
}

func uploadPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl := `<html><body>
                    <form action="/upload" method="post" enctype="multipart/form-data">
                        <input type="file" name="file" />
                        <input type="submit" value="Upload" />
                    </form>
                    {{if .ImageURL}}<img src="{{.ImageURL}}" alt="Uploaded Image" style="max-width: 100%; max-height: 500px;" />{{end}}
                 </body></html>`
		t, _ := template.New("upload").Parse(tmpl)
		t.Execute(w, map[string]interface{}{
			"ImageURL": "/uploads/" + uploadedFile,
		})
	}
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "File upload error", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		outFile, err := os.Create(filepath.Join(uploadPath, uploadedFile))
		if err != nil {
			http.Error(w, "File save error", http.StatusInternalServerError)
			return
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, file)
		if err != nil {
			http.Error(w, "File copy error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func serveImage(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join(uploadPath, filepath.Base(r.URL.Path))
	http.ServeFile(w, r, filePath)
}
