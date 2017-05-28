package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
	"github.com/sethgrid/curse"
	"github.com/sethgrid/multibar"
	"github.com/urfave/cli"
)

func main() {
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
	app.Usage = "pomodoro timer"
	app.Commands = []cli.Command{
		{
			Name:    "timer",
			Aliases: []string{"t"},
			Usage:   "Start pomodoro timer\n\tYou can set time [task] [short break] [long break]\n\t(default time is 25 5 15)",
			Action:  timerAction,
		},
	}
	app.After = func(c *cli.Context) error {
		fmt.Println("END")
		return nil
	}
	app.Run(os.Args)
}

func timerAction(c *cli.Context) {
	fmt.Println("'ESC' Key to Quit")
	cu, _ := curse.New()
	c0 := "25"
	c1 := "5"
	c2 := "15"
	if c.NArg() == 3 {
		c0 = c.Args().Get(0)
		c1 = c.Args().Get(1)
		c2 = c.Args().Get(2)
	}
	x, _ := strconv.Atoi(c0)
	y, _ := strconv.Atoi(c1)
	z, _ := strconv.Atoi(c2)
	for i := 1; i < 4; i++ {
		str := "[task time] " + ":(" + strconv.Itoa(i) + "/4)" + c1 + " Second"
		timer(x, str)
		cu.MoveUp(1)
		cu.EraseCurrentLine()
		checkContinue()
		str = "[break time] " + ":(" + strconv.Itoa(i) + "/4)" + c1 + " Second"
		timer(y, str)
		cu.MoveUp(1)
		cu.EraseCurrentLine()
		checkContinue()
	}
	str := "[task time] " + ":(4/4)" + c0 + " Second"
	timer(x, str)
	cu.MoveUp(1)
	cu.EraseCurrentLine()
	checkContinue()
	timer(z, "[long break time] "+c2+" Second")
}

func checkContinue() {
	fmt.Printf("'ENTER' to Continue")
	switch ev := termbox.PollEvent(); ev.Type {
	case termbox.EventKey:
		if ev.Ch == 'q' {
			os.Exit(1)
		} else if ev.Key == termbox.KeyEnter {
			// No Action
		}
	}
}

// t = 経過時間, s = 左に表示するステータス
func timer(t int, s string) {
	progressBars, _ := multibar.New()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	barProgress := progressBars.MakeBar(t, s)
	go handleKeyEvent()
	go progressBars.Listen()
	go func() {
		for i := 0; i <= t; i++ {
			barProgress(i)
			time.Sleep(time.Second * 1)
		}
		wg.Done()
	}()
	wg.Wait()
}

func handleKeyEvent() {
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
