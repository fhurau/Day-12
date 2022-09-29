package main

import (
	"Project/connection"
	"Project/middleware"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	route := mux.NewRouter()

	connection.DatabaseConnect()

	route.PathPrefix("/assets/").Handler(http.StripPrefix("/assets", http.FileServer(http.Dir("./assets"))))
	route.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads/"))))

	route.HandleFunc("/", Home).Methods("GET")
	route.HandleFunc("/contact", Contact).Methods("GET")
	route.HandleFunc("/addMyProject", AddMyProject).Methods("GET")
	route.HandleFunc("/addMP", middleware.UploadFile(AddMP)).Methods("POST")
	route.HandleFunc("/myProjectDetail/{id}", MyProjectDetail).Methods("GET")
	route.HandleFunc("/deleteMP/{id}", deleteMP).Methods("GET")
	route.HandleFunc("/editProject/{id}", edit).Methods("GET")
	route.HandleFunc("/update/{id}", middleware.UploadFileUpdate(update)).Methods("POST")
	route.HandleFunc("/form-register", formRegister).Methods("GET")
	route.HandleFunc("/register", register).Methods("POST")
	route.HandleFunc("/form-login", formlogin).Methods("GET")
	route.HandleFunc("/login", login).Methods("POST")
	route.HandleFunc("/logout", logout).Methods("GET")

	fmt.Println("Server Running")
	http.ListenAndServe("localhost:5000", route)

}

type MP struct {
	Title           string
	Description     string
	Duration        string
	ID              int
	StartDate       time.Time
	EndDate         time.Time
	Formatstartdate string
	Formatenddate   string
	User            string
	Image           string
	IsLogin         bool
	Technologies    []string
}

type User struct {
	ID       int
	Name     string
	Email    string
	Password string
}

type SessionData struct {
	IsLogin   bool
	UserName  string
	FlashData string
}

var Data = SessionData{}

func Home(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/index.html")

	if err != nil {
		w.Write([]byte("error : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}

	if session.Values["IsLogin"] != true {
		data, _ := connection.Con.Query(context.Background(), "SELECT  tb_project.id, title, description, duration, image, technologies, tb_user.name as user FROM tb_project LEFT JOIN tb_user ON tb_project.user_id = tb_user.id")

		var result []MP
		for data.Next() {
			var each = MP{}

			err := data.Scan(&each.ID, &each.Title, &each.Description, &each.Duration, &each.Image, &each.Technologies, &each.User)
			if err != nil {
				w.Write([]byte("error : " + err.Error()))
				return
			}
			result = append(result, each)
		}
		resData := map[string]interface{}{
			"DataSession": Data,
			"MP":          result,
		}

		w.WriteHeader(http.StatusOK)
		tmpl.Execute(w, resData)

	} else {

		sessionID := session.Values["ID"].(int)

		data, _ := connection.Con.Query(context.Background(), "SELECT tb_project.id, title, description, duration, image, technologies, tb_user.name as user FROM tb_project LEFT JOIN tb_user ON tb_project.user_id = tb_user.id  WHERE tb_project.user_id =$1", sessionID)

		var result []MP
		for data.Next() {
			var each = MP{}

			err := data.Scan(&each.ID, &each.Title, &each.Description, &each.Duration, &each.Image, &each.Technologies, &each.User)
			if err != nil {
				w.Write([]byte("error : " + err.Error()))
				return
			}
			result = append(result, each)
		}
		resData := map[string]interface{}{
			"DataSession": Data,
			"MP":          result,
		}

		w.WriteHeader(http.StatusOK)
		tmpl.Execute(w, resData)
	}

}
func Contact(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/contact.html")

	if err != nil {
		w.Write([]byte("error : " + err.Error()))
		return
	}
	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}
	data := map[string]interface{}{
		"DataSession": Data,
	}

	tmpl.Execute(w, data)

}
func AddMyProject(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/addMyProject.html")

	if err != nil {
		w.Write([]byte("error : " + err.Error()))
		return
	}

	tmpl.Execute(w, nil)

}

func AddMP(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var title = r.PostForm.Get("title")
	var description = r.PostForm.Get("description")
	var startDate = r.PostForm.Get("startDate")
	var endDate = r.PostForm.Get("endDate")
	var technologies = r.Form["checkbox"]

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	user := session.Values["ID"].(int)

	dataContext := r.Context().Value("dataFile")
	image := dataContext.(string)

	layout := "2006-01-02"
	parsingstartdate, _ := time.Parse(layout, startDate)
	parsingenddate, _ := time.Parse(layout, endDate)

	hours := parsingenddate.Sub(parsingstartdate).Hours()
	days := hours / 24

	var duration string
	if days > 0 {
		duration = strconv.FormatFloat(days, 'f', 0, 64) + " days"
	}

	_, err = connection.Con.Exec(context.Background(), "INSERT INTO tb_project(title, start_date, end_date, description, duration, user_id, image, technologies) VAlUES ($1, $2, $3, $4, $5, $6, $7, $8)", title, parsingstartdate, parsingenddate, description, duration, user, image, technologies)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)

}

func MyProjectDetail(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/myProjectDetail.html")

	if err != nil {
		w.Write([]byte("error : " + err.Error()))
		return
	}
	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}

	var MPDetail = MP{}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	err = connection.Con.QueryRow(context.Background(), "SELECT tb_project.id, title, start_date, end_date, description, duration, image, technologies, tb_user.name as user FROM tb_project LEFT JOIN tb_user ON tb_project.user_id = tb_user.id  WHERE tb_project.id=$1", id).Scan(
		&MPDetail.ID, &MPDetail.Title, &MPDetail.StartDate, &MPDetail.EndDate, &MPDetail.Description, &MPDetail.Duration, &MPDetail.Image, &MPDetail.Technologies, &MPDetail.User)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
	}

	MPDetail.Formatstartdate = MPDetail.StartDate.Format("2 january 2006")
	MPDetail.Formatenddate = MPDetail.EndDate.Format("2 january 2006")

	data := map[string]interface{}{
		"DataSession": Data,
		"MP":          MPDetail,
	}

	tmpl.Execute(w, data)

}

func deleteMP(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, err := connection.Con.Exec(context.Background(), "DELETE FROM tb_project WHERE id=$1", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func edit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset-utf8")

	var tmplt, err = template.ParseFiles("views/editProject.html")
	if err != nil {
		w.Write([]byte("file doesn't exist: " + err.Error()))
		return
	}
	var MPDetail = MP{}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	err = connection.Con.QueryRow(context.Background(), "SELECT id, title, start_date, end_date, description, duration FROM tb_project WHERE id=$1", id).Scan(&MPDetail.ID, &MPDetail.Title, &MPDetail.StartDate, &MPDetail.EndDate, &MPDetail.Description, &MPDetail.Duration)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
	}

	data := map[string]interface{}{
		"editProject": MPDetail,
	}
	tmplt.Execute(w, data)

}

func update(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var title = r.PostForm.Get("title")
	var description = r.PostForm.Get("description")
	var startDate = r.PostForm.Get("startDate")
	var endDate = r.PostForm.Get("endDate")
	var technologies = r.Form["checkbox"]

	layout := "2006-01-02"
	parsingstartdate, _ := time.Parse(layout, startDate)
	parsingenddate, _ := time.Parse(layout, endDate)

	hours := parsingenddate.Sub(parsingstartdate).Hours()
	days := hours / 24

	var duration string

	if days > 0 {
		duration = strconv.FormatFloat(days, 'f', 0, 64) + " days"
	}
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	dataContext := r.Context().Value("dataFile")
	image := dataContext.(string)

	sqlStatement := `UPDATE public.tb_project SET title=$2, start_date=$3, end_date=$4, description=$5, technologies=$6, image=$7, duration=$8
	WHERE id=$1;`

	_, err = connection.Con.Exec(context.Background(), sqlStatement, id, title, parsingstartdate, parsingenddate, description, technologies, image, duration)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func formRegister(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/form-register.html")

	if err != nil {
		w.Write([]byte("message : " + err.Error()))
		return
	}

	tmpl.Execute(w, nil)
}

func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var name = r.PostForm.Get("inputName")
	var email = r.PostForm.Get("inputEmail")
	var password = r.PostForm.Get("inputPassword")

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	_, err = connection.Con.Exec(context.Background(), "INSERT INTO tb_user(name, email, password) VALUES ($1, $2, $3)", name, email, passwordHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/form-login", http.StatusMovedPermanently)
}

func formlogin(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/form-login.html")

	if err != nil {
		w.Write([]byte("message : " + err.Error()))
		return
	}

	tmpl.Execute(w, nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var email = r.PostForm.Get("inputEmail")
	var password = r.PostForm.Get("inputPassword")

	user := User{}

	err = connection.Con.QueryRow(context.Background(), "SELECT * FROM tb_user WHERE email=$1", email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		w.Write([]byte("message : " + err.Error()))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	session.Values["Name"] = user.Name
	session.Values["Email"] = user.Email
	session.Values["ID"] = user.ID
	session.Values["IsLogin"] = true
	session.Options.MaxAge = 1800
	session.AddFlash("succesfull login", "message")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func logout(w http.ResponseWriter, r *http.Request) {

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
