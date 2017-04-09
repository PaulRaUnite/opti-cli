package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/PaulRaUnite/opti-transport"
	"gopkg.in/urfave/cli.v2"
)

func main() {
	app := cli.App{
		Name:        "opti-cli",
		Usage:       "application for solving transportation problem",
		UsageText:   "opti-cli SRC",
		Version:     "0.0.1",
		Action:      action,
		HideVersion: true,
	}
	app.Run(os.Args)
}

var (
	errorSrcArg = errors.New("SRC field is empty")
)

func action(c *cli.Context) error {
	//get src
	srcArg := c.Args().First()
	if srcArg == "" {
		return errorSrcArg
	}
	//open file
	file, err := os.Open(srcArg)
	if err != nil {
		return err
	}
	//get lines and split by " "
	var strMatrix [][]string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		strMatrix = append(strMatrix, strings.Split(scanner.Text(), " "))
	}
	//if something goes wrong
	if err := scanner.Err(); err != nil {
		return err
	}
	//convert string into int
	var taxes [][]float64
	var productPoints, salePoints []float64
	for i := 0; i < len(strMatrix); i++ {
		var temp []float64
		if i == len(strMatrix)-1 {
			for j := 0; j < len(strMatrix[i]); j++ {
				if len(strMatrix[i][j]) == 0 {
					continue
				}
				value, err := strconv.ParseFloat(strMatrix[i][j], 64)
				if err != nil {
					return err
				}

				temp = append(temp, value)
			}
			salePoints = temp
			break
		}
		for j := 0; j < len(strMatrix[i]); j++ {
			if len(strMatrix[i][j]) == 0 {
				continue
			}
			value, err := strconv.ParseFloat(strMatrix[i][j], 64)
			if err != nil {
				return err
			}
			if j == len(strMatrix[i])-1 {
				productPoints = append(productPoints, value)
				break
			}
			temp = append(temp, value)
		}
		taxes = append(taxes, temp)
	}

	cond, err := opti_transport.NewCondition(productPoints, salePoints, taxes)
	if err != nil {
		return err
	}
	presolve, err := cond.MinimalTaxesMethod()
	if err != nil {
		fmt.Println(presolve)
		return err
	}
	fmt.Println("Minimal taxes method")
	fmt.Println(presolve.WellPrintedString())
	fmt.Println("Cost function:", presolve.CostFunc())

	fmt.Println("Optimizing...")
	presolve.Optimize()

	fmt.Println(presolve.WellPrintedString())
	fmt.Println("Cost function:", presolve.CostFunc())
	return nil
}
