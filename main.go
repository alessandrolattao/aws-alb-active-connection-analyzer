package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gizak/termui/v3"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type request struct {
	RequestCreationTime time.Time `json:"RequestCreationTime"`
	ElbStatusCode       int       `json:"ElbStatusCode"`
	TargetStatusCode    int       `json:"TargetStatusCode"`
	Time                time.Time `json:"Time"`
}

func main() {

	data := []float64{}
	labels := []string{}
	colors := []termui.Color{}

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	uiEvents := ui.PollEvents()
	termWidth, termHeight := ui.TerminalDimensions()

	bc := widgets.NewBarChart()
	bc.Title = "Open connections"
	bc.BarWidth = 12
	bc.MaxVal = 10
	bc.LabelStyles = []ui.Style{ui.NewStyle(ui.ColorWhite)}
	bc.NumStyles = []ui.Style{ui.NewStyle(ui.ColorWhite)}
	onScreenElements := int(termWidth / 100 * 70 / bc.BarWidth)

	p := widgets.NewParagraph()
	p.Title = "Open connections JSON"

	grid := ui.NewGrid()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		ui.NewRow(1.0,
			ui.NewCol(0.7, bc),
			ui.NewCol(0.3, p),
		),
	)

	csvFile, err := os.Open("alb_requests.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	if err != nil {
		fmt.Println(err)
	}

	reader.Comma = ';'

	csvLines, err := reader.ReadAll()
	if err != nil {
		fmt.Println(err)
	}

	openConnections := []request{}

	for counter, line := range csvLines {

		if counter < 1 {
			continue
		}

		requestCreationTime, err := time.Parse("2006-01-02T15:04:05.000000Z", line[0])
		if err != nil {
			panic(err)
		}

		durationTime, err := time.Parse("2006-01-02T15:04:05.000000Z", line[3])
		if err != nil {
			panic(err)
		}

		elbStatusCode, err := strconv.Atoi(line[1])
		if err != nil {
			panic(err)
		}

		targetStatusCode, err := strconv.Atoi(line[2])
		if err != nil {
			panic(err)
		}

		riga := request{
			RequestCreationTime: requestCreationTime,
			ElbStatusCode:       elbStatusCode,
			TargetStatusCode:    targetStatusCode,
			Time:                durationTime,
		}

		activeConnections := []request{}
		active502 := false

		// check for ended calls
		for _, connection := range openConnections {
			// only active connections
			if connection.Time.After(riga.RequestCreationTime) {
				activeConnections = append(activeConnections, connection)
				if connection.ElbStatusCode == 502 {
					active502 = true
				}
			}
		}

		openConnections = append(activeConnections, riga)

		j, err := json.MarshalIndent(openConnections, "", "   ")
		if err != nil {
			panic(err)
		}
		p.Text = string(j)

		data = append(data, float64(len(openConnections)))
		labels = append(labels, riga.RequestCreationTime.Format("15:04:05.000"))

		color := ui.ColorBlue

		if active502 {
			color = ui.ColorMagenta
		}

		if riga.TargetStatusCode == 502 {
			color = ui.ColorRed
		}

		colors = append(colors, color)

		if counter >= onScreenElements {
			data = data[1:]
			labels = labels[1:]
			colors = colors[1:]
		}

		bc.BarColors = colors
		bc.Data = data
		bc.Labels = labels

		ui.Render(grid)

		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		}
	}
}
