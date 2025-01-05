package server

import (
	"io"
	"net/http"
	"net/url"
	"os"
	path "path/filepath"
	"strings"

	"git.sr.ht/~kvo/go-std/errors"

	"main/logger"
	"main/site"
)

// Handle things like submission and file uploads/removals.
func handleTask(r *http.Request, user site.User, platform, id, cmd string) (int, pageData, [][2]string) {
	data := pageData{}

	res := r.URL.EscapedPath()
	statusCode := 200
	var headers [][2]string

	if cmd == "submit" {
		school, ok := schools[user.School]
		if !ok {
			logger.Debug(errors.New(nil, "unsupported platform"))
			statusCode = 500
			return statusCode, pageData{}, headers
		}
		err := school.Submit(user, platform, id)
		if err != nil {
			logger.Debug(errors.New(err, "cannot submit task"))
			statusCode = 500
			return statusCode, pageData{}, headers
		}
		index := strings.Index(res, "/submit")
		headers = [][2]string{{"Location", res[:index]}}
		statusCode = 302
	} else if cmd == "upload" {
		school, ok := schools[user.School]
		if !ok {
			logger.Debug(errors.New(nil, "unsupported platform"))
			statusCode = 500
			return statusCode, pageData{}, headers
		}
		err := school.UploadWork(user, platform, id, r)
		if err != nil {
			logger.Debug(errors.New(err, "cannot upload work"))
			statusCode = 500
			return statusCode, pageData{}, headers
		}
		index := strings.Index(res, "/upload")
		headers = [][2]string{{"Location", res[:index]}}
		statusCode = 302
	} else if cmd == "remove" {
		filenames := []string{}
		for name := range r.URL.Query() {
			filenames = append(filenames, name)
		}
		school, ok := schools[user.School]
		if !ok {
			logger.Debug(errors.New(nil, "unsupported platform"))
			statusCode = 500
			return statusCode, pageData{}, headers
		}
		err := school.RemoveWork(user, platform, id, filenames)
		if err != nil {
			logger.Debug(errors.New(err, "cannot remove worklink"))
			statusCode = 500
			return statusCode, pageData{}, headers
		}
		index := strings.Index(res, "/remove")
		headers = [][2]string{{"Location", res[:index]}}
		statusCode = 302
	} else {
		statusCode = 404
	}

	return statusCode, data, headers
}

func handleTaskReq(r *http.Request, user site.User) (int, pageData, [][2]string) {
	res := r.URL.EscapedPath()

	statusCode := 200
	var data pageData
	var headers [][2]string

	platform := res[7:]
	index := strings.Index(platform, "/")

	if index == -1 {
		statusCode = 404
		return statusCode, data, headers
	}

	taskId := platform[index+1:]
	platform = platform[:index]
	index = strings.Index(taskId, "/")

	if index == -1 {
		school, ok := schools[user.School]
		if !ok {
			logger.Debug(errors.New(nil, "unsupported platform"))
			statusCode = 500
			return statusCode, pageData{}, headers
		}
		assignment, err := school.Task(user, platform, taskId)
		if err != nil {
			logger.Debug(errors.New(err, "cannot fetch task"))
			statusCode = 500
			return statusCode, pageData{}, headers
		}

		data = genTaskPage(assignment, user)
	} else {
		taskCmd := taskId[index+1:]
		taskId = taskId[:index]

		statusCode, data, headers = handleTask(
			r,
			user,
			platform,
			taskId,
			taskCmd,
		)
	}

	data.User = userData{Name: user.DispName}
	return statusCode, data, headers
}

func dispatchAsset(w http.ResponseWriter, fullPath string, mimeType string) {
	w.Header().Set("Content-Type", mimeType+`, charset="utf-8"`)

	file, err := os.Open(fullPath)
	if err != nil {
		logger.Error(errors.New(err, "could not open %s", fullPath))
		w.WriteHeader(500)
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	if err != nil {
		logger.Debug(errors.New(err, "could not copy contents of %s", fullPath))
		w.WriteHeader(500)
	}
}

func assetHandler(w http.ResponseWriter, r *http.Request) {
	res := strings.Replace(r.URL.EscapedPath(), "/assets", "", 1)

	if strings.HasPrefix(res, "/icons") {
		fileStr := ""
		mimeType := ""

		switch name := strings.Replace(res, "/icons", "", 1); name {
		case "/apple-touch-icon.png":
			mimeType = "image/png"
			fileStr = "apple-touch-icon.png"
		case "/icon.svg":
			mimeType = "image/svg+xml"
			fileStr = "icon.svg"
		case "/icon-192.png":
			mimeType = "image/png"
			fileStr = "icon-512.png"
		case "/icon-512.png":
			mimeType = "image/png"
			fileStr = "icon-512.png"
		default:
			if name == "/" {
				alert(w, site.User{}, nil, 403)
			} else {
				alert(w, site.User{}, nil, 404)
			}
			return
		}

		fullPath := path.Join(resdir, "icons", fileStr)
		dispatchAsset(w, fullPath, mimeType)

	} else if res == "/wordmark.svg" {
		w.Header().Set("Cache-Control", "max-age=3600")
		fullPath := path.Join(resdir, "brand/wordmark.svg")
		dispatchAsset(w, fullPath, "image/svg+xml")

	} else if res == "/manifest.webmanifest" {
		fullPath := path.Join(resdir, "manifest.webmanifest")
		dispatchAsset(w, fullPath, "application/json")

	} else if res == "/styles.css" {
		fullPath := path.Join(resdir, "styles.css")
		dispatchAsset(w, fullPath, "text/css")

	} else if res == "/script.js" {
		w.Header().Set("Cache-Control", "max-age=3600")
		fullPath := path.Join(resdir, "script.js")
		dispatchAsset(w, fullPath, "text/javascript")

	} else if res == "/mainfont.woff2" {
		w.Header().Set("Cache-Control", "max-age=259200")
		fullPath := path.Join(resdir, "fonts/lato/mainfont.woff2")
		dispatchAsset(w, fullPath, "font/woff2")

	} else if res == "/navfont.woff2" {
		w.Header().Set("Cache-Control", "max-age=259200")
		fullPath := path.Join(resdir, "fonts/redhat/navfont.woff2")
		dispatchAsset(w, fullPath, "font/woff2")

	} else {
		if res == "/" {
			alert(w, site.User{}, nil, 403)
		} else {
			alert(w, site.User{}, nil, 404)
		}
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	validAuth := true
	redirect := r.URL.Query().Get("redirect")

	_, err := creds.LookupToken(r.Header.Get("Cookie"))
	if err != nil {
		validAuth = false
	}

	data := loginPageData
	if strings.HasPrefix(redirect, "/") {
		data.Body.LoginData.Redirect = "?" + r.URL.RawQuery
	}

	if !validAuth {
		if r.URL.Query().Get("auth") == "failed" {
			w.WriteHeader(401)
			data.Body.LoginData.Failed = true
			serve(w, data)
		} else {
			serve(w, data)
		}
	} else if !strings.HasPrefix(redirect, "/") {
		w.Header().Set("Location", "/timetable")
		w.WriteHeader(302)
	} else {
		w.Header().Set("Location", redirect)
		w.WriteHeader(302)
	}
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	validAuth := true

	_, err := creds.LookupToken(r.Header.Get("Cookie"))
	if err != nil {
		validAuth = false
	}

	if !validAuth {
		var cookie string
		err := r.ParseForm()

		// If err != nil, the "else" section of the next if/else block will
		// execute, which returns the "could not authenticate user" error.
		if err == nil {
			cookie, err = creds.Login(r.PostForm)
		}

		redirect := r.URL.Query().Get("redirect")
		if !strings.HasPrefix(redirect, "/") {
			redirect = "/timetable"
		}

		if err == nil {
			w.Header().Set("Location", redirect)
			w.Header().Set("Set-Cookie", cookie)
			w.WriteHeader(302)
		} else {
			logger.Debug(errors.New(err, "auth failed"))
			w.Header().Set("Location", "/login?auth=failed")
			w.WriteHeader(302)
		}
	} else {
		shoo(w, r)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	validAuth := true

	user, err := creds.LookupToken(r.Header.Get("Cookie"))
	if err != nil {
		validAuth = false
	}

	if validAuth {
		err = creds.Logout(r.Header.Get("Cookie"))
		if err == nil {
			w.Header().Set("Location", "/login")
			w.WriteHeader(302)
		} else {
			alert(w, user, err, 500)
		}
	} else {
		shoo(w, r)
	}
}

func resourceHandler(w http.ResponseWriter, r *http.Request) {
	user, err := creds.LookupToken(r.Header.Get("Cookie"))
	if err == nil {
		reqRes := r.URL.EscapedPath()

		statusCode := 200
		var respBody pageData
		platform := reqRes[5:]
		index := strings.Index(platform, "/")

		if index == -1 {
			alert(w, user, nil, 404)
			return
		}

		resId := platform[index+1:]
		platform = platform[:index]

		school, ok := schools[user.School]
		if !ok {
			err := errors.New(nil, "unsupported platform")
			alert(w, user, err, 404)
			return
		}
		res, err := school.Resource(user, platform, resId)
		if err != nil {
			alert(w, user, err, 500)
		}

		respBody = genResPage(res, user)
		w.WriteHeader(statusCode)
		serve(w, respBody)
	} else {
		shoo(w, r)
	}
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	user, err := creds.LookupToken(r.Header.Get("Cookie"))
	if err == nil {
		statusCode, respBody, respHeaders := handleTaskReq(r, user)
		for _, respHeader := range respHeaders {
			w.Header().Set(respHeader[0], respHeader[1])
		}
		w.WriteHeader(statusCode)
		serve(w, respBody)
	} else {
		shoo(w, r)
	}
}


func alert(w http.ResponseWriter, user site.User, err error, code int) {
	titles := map[int]string{
		403: "403 Forbidden",
		404: "404 Not Found",
		500: "500 Internal Server Error",
	}
	msgs := map[int]string{
		403: "You do not have permission to access this resource.",
		404: "The requested resource was not found on the server.",
		500: "The server encountered an unexpected error and cannot continue.",
	}
	msg := msgs[code]
	if code == 500 && err != nil {
		msg = err.Error()
		logger.Debug(err)
	}
	username := user.DispName
	if username == "" {
		username = "none"
	}
	w.WriteHeader(code)
	serve(w, pageData{
		PageType: "error",
		Head: headData{
			Title: titles[code],
		},
		Body: bodyData{
			ErrorData: errData{
				Heading: titles[code],
				Message: msg,
			},
		},
		User: userData{
			Name: username,
		},
	})
}

func shoo(w http.ResponseWriter, r *http.Request) {
	redirect := "/login?redirect=" + url.QueryEscape(r.URL.String())
	w.Header().Set("Location", redirect)
	w.WriteHeader(302)
}

func serve(w http.ResponseWriter, data pageData) {
	err := templates.ExecuteTemplate(w, "page", data)
	if err != nil {
		logger.Debug(errors.New(err, "serve %s", data.PageType))
	}
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	user, err := creds.LookupToken(r.Header.Get("Cookie"))
	if err == nil {
		content, err := genRes("/tasks", user)
		if err != nil {
			alert(w, user, err, 500)
		} else {
			w.Header().Set("Cache-Control", "max-age=2400")
			serve(w, content)
		}
	} else {
		shoo(w, r)
	}
}

func timetableHandler(w http.ResponseWriter, r *http.Request) {
	user, err := creds.LookupToken(r.Header.Get("Cookie"))
	if err == nil {
		content, err := genRes("/timetable", user)
		if err != nil {
			alert(w, user, err, 500)
		} else {
			w.Header().Set("Cache-Control", "max-age=2400")
			serve(w, content)
		}
	} else {
		shoo(w, r)
	}
}

func gradesHandler(w http.ResponseWriter, r *http.Request) {
	user, err := creds.LookupToken(r.Header.Get("Cookie"))
	if err == nil {
		content, err := genRes("/grades", user)
		if err != nil {
			alert(w, user, err, 500)
		} else {
			w.Header().Set("Cache-Control", "max-age=2400")
			serve(w, content)
		}
	} else {
		shoo(w, r)
	}
}

func resHandler(w http.ResponseWriter, r *http.Request) {
	user, err := creds.LookupToken(r.Header.Get("Cookie"))
	if err == nil {
		content, err := genRes("/res", user)
		if err != nil {
			alert(w, user, err, 500)
		} else {
			w.Header().Set("Cache-Control", "max-age=2400")
			serve(w, content)
		}
	} else {
		shoo(w, r)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	user, err := creds.LookupToken(r.Header.Get("Cookie"))
	if r.URL.EscapedPath() == "/favicon.ico" {
		fullPath := path.Join(resdir, "/icons/favicon.ico")
		dispatchAsset(w, fullPath, "text/plain")
	} else if err != nil {
		// user not logged in (and not on login page)
		redirect := "/login?redirect=" + url.QueryEscape(r.URL.String())
		w.Header().Set("Location", redirect)
		w.WriteHeader(302)
	} else if err == nil && r.URL.EscapedPath() == "/timetable.png" {
		err := TimetablePNG(user, w)
		if err != nil {
			logger.Error(err)
		}
	} else if err == nil && r.URL.EscapedPath() == "/" {
		w.Header().Set("Location", "/timetable")
		w.WriteHeader(302)
	} else if err == nil && r.URL.EscapedPath() != "/" {
		// requested path has no associated handler
		alert(w, user, nil, 404)
	}
}
