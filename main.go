package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

type output struct {
	name   string
	status string // connected/disconnected/configured
	res    resolution
}

type resolution struct {
	height int
	width  int
	x      int
	y      int
}

func parseResPos(res string) *resolution {
	var width, height, x, y int
	n, err := fmt.Sscanf(res, "%dx%d+%d+%d", &width, &height, &y, &x)
	if err == nil {
		if n != 4 {
			panic(fmt.Errorf("Did not read 4 fields, read %d", n))
		}

		return &resolution{
			height: height,
			width:  width,
			x:      x,
			y:      y,
		}
	}
	return nil
}

func parseRes(res string) *resolution {
	var width, height int
	n, err := fmt.Sscanf(res, "%dx%d", &width, &height)
	if err == nil {
		if n != 2 {
			panic(fmt.Errorf("Did not read 4 fields, read %d", n))
		}

		return &resolution{
			height: height,
			width:  width,
		}
	}
	return nil
}

func getOutputs() []output {
	cmd := exec.Command("xrandr", "--query")

	r, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(r)

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	outputs := []output{}

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if line[0] != ' ' {
			name := fields[0]
			if name == "Screen" {
				continue
			}
			status := fields[1]
			var res resolution
			if status == "connected" {
				rawRes := fields[2]
				if fields[2] == "primary" {
					rawRes = fields[3]
				}

				parsedRes := parseResPos(rawRes)

				if parsedRes != nil {
					status = "configured"
					res = *parsedRes
				}

			}
			outputs = append(outputs, output{
				name:   name,
				status: status,
				res:    res,
			})
		} else if len(outputs) != 0 {
			lastOutput := &outputs[len(outputs)-1]
			if lastOutput.status == "connected" {
				if strings.Contains(line, "+") { // primary resolution
					res := parseRes(fields[0])
					if res != nil {
						lastOutput.res = *res
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	if err := cmd.Wait(); err != nil {
		panic(err)
	}
	return outputs
}

func main() {
	outputs := getOutputs()

	connectedOutputs := []output{}
	for _, output := range outputs {
		if output.status != "disconnected" {
			connectedOutputs = append(connectedOutputs, output)
		}
	}

	sort.Slice(connectedOutputs, func(i, j int) bool {
		return connectedOutputs[i].name > connectedOutputs[j].name
	})

	reses := make([]string, len(connectedOutputs))
	for i, output := range connectedOutputs {
		reses[i] = fmt.Sprintf("%dx%d", output.res.width, output.res.height)
	}
	resMatch := strings.Join(reses, " ")
	switch resMatch {
	case "1920x1080 3440x1440":
		laptop := connectedOutputs[0]
		wide := connectedOutputs[1]
		if laptop.status == "connected" {
			fmt.Println([]string{"xrandr", "--output", laptop.name, "--auto"})
		}
		if wide.status == "connected" {
			cmd := exec.Command("xrandr", "--output", wide.name, "--auto", "--above", laptop.name)
			if err := cmd.Run(); err != nil {
				panic(err)
			}
		}
	case "1920x1080 2560x1440 2560x1440":
		fmt.Printf("connectedOutputs = %+v\n", connectedOutputs)
		laptop := connectedOutputs[0]
		middle := connectedOutputs[1]
		right := connectedOutputs[2]
		if laptop.status == "connected" {
			fmt.Println([]string{"xrandr", "--output", laptop.name, "--auto"})
		}
		if middle.status == "connected" {
			fmt.Println("Configuring middle monitor")
			cmd := exec.Command("xrandr", "--output", middle.name, "--auto", "--above", laptop.name)
			if out, err := cmd.CombinedOutput(); err != nil {
				fmt.Println(string(out))
				panic(err)
			}
		}
		if right.status == "connected" {
			time.Sleep(1 * time.Second)
			fmt.Println("Configuring right monitor")
			cmd := exec.Command("xrandr", "--output", right.name, "--auto", "--right-of", middle.name, "--rotate", "left")
			if out, err := cmd.CombinedOutput(); err != nil {
				fmt.Println(string(out))
				panic(err)
			}
		}
	case "1920x1080":
		for _, output := range outputs {
			if output.status == "disconnected" {
				cmd := exec.Command("xrandr", "--output", output.name, "--off")
				if err := cmd.Run(); err != nil {
					panic(err)
				}
			}
		}
	default:
		fmt.Println("Unsupported setup")
		os.Exit(1)
	}

	for _, output := range connectedOutputs {
		fmt.Println(output.name)
	}
}
