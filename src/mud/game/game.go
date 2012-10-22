package game

import (
	"fmt"
	"io"
	"labix.org/v2/mgo"
	"mud/database"
	"mud/utils"
	"net"
	"strings"
	"time"
)

func getToggleExitMenu(room database.Room) utils.Menu {

	onOrOff := func(direction database.ExitDirection) string {

		if room.HasExit(direction) {
			return "On"
		}

		return "Off"
	}

	menu := utils.NewMenu("Edit Exits")

	menu.AddAction("n", "[N]orth: "+onOrOff(database.DirectionNorth))
	menu.AddAction("e", "[E]ast: "+onOrOff(database.DirectionEast))
	menu.AddAction("s", "[S]outh: "+onOrOff(database.DirectionSouth))
	menu.AddAction("w", "[W]est: "+onOrOff(database.DirectionWest))
	menu.AddAction("u", "[U]p: "+onOrOff(database.DirectionUp))
	menu.AddAction("d", "[D]own: "+onOrOff(database.DirectionDown))

	return menu
}

func Exec(session *mgo.Session, conn net.Conn, character database.Character) {

	room, err := database.GetCharacterRoom(session, character)

	printString := func(data string) {
		io.WriteString(conn, data)
	}

	printLine := func(line string) {
		utils.WriteLine(conn, line)
	}

	printRoom := func() {
		printLine(room.ToString(database.ReadMode))
	}

	printRoomEditor := func() {
		printLine(room.ToString(database.EditMode))
	}

	prompt := func() string {
		return "> "
	}

	processAction := func(input string) {
		inputFields := strings.Fields(input)
		fieldCount := len(inputFields)
		action := inputFields[0]

		switch action {
		case "l":
			fallthrough
		case "look":
			if fieldCount == 1 {
				printRoom()
			} else if fieldCount == 2 {
				arg := database.StringToDirection(inputFields[1])

				if arg == database.DirectionNone {
					printLine("Nothing to see")
				} else {
					loc := room.Location.Next(arg)
					roomToSee, err := database.GetRoomByLocation(session, loc)
					if err == nil {
						printLine(roomToSee.ToString(database.ReadMode))
					} else {
						printLine("Nothing to see")
					}
				}
			}

		case "i":
			printLine("You aren't carrying anything")

		case "":
			fallthrough
		case "logout":
			return

		case "quit":
			fallthrough
		case "exit":
			printLine("Take luck!")
			conn.Close()
			panic("User quit")

		default:
			direction := database.StringToDirection(input)

			if direction != database.DirectionNone {
				if room.HasExit(direction) {
					newRoom, err := database.MoveCharacter(session, &character, direction)
					if err == nil {
						room = newRoom
						printRoom()
					} else {
						printLine(err.Error())
					}

				} else {
					printLine("You can't go that way")
				}
			} else {
				printLine("You can't do that")
			}
		}
	}

	processCommand := func(command string) {
		switch command {
		case "?":
			fallthrough
		case "help":
		case "dig":
		case "edit":
			printRoomEditor()

			for {
				input := utils.GetUserInput(conn, "Select a section to edit> ")

				switch input {
				case "":
					printRoom()
					return

				case "1":
					input = utils.GetRawUserInput(conn, "Enter new title: ")

					if input != "" {
						room.Title = input
						database.CommitRoom(session, room)
					}
					printRoomEditor()

				case "2":
					input = utils.GetRawUserInput(conn, "Enter new description: ")

					if input != "" {
						room.Description = input
						database.CommitRoom(session, room)
					}
					printRoomEditor()

				case "3":
					for {
						menu := getToggleExitMenu(room)
						choice, _ := menu.Exec(conn)

						toggleExit := func(direction database.ExitDirection) {
							enable := !room.HasExit(direction)
							room.SetExitEnabled(direction, enable)
							database.CommitRoom(session, room)
						}

						if choice == "" {
							break
						} else {
							direction := database.StringToDirection(choice)
							if direction != database.DirectionNone {
								toggleExit(direction)
							}
						}
					}

					printRoomEditor()

				default:
					printLine("Invalid selection")
				}
			}

		case "rebuild":
			input := utils.GetUserInput(conn, "Are you sure (delete all rooms and starts from scratch)? ")
			if input[0] == 'y' || input == "yes" {
				database.GenerateDefaultMap(session)
			}
			room, err = database.GetCharacterRoom(session, character)
			printRoom()

		case "loc":
			fallthrough
		case "location":
			printLine(fmt.Sprintf("%v", room.Location))

		case "map":
			width := 20 // Should be even

			startX := room.Location.X - (width / 2)
			startY := room.Location.Y - (width / 2)
			endX := startX + width
			endY := startY + width

			z := room.Location.Z

			for y := startY; y <= endY; y += 1 {
				printString("\n")
				for x := startX; x <= endX; x += 1 {
					currentRoom, err := database.GetRoomByLocation(session, database.Coordinate{x, y, z})
					if err == nil {
						if currentRoom == room {
							printString("*")
						} else {
							printString("#")
						}
					} else {
						printString(" ")
					}
				}
			}
			printString("\n")

		default:
			printLine("Unrecognized command")
		}
	}

	processEvent := func(event string) {
		printLine(fmt.Sprintf("\nAn event happened: %s", event))
		printString(prompt())
	}

	printLine("Welcome, " + utils.FormatName(character.Name))
	printRoom()

	userInputChannel := make(chan string)
	sync := make(chan bool)

	go func() {
		// TODO: Don't crash (due to panic) when client disconnects at the prompt
		for {
			input := utils.GetUserInput(conn, prompt())
			userInputChannel <- input
			<-sync
		}
	}()

	eventChannel := make(chan string)

	go func() {
		for {
			eventChannel <- "event"
			time.Sleep(5 * time.Second)
		}
	}()

	// Main loop
	for {
		select {
		case input := <-userInputChannel:
			if strings.HasPrefix(input, "/") {
				processCommand(input[1:len(input)])
			} else {
				processAction(input)
			}
			sync <- true
		case event := <-eventChannel:
			processEvent(event)
		}
	}
}

// vim: nocindent
