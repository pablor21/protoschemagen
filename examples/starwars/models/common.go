// @namespace("common")
package models

// Metadata holds common metadata fields
// @message name:"Metadata"
type Metadata struct {
	// @field number:1
	Version string
	// @field number:2
	Timestamp int64
	// @field number:3
	Source string
}

// Status represents operation status
// @enum name:"Status"
type Status int

const (
	// @enum_value name:"STATUS_UNKNOWN" number:0
	StatusUnknown Status = iota
	// @enum_value name:"STATUS_SUCCESS" number:1
	StatusSuccess
	// @enum_value name:"STATUS_ERROR" number:2
	StatusError
	// @enum_value name:"STATUS_PENDING" number:3
	StatusPending
)

// PageInfo contains pagination information
// @message name:"PageInfo"
type PageInfo struct {
	// @field number:1
	HasNextPage bool
	// @field number:2
	HasPreviousPage bool
	// @field number:3
	TotalCount int32
	// @field number:4
	CurrentPage int32
}
