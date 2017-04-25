package main

import (
	"bufio"
	"errors"
	"fmt"

	"log"

	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PaulRaUnite/opti-transport"
	"github.com/jroimartin/gocui"
	"github.com/sqweek/dialog"
)

var (
	viewArr  = []string{"load", "precision", "process", "output"}
	viewFunc = []func(*gocui.View, *gocui.Gui) error{
		loadFunc,
		precisionFunc,
		processFunc,
		outputFunc,
	}
	active    = 0
	filename  string
	precision int
)

func loadFunc(_ *gocui.View, _ *gocui.Gui) error {
	var err error
	filename, err = dialog.File().Load()
	if err != nil {
		return err
	}
	return nil
}

var (
	errNegativePrecision = errors.New("negative precision")
)

func precisionFunc(v *gocui.View, _ *gocui.Gui) error {
	buf := v.Buffer()
	var err error = nil
	ss := strings.Split(buf[:len(buf)-1], " ")
	for _, str := range ss {
		precision, err = strconv.Atoi(str)
		if err != nil {
			continue
		} else if precision < 0 {
			err = errNegativePrecision
			continue
		}
		return nil
	}
	return err
}

func processFunc(_ *gocui.View, g *gocui.Gui) error {
	errG := action(g)
	debug, err := g.View("debug")
	if err != nil {
		return err
	}
	if errG != nil {
		debug.Clear()
		debug.Write([]byte(errG.Error()))
	}
	return nil
}

func outputFunc(_ *gocui.View, _ *gocui.Gui) error {
	return nil
}

func setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		return nil, err
	}
	return g.SetViewOnTop(name)
}
func nextView(g *gocui.Gui, v *gocui.View) error {
	nextIndex := (active + 1) % len(viewArr)
	name := viewArr[nextIndex]
	f := viewFunc[active]

	debug, err := g.View("debug")
	if err != nil {
		return err
	}
	v, err = g.View(viewArr[active])
	if err != nil {
		debug.Clear()
		debug.Write([]byte(err.Error()))
		return err
	}
	if err = f(v, g); err != nil {
		debug.Clear()
		debug.Write([]byte(err.Error()))
		return nil
	}

	if _, err := setCurrentViewOnTop(g, name); err != nil {
		return err
	}

	active = nextIndex
	return nil
}

func layout(g *gocui.Gui) error {
	if v, err := g.SetView("view1", 0, 0, 30, 3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "keybindings"
		v.Write([]byte("^C:     exit Arrows: navigate\nEnter:  next"))
	}
	if v, err := g.SetView("load", 31, 0, 41, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		if _, err = setCurrentViewOnTop(g, "load"); err != nil {
			return err
		}
		v.Write([]byte("load file"))
	}
	if v, err := g.SetView("precision", 42, 0, 54, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "precision"
		v.Editable = true
		v.Write([]byte("4"))
	}
	if v, err := g.SetView("process", 55, 0, 65, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Write([]byte("calculate"))
	}
	maxX, maxY := g.Size()
	if v, err := g.SetView("output", 0, 4, maxX-1, maxY-4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		//v.Wrap = true
		v.Title = "result"
		v.Editable = true
		v.Editor = gocui.EditorFunc(func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {})
	}
	if v, err := g.SetView("costfunc", 0, maxY-3, 20, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "cost function"
	}
	if v, err := g.SetView("time", 21, maxY-3, 40, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "time"
	}
	if v, err := g.SetView("debug", 41, maxY-3, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "debug"
	}
	return nil
}
func cursorDown(_ *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorUp(_ *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}
func cursorLeft(_ *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		ox, oy := v.Origin()
		if err := v.SetCursor(cx-1, cy); err != nil && ox > 0 {
			if err := v.SetOrigin(ox-1, oy); err != nil {
				return err
			}
		}
	}
	return nil
}
func cursorRight(_ *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx+1, cy); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox+1, oy); err != nil {
				return err
			}
		}
	}
	return nil
}

func quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Highlight = true
	g.Cursor = true
	g.SelFgColor = gocui.ColorGreen

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, nextView); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("output", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("output", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("output", gocui.KeyArrowLeft, gocui.ModNone, cursorLeft); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("output", gocui.KeyArrowRight, gocui.ModNone, cursorRight); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func processFile(filename string) ([][]float64, []float64, []float64, error) {
	//open file
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, nil, err
	}
	//get lines and split by " "
	var strMatrix [][]string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		strMatrix = append(strMatrix, strings.Split(scanner.Text(), " "))
	}
	//if something goes wrong
	if err := scanner.Err(); err != nil {
		return nil, nil, nil, err
	}
	//convert string into float64
	var all [][]float64
	for _, subarr := range strMatrix {
		var temp []float64
		//main matrix and products
		for _, value := range subarr {
			if len(value) == 0 {
				continue
			}
			product, err := strconv.ParseFloat(strings.Replace(value, ",", ".", 1), 64)
			if err != nil {
				return nil, nil, nil, err
			}
			temp = append(temp, product)
		}
		all = append(all, temp)
	}
	var taxes [][]float64
	var productCenters []float64
	consumptionCenters := all[len(all)-1]
	for _, subarr := range all[:len(all)-1] {
		taxes = append(taxes, subarr[:len(subarr)-1])
		productCenters = append(productCenters, subarr[len(subarr)-1])
	}
	return taxes, productCenters, consumptionCenters, nil
}

func action(gui *gocui.Gui) error {
	taxes, products, sales, err := processFile(filename)
	if err != nil {
		return err
	}

	cond, err := opti_transport.NewCondition(products, sales, taxes, precision)
	if err != nil {
		return err
	}
	point1 := time.Now()
	presolve, err := cond.MinimalTaxesMethod()
	if err != nil {
		return err
	}
	presolve.Optimize()
	point2 := time.Now()
	outV, err := gui.View("output")
	if err != nil {
		return err
	}
	outV.Clear()
	outV.Write([]byte(presolve.WellPrintedString()))
	costV, err := gui.View("costfunc")
	if err != nil {
		return err
	}
	debug, err := gui.View("debug")
	if err != nil {
		return err
	}
	debug.Clear()
	debug.Write([]byte(opti_transport.Export))
	costV.Clear()
	costV.Write([]byte(fmt.Sprintf("%f", presolve.CostFunc())))
	timeV, err := gui.View("time")
	if err != nil {
		return err
	}
	timeV.Clear()
	timeV.Write([]byte(point2.Sub(point1).String()))
	return nil
}
