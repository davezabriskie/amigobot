package mfp

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/olekukonko/tablewriter"
)

//Commands that trigger a message handle
const (
	Cals   = "?cals"
	Macros = "?macros"
)

//HandleCalsMessage is triggered by the Cals command and sends a table
//of the foods and calories of the day
func HandleCalsMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	handleMfpMessage(s, m, Cals, newCaloriesMessage)
}

//HandleMacrosMessage is triggered by the Macros command and sends a table
//of the macro grams and percentages of the day
func HandleMacrosMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	handleMfpMessage(s, m, Macros, newMacrosMessage)
}

func handleMfpMessage(s *discordgo.Session, m *discordgo.MessageCreate, cmd string, fn func(string) (string, error)) {
	content := strings.SplitN(m.Content, " ", 2)
	if len(content) < 2 {
		s.ChannelMessageSend(m.ChannelID, "No username provided")
		return
	}
	username := content[1]
	message, err := fn(username)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}
	s.ChannelMessageSend(m.ChannelID, message)
}

func newCaloriesMessage(username string) (string, error) {
	d, err := newDiary(username)
	if err != nil {
		return "", err
	}
	message := calsMessage(d)
	if len(message) > 2000 {
		totalStart := len(message) - 102
		return "```" + message[totalStart:], nil
	}
	return message, nil
}

type macroPercentages struct {
	carbs   int
	protein int
	fat     int
}

func newMacrosMessage(username string) (string, error) {
	d, err := newDiary(username)
	if err != nil {
		return "", err
	}
	table, buffer := newTable([]string{"Macros", "Grams", "Percent"}, 10)
	m, err := newMacroPercentages(d)
	if err != nil {
		return "", nil
	}
	table.Append([]string{"Carbs", d.total.carbs, fmt.Sprintf("%d%%", m.carbs)})
	table.Append([]string{"Protein", d.total.protein, fmt.Sprintf("%d%%", m.protein)})
	table.Append([]string{"Fat", d.total.fat, fmt.Sprintf("%d%%", m.fat)})
	table.Render()
	return "```" + buffer.String() + "```", nil
}

func calsMessage(diary *diary) string {
	table, buffer := newTable([]string{"Food", "Calories"}, 17)
	for _, v := range formatTableData(diary) {
		table.Append(v)
	}
	table.Render()
	return "```" + buffer.String() + "```"
}

func newTable(headers []string, colWidth int) (*tablewriter.Table, *bytes.Buffer) {
	buffer := new(bytes.Buffer)
	table := tablewriter.NewWriter(buffer)
	table.SetColWidth(colWidth)
	table.SetHeader(headers)
	table.SetRowLine(true)
	return table, buffer
}

func formatTableData(d *diary) [][]string {
	var data [][]string
	meals := []string{"Breakfast", "Lunch", "Dinner", "Snacks"}
	for _, m := range meals {
		if f, ok := d.meals[m]; ok {
			data = append(data, []string{strings.ToUpper(m), ""})
			data = addFoods(data, f)
		}
	}
	data = append(data, []string{"Total", d.total.calories})
	return data
}

func addFoods(data [][]string, foods []food) [][]string {
	for _, food := range foods {
		name := formatFoodName(food.name)
		data = append(data, []string{name, food.calories})
	}
	return data
}

func formatFoodName(name string) string {
	strippedBrandSlice := strings.SplitN(name, "- ", 2)
	stripped := strippedBrandSlice[len(strippedBrandSlice)-1]
	if len(stripped) > 32 {
		return stripped[:30] + "..."
	}
	return stripped
}

func newMacroPercentages(d *diary) (macroPercentages, error) {
	m := macroPercentages{}
	carbs, cErr := parseMacro(d.total.carbs)
	protein, pErr := parseMacro(d.total.protein)
	fat, fErr := parseMacro(d.total.fat)
	if cErr != nil || pErr != nil || fErr != nil {
		return m, errors.New("Error parsing macros")
	}
	total := carbs + protein + fat
	m = macroPercentages{
		carbs:   (100 * carbs) / total,
		protein: (100 * protein) / total,
		fat:     (100 * fat) / total,
	}
	return m, nil
}
func parseMacro(macro string) (int, error) {
	return strconv.Atoi(macro[:len(macro)-1])
}
