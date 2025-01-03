package server

import (
	"html/template"
	"io/fs"
	"net/http"
	"os"
	path "path/filepath"
	_ "time/tzdata"

	"git.sr.ht/~kvo/go-std/errors"
	"git.sr.ht/~kvo/go-std/slices"
	"github.com/BurntSushi/toml"

	"main/logger"
	"main/site"
)

var (
	cfgdir    string
	creds     Creds
	logdir    string
	resdir    string
	schools   = make(map[string]*site.Mux)
	templates *template.Template
)

type Config struct {
	SaveLogs bool `toml:"save-logs"`
}

func announce(version string) {
	logger.Info("Running %s", version)
}

func loadcfg(cfgpath string) error {
	// Default config
	cfg := Config{
		SaveLogs: false,
	}
	file, err := os.Open(cfgpath)
	if err != nil {
		// user has missing (and thus, default) config
		return nil
	}
	defer file.Close()
	_, err = toml.NewDecoder(file).Decode(&cfg)
	if err != nil {
		return errors.New(err, "cannot parse server config: %s", cfgpath)
	}
	if cfg.SaveLogs {
		err = logger.UseConfigFile(logdir)
		if err != nil {
			return errors.Wrap(err)
		}
	}
	return nil
}

func loadtmpl(resdir string) error {
	tmplPath := path.Join(resdir, "templates")
	err := os.MkdirAll(tmplPath, os.ModePerm)
	if err != nil {
		return errors.Wrap(err)
	}
	required := []string{
		"body/error",
		"body/grades",
		"body/login",
		"body/main",
		"body/resource",
		"body/resources",
		"body/task",
		"body/tasks",
		"body/timetable",
		"components/footer",
		"components/header",
		"components/nav",
		"head",
		"page",
	}
	for i := range required {
		required[i] = path.Join(tmplPath, required[i]+".tmpl")
	}
	var files []string
	err = path.WalkDir(tmplPath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == ".DS_Store" {
			return fs.SkipDir
		} else if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	files = slices.Remove(files, tmplPath)
	missing := make([]string, 0, len(required))
	for _, file := range required {
		if !slices.Has(files, file) {
			missing = append(missing, file)
		}
	}
	if len(missing) != 0 {
		return errors.New(nil, "Missing templates:\n%+v", missing)
	}
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
	}
	templates = template.Must(template.New("").Funcs(funcMap).ParseFiles(files...))
	return nil
}

func Configure(version string) error {
	creds.Tokens = make(map[string]site.Uid)
	creds.Users = make(map[site.Uid]site.User)
	execpath, err := os.Executable()
	if err != nil {
		logger.Fatal(errors.New(err, "cannot get path to executable"))
	}
	resdir = path.Join(path.Dir(execpath), "../../../res/")
	logdir = path.Join(path.Dir(execpath), "../../../log/")
	cfgdir = path.Join(path.Dir(execpath), "../../../cfg/")
	cfgpath := path.Join(cfgdir, "server.cfg")
	announce(version)
	if err := loadcfg(cfgpath); err != nil {
		logger.Error(err)
		logger.Warn("Resorting to default configuration...")
	}
	if err := loadtmpl(resdir); err != nil {
		return errors.New(err, "cannot load HTML templates")
	}
	logger.Info("Successfully loaded HTML templates")
	return nil
}

func Run(tls bool) error {
	cert := path.Join(resdir, "cert.pem")
	key := path.Join(resdir, "key.pem")

	mux := http.NewServeMux()

	mux.HandleFunc("/assets/", assetHandler)
	mux.HandleFunc("/res", resHandler)
	mux.HandleFunc("/res/", resourceHandler)
	mux.HandleFunc("/tasks", tasksHandler)
	mux.HandleFunc("/tasks/", taskHandler)
	mux.HandleFunc("/timetable", timetableHandler)
	mux.HandleFunc("/grades", gradesHandler)

	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/auth", authHandler)
	mux.HandleFunc("/", rootHandler)

	if tls {
		logger.Info("Running on port 443")
		return http.ListenAndServeTLS(":443", cert, key, mux)
	} else {
		logger.Warn("Running on port 8080 (without TLS). DO NOT USE THIS IN PRODUCTION!")
		return http.ListenAndServe("localhost:8080", mux)
	}
}
