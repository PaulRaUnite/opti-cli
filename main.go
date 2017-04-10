package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"time"

	"github.com/PaulRaUnite/opti-transport"
	log "github.com/Sirupsen/logrus"
	"gopkg.in/urfave/cli.v2"
)

func init() {
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{ForceColors: true, DisableSorting: true})
}

func main() {
	app := cli.App{
		Name:        "opti-cli",
		Usage:       "application for solving transportation problem",
		UsageText:   "opti-cli SRC",
		Version:     "0.0.1",
		Action:      action,
		HideVersion: true,
		ErrWriter:   (&log.Logger{}).WriterLevel(log.ErrorLevel),
	}
	app.Run(os.Args)
}

var (
	errorSrcArg = errors.New("SRC field is empty")
)

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
	var taxes [][]float64
	var productPoints, salePoints []float64
	for i, subarr := range strMatrix {
		var temp []float64
		//sales
		if i == len(strMatrix)-1 {
			for _, value := range subarr {
				if len(value) == 0 {
					continue
				}
				value, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return nil, nil, nil, err
				}

				temp = append(temp, value)
			}
			salePoints = temp
			break
		}
		//main matrix and products
		for j, value := range subarr {
			if len(value) == 0 {
				continue
			}
			product, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, nil, nil, err
			}
			if j == len(subarr)-1 {
				productPoints = append(productPoints, product)
				break
			}
			temp = append(temp, product)
		}
		taxes = append(taxes, temp)
	}
	return taxes, productPoints, salePoints, nil
}

func action(c *cli.Context) error {
	//get src
	filename := c.Args().First()
	if filename == "" {
		return errorSrcArg
	}
	taxes, products, sales, err := processFile(filename)
	if err != nil {
		return err
	}

	cond, err := opti_transport.NewCondition(products, sales, taxes, 4)
	if err != nil {
		return err
	}
	point1 := time.Now()
	presolve, err := cond.MinimalTaxesMethod()
	point2 := time.Now()
	if err != nil {
		fmt.Println(presolve)
		return err
	}
	log.Info("minimal taxes method...")
	fmt.Println(presolve.WellPrintedString())
	log.Info("cost function := ", presolve.CostFunc())
	log.Info("time := ", point2.Sub(point1).String())

	log.Info("optimizing...")
	point1 = time.Now()
	presolve.Optimize()
	point2 = time.Now()

	fmt.Println(presolve.WellPrintedString())
	log.Info("cost function := ", presolve.CostFunc())
	log.Info("time := ", point2.Sub(point1).String())
	return nil
}
