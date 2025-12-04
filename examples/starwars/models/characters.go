// @namespace("characters")
package models

import "time"

// Human character message
// @reserved
// @message name:"Human" description:"Human character" reserved
// @message name:"HumanRequest" description:"Request message for Human" reserved
// @message name:"HumanDetailed" description:"Detailed Human with timestamps"
type Human struct {
	// @field number:1 description:"The unique human identifier" omit:"HumanRequest"
	ID string
	// @field number:2
	Name string
	// @field number:3 description:"List of friend IDs"
	Friends map[string]Human
	// @field number:4 include:["Human","HumanDetailed"]
	HomePlanet *string
	// @field number:5
	Height float64
	// @field number:6
	Mass *float64
	// @field number:7
	Starships []string
	// @field number:8 description:"Episodes this human appears in"
	AppearsIn []Episode
	// @field number:9 name:"created_at" include:"HumanDetailed"
	CreatedAt time.Time
	// @field number:10 name:"updated_at" omit:["Human","HumanRequest"]
	UpdatedAt time.Time
}

// Droid character message
// @message name:"Droid"
type Droid struct {
	// @field number:1
	ID string
	// @field number:2
	Name string
	// @field number:3 repeated:true
	Friends []string
	// @field number:4
	PrimaryFunction string
	// @field number:"5" description:"Metadata key-value pairs"
	Metadata map[string]string // Auto-detected as map<string, string>
	// @map key:"string" value:"int32" number:"6" description:"Task assignments by ID"
	Tasks map[string]int32
	// @map key:"string" value:"Human" number:"7" description:"Friends by ID - map with complex message type as value"
	FriendsMap map[string]*Human
}

// Episode enum
// @enum name:"Episode"
type Episode int

const (
	NEWHOPE Episode = iota
	EMPIRE
	JEDI
)

// @message name:"HumanResponse"
type HumanResponse = Response[Human]

// Generic response wrapper
type Response[T any] struct {
	Data  T
	Error string
}
