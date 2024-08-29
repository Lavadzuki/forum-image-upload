package handlers

import (
	"fmt"
	"forum/app/models"
	"forum/pkg"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	uploadPath    = "./uploads"
	uploadedFile  = "uploaded_image"
	maxUploadSize = 20 << 20 // 20 MB
)

func (app *App) PostHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		pkg.RenderTemplate(w, "createpost.html", models.Data{})
		return
	case http.MethodPost:
		// Parse the multipart form
		err := r.ParseMultipartForm(maxUploadSize)
		if err != nil {
			pkg.ErrorHandler(w, http.StatusInternalServerError)
			return
		}

		// Retrieve form values
		title := r.FormValue("title")
		message := r.FormValue("message")
		genre := r.Form["category"]

		if len(genre) == 0 {
			fmt.Println("StatusBad is here")
			pkg.ErrorHandler(w, http.StatusBadRequest)
			return
		}

		user, ok := r.Context().Value(KeyUserType(keyUser)).(models.User)
		if !ok {
			pkg.ErrorHandler(w, http.StatusUnauthorized)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil && err != http.ErrMissingFile {
			http.Error(w, "File upload error", http.StatusInternalServerError)
			return
		}

		if err == http.ErrMissingFile {
			fmt.Println("NO FILE!!")
			file = nil
			header = nil
		} else {
			defer file.Close()

			// Check file size
			if header.Size > maxUploadSize {
				// http.Error(w, "File too large. Maximum size is 20MB.", http.StatusBadRequest)
				pkg.ErrorHandler(w, http.StatusInternalServerError)
				return
			}

			// Ensure the upload directory exists
			if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
				err = os.MkdirAll(uploadPath, os.ModePerm)
				if err != nil {
					http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
					return
				}
			}

			// Create the file in the upload directory
			outFile, err := os.Create(filepath.Join(uploadPath, header.Filename))
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
		}

		var post models.Post
		if header != nil {
			post = models.Post{
				Title:       title,
				Content:     message,
				Category:    models.Stringslice(genre),
				Author:      user,
				CreatedTime: time.Now().Format(time.RFC822),
				ImageURL:    "/uploads/" + header.Filename,
			}
		} else {
			post = models.Post{
				Title:       title,
				Content:     message,
				Category:    models.Stringslice(genre),
				Author:      user,
				CreatedTime: time.Now().Format(time.RFC822),
				ImageURL:    "",
			}
		}

		// fmt.Println("This is ImageURL", post.ImageURL)
		status, err := app.postService.CreatePost(&post)
		if err != nil {
			log.Println(err)
			switch status {
			case http.StatusInternalServerError:
				pkg.ErrorHandler(w, http.StatusInternalServerError)
				return
			case http.StatusBadRequest:
				pkg.ErrorHandler(w, http.StatusBadRequest)
				return
			}
		}

		// Redirect the user to the main page after successful post creation
		http.Redirect(w, r, "/", http.StatusSeeOther)

	default:
		// Handle unsupported methods
		pkg.ErrorHandler(w, http.StatusMethodNotAllowed)
		return
	}
}
