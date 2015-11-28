package database

import "github.com/Cristofori/kmud/types"

type Exit struct {
	Locked bool
}

type Room struct {
	DbObject  `bson:",inline"`
	Container `bson:",inline"`

	ZoneId      types.Id
	AreaId      types.Id `bson:",omitempty"`
	Title       string
	Description string
	Links       map[string]types.Id
	Location    types.Coordinate

	Exits map[types.Direction]*Exit
}

func NewRoom(zoneId types.Id, location types.Coordinate) *Room {
	room := &Room{
		Title: "The Void",
		Description: "You are floating in the blackness of space. Complete darkness surrounds " +
			"you in all directions. There is no escape, there is no hope, just the emptiness. " +
			"You are likely to be eaten by a grue.",
		Location: location,
		ZoneId:   zoneId,
	}

	room.init(room)
	return room
}

func (self *Room) HasExit(dir types.Direction) bool {
	self.ReadLock()
	defer self.ReadUnlock()

	_, found := self.Exits[dir]
	return found
}

func (self *Room) SetExitEnabled(dir types.Direction, enabled bool) {
	self.WriteLock()
	defer self.WriteUnlock()

	if self.Exits == nil {
		self.Exits = map[types.Direction]*Exit{}
	}

	if enabled {
		self.Exits[dir] = &Exit{}
	} else {
		delete(self.Exits, dir)
	}

	self.modified()
}

func (self *Room) SetLink(name string, roomId types.Id) {
	self.WriteLock()
	defer self.WriteUnlock()

	if self.Links == nil {
		self.Links = map[string]types.Id{}
	}

	self.Links[name] = roomId

	self.modified()
}

func (self *Room) RemoveLink(name string) {
	self.WriteLock()
	defer self.WriteUnlock()

	delete(self.Links, name)
	self.modified()
}

func (self *Room) GetLinks() map[string]types.Id {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Links
}

func (self *Room) LinkNames() []string {
	names := make([]string, len(self.GetLinks()))

	i := 0
	for name := range self.Links {
		names[i] = name
		i++
	}
	return names
}

func (self *Room) SetTitle(title string) {
	self.WriteLock()
	defer self.WriteUnlock()

	if title != self.Title {
		self.Title = title
		self.modified()
	}
}

func (self *Room) GetTitle() string {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Title
}

func (self *Room) SetDescription(description string) {
	self.WriteLock()
	defer self.WriteUnlock()

	if self.Description != description {
		self.Description = description
		self.modified()
	}
}

func (self *Room) GetDescription() string {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Description
}

func (self *Room) SetLocation(location types.Coordinate) {
	self.WriteLock()
	defer self.WriteUnlock()

	if location != self.Location {
		self.Location = location
		self.modified()
	}
}

func (self *Room) GetLocation() types.Coordinate {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Location
}

func (self *Room) SetZoneId(zoneId types.Id) {
	self.WriteLock()
	defer self.WriteUnlock()

	if zoneId != self.ZoneId {
		self.ZoneId = zoneId
		self.modified()
	}
}

func (self *Room) GetZoneId() types.Id {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.ZoneId
}

func (self *Room) SetAreaId(areaId types.Id) {
	self.WriteLock()
	defer self.WriteUnlock()

	if areaId != self.AreaId {
		self.AreaId = areaId
		self.modified()
	}
}

func (self *Room) GetAreaId() types.Id {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.AreaId
}

func (self *Room) NextLocation(direction types.Direction) types.Coordinate {
	loc := self.GetLocation()
	return loc.Next(direction)
}

func (self *Room) GetExits() []types.Direction {
	self.ReadLock()
	defer self.ReadUnlock()

	exits := make([]types.Direction, len(self.Exits))

	i := 0
	for dir := range self.Exits {
		exits[i] = dir
		i++
	}

	return exits
}

func (self *Room) SetLocked(dir types.Direction, locked bool) {
	if self.HasExit(dir) {
		self.WriteLock()
		defer self.WriteUnlock()

		self.Exits[dir].Locked = locked
		self.modified()
	}
}

func (self *Room) IsLocked(dir types.Direction) bool {
	self.ReadLock()
	defer self.ReadUnlock()

	if self.HasExit(dir) {
		return self.Exits[dir].Locked
	}

	return false
}

func (self *Room) AddItem(id types.Id) {
	self.Container.AddItem(id)
	self.modified()
}

func (self *Room) RemoveItem(id types.Id) bool {
	self.modified()
	return self.Container.RemoveItem(id)
}

func (self *Room) SetCash(cash int) {
	self.Container.SetCash(cash)
	self.modified()
}

func (self *Room) AddCash(amount int) {
	self.SetCash(self.GetCash() + amount)
}

func (self *Room) RemoveCash(amount int) {
	self.SetCash(self.GetCash() - amount)
}
