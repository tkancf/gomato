package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/0xAX/notificator"
	"github.com/gin-gonic/gin"
	"github.com/nsf/termbox-go"
	"github.com/sethgrid/multibar"
	"github.com/tkancf/curse"
	"github.com/urfave/cli"
)

type Task struct {
	Name    string `json:"name"`
	State   string `json:"state"`
	Date    string `json:"date"`
	Time    string `json:"time"`
	Elapsed int    `json:"Elapsed"`
}
type Tasks []Task

var data Task

type File struct {
	TemplateDir string
	Json        string
}

var files File

var notify *notificator.Notificator

func main() {
	files.Json = getJsonFile()
	files.TemplateDir = getTemplateDir()
	notify = notificator.New(notificator.Options{
		DefaultIcon: "icon/default.png",
		AppName:     "gomato",
	})
	err := termbox.Init()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	commands()
}

func commands() {
	// init cli app
	app := cli.NewApp()
	app.Name = "gomato"
	app.Usage = "Pomodoro Timer in Console written in golang."
	app.Version = "0.1.0"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Taikan Yamaguchi",
			Email: "tkncf789@gmail.com",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "timer",
			Aliases: []string{"t"},
			Usage:   "Start pomodoro timer\n\tYou can set time [task] [short break] [long break]\n\t(default time is 25 5 15)",
			Action:  timerAction,
		},
		{
			Name:    "server",
			Aliases: []string{"s"},
			Usage:   "Start server to show tasks \n\thttp://localhost:3000/",
			Action:  serverAction,
		},
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "List tasks",
			Action:  listAction,
		},
	}
	app.After = func(c *cli.Context) error {
		return nil
	}
	app.Run(os.Args)
}

func getTemplateDir() string {
	gopath := os.Getenv("GOPATH")
	tmp := filepath.Join(gopath, "src", "github.com", "tkancf", "gomato", "templates")
	return tmp
}

func getJsonFile() string {
	gopath := os.Getenv("GOPATH")
	json := filepath.Join(gopath, "src", "github.com", "tkancf", "gomato", "data", "data.json")
	return json
}

func timerAction(c *cli.Context) {
	if !fileExists(files.Json) {
		ioutil.WriteFile(files.Json, []byte(""), os.ModePerm)
	}
	fmt.Println("'q' Key to Quit")
	cu, _ := curse.New()
	c0 := "25"
	c1 := "5"
	c2 := "15"
	data.Name = "unknown"

	if c.NArg() == 4 {
		data.Name = c.Args().Get(0)
		c0 = c.Args().Get(1)
		c1 = c.Args().Get(2)
		c2 = c.Args().Get(3)
	} else if c.NArg() == 1 {
		data.Name = c.Args().Get(0)
	}
	t0, _ := strconv.Atoi(c0)
	t1, _ := strconv.Atoi(c1)
	t2, _ := strconv.Atoi(c2)
	for {
		for i := 1; i < 4; i++ {
			data.State = "task"
			timer(t0, getTaskString(data.State, t0, i))
			cu.MoveUp(1)
			cu.EraseCurrentLine()
			data.Elapsed = t0
			notify.Push("gomato", data.Name+" task is finished!", "/home/user/icon.png", notificator.UR_CRITICAL)
			checkContinue()

			data.State = "break"
			timer(t1, getTaskString(data.State, t1, i))
			cu.MoveUp(1)
			cu.EraseCurrentLine()
			data.Elapsed = t1
			notify.Push("gomato", "break time is finished", "/home/user/icon.png", notificator.UR_CRITICAL)
			checkContinue()
		}
		data.State = "task"
		timer(t0, getTaskString(data.State, t0, 4))
		cu.MoveUp(1)
		cu.EraseCurrentLine()
		data.Elapsed = t0
		notify.Push("gomato", data.Name+" task is finished!", "/home/user/icon.png", notificator.UR_CRITICAL)
		checkContinue()

		data.State = "lbreak"
		timer(t2, getTaskString(data.State, t2, 0))
		data.Elapsed = t2
		notify.Push("gomato", "long break time is finished", "/home/user/icon.png", notificator.UR_CRITICAL)
		checkContinue()
	}
}

// t = 経過時間, s = 左に表示するステータス
func timer(t int, s string) {
	t = t * 60
	data.Time = getTime()
	data.Date = getDate()
	progressBars, _ := multibar.New()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	barProgress := progressBars.MakeBar(t, s)
	go handleKeyEvent()
	go progressBars.Listen()
	go func() {
		for i := 0; i <= t; i++ {
			data.Elapsed = i
			barProgress(i)
			time.Sleep(time.Second * 1)
		}
		wg.Done()
	}()
	wg.Wait()
}

func getTime() string {
	t := time.Now()
	const layout = "15:04"
	ts := t.Format(layout)
	return ts
}

func getDate() string {
	t := time.Now()
	const layout = "2006-01-02"
	ts := t.Format(layout)
	return ts
}

// タイマー終了後の継続確認
func checkContinue() {
	fmt.Printf("'ENTER' to Continue")
	saveData()
	switch ev := termbox.PollEvent(); ev.Type {
	}
}

func getTaskString(s string, t int, i int) string {
	ts := strconv.Itoa(t)
	is := strconv.Itoa(i)
	str := ""
	if s == "task" {
		data.State = s
		str = data.Name + "[task time]" + "(" + is + "/4)<" + ts + "Minutes>"
	} else if s == "break" {
		data.State = s
		str = "[break time]" + "(" + is + "/4)<" + ts + "Minutes>"
	} else if s == "lbreak" {
		data.State = s
		str = "[long break time]<" + ts + "Minutes>"
	} else {
		return str
	}
	return str
}

func handleKeyEvent() {
loop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc {
				os.Exit(1)
			} else if ev.Ch == 'q' {
				saveFile(saveData())
				break loop
			}
		}
	}
	os.Exit(1)
}

func handleKeyEventNoSave() {
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc {
				os.Exit(1)
			} else if ev.Ch == 'q' {
				os.Exit(1)
			}
		}
	}
}

func saveData() Tasks {
	var s Tasks
	if len(s) == 0 {
		s = getJson(files.Json)
	}
	s = append(s, Task{Name: data.Name, State: data.State, Date: data.Date, Time: data.Time, Elapsed: data.Elapsed})
	return s
}

func saveFile(s Tasks) {
	writeJson(files.Json, &s)
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func getJson(p string) Tasks {
	var t Tasks
	raw, err := ioutil.ReadFile(p)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	json.Unmarshal(raw, &t)
	return t
}

func writeJson(p string, t *Tasks) {
	json, err := json.Marshal(t)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	ioutil.WriteFile(p, json, os.ModePerm)
}

func serverAction(c *cli.Context) {
	go handleKeyEventNoSave()
	keys, values := getTaskTimeArray(files.Json)
	fmt.Println("Server start\nAccess to http://localhost:3000/\n'q' Key to Quit")
	router := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	router.LoadHTMLGlob(filepath.Join(files.TemplateDir, "*"))
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"keys":   keys,
			"values": values,
		})
	})
	router.Run(":3000")
}

func getTaskTimeArray(path string) ([]string, []int) {
	jsonData := getJson(path)
	m := map[string]int{}
	keys := []string{}
	values := []int{}
	for i := 0; i < len(jsonData); i++ {
		_, ok := m[jsonData[i].Name]
		if ok == false {
			m[jsonData[i].Name] = jsonData[i].Elapsed
		} else {
			m[jsonData[i].Name] = m[jsonData[i].Name] + jsonData[i].Elapsed
		}
	}
	for k, v := range m {
		keys = append(keys, k)
		values = append(values, v)
	}
	return keys, values
}

func listAction(c *cli.Context) {
	jsonData := getJson(files.Json)
	s := []string{}
	for i := 0; i < len(jsonData); i++ {
		if !contains(s, jsonData[i].Name) {
			s = append(s, jsonData[i].Name)
		}
	}
	for i := 0; i < len(s); i++ {
		fmt.Println(s[i])
	}
}
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
