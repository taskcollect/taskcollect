package server

import (
	"html/template"
)

// Primary (page, head, body)

type pageData struct {
	PageType string
	Head     headData
	Body     bodyData
	User     userData
}

type headData struct {
	Title string
}

type bodyData struct {
	ErrorData     errData
	LoginData     loginData
	TimetableData timetableData
	GradesData    gradesData
	ResourceData  resourceData
	ResData       resData
	TasksData     tasksData
	TaskData      taskData
}

type userData struct {
	Name string
}

// Error Page

type errData struct {
	Heading  string
	Message  string
	InfoLink string
}

// Login page

type loginData struct {
	Failed   bool
	Redirect string
}

// Timetable

type timetableData struct {
	CurrentDay int
	Days       []ttDay
}

type ttDay struct {
	Day     string
	Lessons []ttLesson
}

type ttLesson struct {
	Class         string
	FormattedTime string
	Duration      string
	Height        float64
	TopOffset     float64
	Room          string
	Teacher       string
	Notice        string
	Color         string
	BGColor       string
}

// Resources list

type resData struct {
	Heading string
	Classes []resClass
}

type resClass struct {
	Name     string
	ResItems []resItem
}

type resItem struct {
	Id       string
	Name     string
	Platform string //e.g. daymap, gclass
	Posted   string
	URL      string
}

// Individual resource

type resourceData struct {
	Name        string
	Class       string
	URL         string
	Desc        template.HTML
	Posted      string
	HasResLinks bool
	ResLinks    map[string]string
	Platform    string
	Id          string
}

// Tasks list

type taskItem struct {
	Id       string
	Name     string
	Platform string
	Class    string
	DueDate  string
	Posted   string
	Grade    string
	URL      string
}

type taskType struct {
	Name     string
	NoteType string
	Tasks    []taskItem
}

type tasksData struct {
	Heading   string
	TaskTypes []taskType
}

// Individual task

type taskData struct {
	Id           string
	Name         string
	Platform     string
	Class        string
	URL          string
	IsDue        bool
	DueDate      string
	Submitted    bool
	Desc         template.HTML
	HasResLinks  bool
	ResLinks     map[string]string
	HasWorkLinks bool
	WorkLinks    map[string]string
	HasUpload    bool
	Comment      template.HTML
	Graded       bool
	TaskGrade    taskGrade
}

type taskGrade struct {
	Grade   string
	Mark    string
	Color   string
	BGColor string
}

// Grades

type gradesData struct {
	Heading string
	Tasks   []taskItem
}

var loginPageData = pageData{
	PageType: "login",
	Head: headData{
		Title: "Login",
	},
	Body: bodyData{
		LoginData: loginData{
			Failed: false,
		},
	},
}
