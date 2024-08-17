package handlers

import (
	"fmt"
	"forum/app/models"
	"forum/pkg"
	"log"
	"net/http"
)

func (app *App) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		fmt.Println("This error in Home")
		pkg.ErrorHandler(w, http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		pkg.ErrorHandler(w, http.StatusMethodNotAllowed)
		return
	}
	// fmt.Println(r.Context().Value(KeyUserType(keyUser)), 1111)
	// user, ok := r.Context().Value(KeyUserType(keyUser)).(models.User)
	// if !ok {
	// 	pkg.ErrorHandler(w, http.StatusUnauthorized)
	// 	return
	// }
	posts, err := app.postService.GetAllPosts()
	if err != nil {
		log.Println(err)
		pkg.ErrorHandler(w, http.StatusInternalServerError)
		return
	}
	data := models.Data{
		Posts: posts,
		// User:  user,
		Genre: "/",
	}
	fmt.Println("This is A ImageURL: ", posts[0].ImageURL)
	pkg.RenderTemplate(w, "index.html", data)
}

func (app *App) WelcomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {

		pkg.ErrorHandler(w, http.StatusMethodNotAllowed)
		return
	}
	posts, err := app.postService.GetAllPosts()
	if err != nil {
		log.Println(err)
		pkg.ErrorHandler(w, http.StatusInternalServerError)
		return
	}
	data := models.Data{
		Posts: posts,
	}
	pkg.RenderTemplate(w, "welcome.html", data)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

	}
}
