package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Cristofori/kmud/types"
)

const decorator = "-=-=-"

type Menu struct {
	actions     []action
	title       string
	exit        bool
	exitHandler func()
}

func ExecMenu(title string, comm types.Communicable, build func(*Menu)) {
	pageIndex := 0
	pageCount := 1
	filter := ""

	for {
		var menu Menu
		menu.title = title
		menu.exit = false
		build(&menu)

		pageIndex = Bound(pageIndex, 0, pageCount-1)
		pageCount = menu.Print(comm, pageIndex, filter)
		filter = ""

		prompt := ""
		if pageCount > 1 {
			prompt = fmt.Sprintf("Page %v of %v (<, >, <<, >>)\r\n> ", pageIndex+1, pageCount)
		} else {
			prompt = "> "
		}

		input := comm.GetInput(types.Colorize(types.ColorWhite, prompt))

		if input == "" {
			if menu.exitHandler != nil {
				menu.exitHandler()
			}
			return
		}

		if input == ">" {
			pageIndex++
		} else if input == "<" {
			pageIndex--
		} else if input == ">>" {
			pageIndex = pageCount - 1
		} else if input == "<<" {
			pageIndex = 0
		} else if input[0] == '/' {
			filter = input[1:]
		} else {
			action := menu.getAction(input)

			if action.handler != nil {
				action.handler()
				if menu.exit {
					if menu.exitHandler != nil {
						menu.exitHandler()
					}
					return
				}
			} else if input != "?" && input != "help" {
				comm.WriteLine(types.Colorize(types.ColorRed, "Invalid selection"))
			}
		}
	}
}

type action struct {
	key     string
	text    string
	data    types.Id
	handler func()
}

func (self *Menu) AddAction(key string, text string, handler func()) {
	if self.HasAction(key) {
		panic(fmt.Sprintf("Duplicate action added to menu: %s %s", key, text))
	}

	self.actions = append(self.actions,
		action{key: strings.ToLower(key),
			text:    text,
			handler: handler,
		})
}

func (self *Menu) AddActionI(index int, text string, handler func()) {
	self.AddAction(strconv.Itoa(index+1), text, handler)
}

func (self *Menu) SetTitle(title string) {
	self.title = title
}

func (self *Menu) Exit() {
	self.exit = true
}

func (self *Menu) OnExit(handler func()) {
	self.exitHandler = handler
}

func (self *Menu) GetData(choice string) types.Id {
	for _, action := range self.actions {
		if action.key == choice {
			return action.data
		}
	}

	return nil
}

func (self *Menu) getAction(key string) action {
	key = strings.ToLower(key)

	for _, action := range self.actions {
		if action.key == key {
			return action
		}
	}
	return action{}
}

func (self *Menu) HasAction(key string) bool {
	action := self.getAction(key)
	return action.key != ""
}

func filterActions(actions []action, filter string) []action {
	var filtered []action

	for _, action := range actions {
		if FilterItem(action.text, filter) {
			filtered = append(filtered, action)
		}
	}

	return filtered
}

func (self *Menu) Print(comm types.Communicable, page int, filter string) int {
	border := types.Colorize(types.ColorWhite, decorator)
	title := types.Colorize(types.ColorBlue, self.title)
	header := fmt.Sprintf("%s %s %s", border, title, border)

	if filter != "" {
		header = fmt.Sprintf("%s (/%s)", header, filter)
	}

	comm.WriteLine(header)

	filteredActions := filterActions(self.actions, filter)
	options := make([]string, len(filteredActions))

	for i, action := range filteredActions {
		index := strings.Index(strings.ToLower(action.text), action.key)

		actionText := ""

		if index == -1 {
			actionText = fmt.Sprintf("%s[%s%s%s]%s%s",
				types.ColorDarkBlue,
				types.ColorBlue,
				strings.ToUpper(action.key),
				types.ColorDarkBlue,
				types.ColorWhite,
				action.text)
		} else {
			keyLength := len(action.key)
			actionText = fmt.Sprintf("%s%s[%s%s%s]%s%s",
				action.text[:index],
				types.ColorDarkBlue,
				types.ColorBlue,
				action.text[index:index+keyLength],
				types.ColorDarkBlue,
				types.ColorWhite,
				action.text[index+keyLength:])
		}

		options[i] = fmt.Sprintf("  %s", actionText)
	}

	width, height := comm.GetWindowSize()
	pages := Paginate(options, width, height/2)

	if len(options) == 0 && filter != "" {
		comm.WriteLine("No items match your search")
	} else {
		comm.Write(pages[page])
	}

	return len(pages)
}
