package models

// @protobuf
type Adopter struct {
	// @protobuf(1)
	ID string `json:"id"`
	// @protobuf(2)
	FirstName string `json:"first_name"`
	// @protobuf(3)
	LastName string `json:"last_name"`
	// @protobuf(4)
	Email string `json:"email"`
	// @protobuf(5,optional)
	Phone string `json:"phone,omitempty"`
	// @protobuf(6)
	Address Address `json:"address"`
	// @protobuf(7)
	Verified bool `json:"verified"`
	// @protobuf(8)
	Metadata Metadata `json:"metadata"`
}

// @protobuf
type Address struct {
	// @protobuf(1)
	Street string `json:"street"`
	// @protobuf(2)
	City string `json:"city"`
	// @protobuf(3)
	State string `json:"state"`
	// @protobuf(4)
	Country string `json:"country"`
	// @protobuf(5)
	ZipCode string `json:"zip_code"`
}

// @protobuf
type CreateAdopterRequest struct {
	// @protobuf(1)
	FirstName string `json:"first_name"`
	// @protobuf(2)
	LastName string `json:"last_name"`
	// @protobuf(3)
	Email string `json:"email"`
	// @protobuf(4,optional)
	Phone string `json:"phone,omitempty"`
	// @protobuf(5)
	Address Address `json:"address"`
}

// @protobuf
type UpdateAdopterRequest struct {
	// @protobuf(1)
	ID string `json:"id"`
	// @protobuf(2,optional)
	FirstName string `json:"first_name,omitempty"`
	// @protobuf(3,optional)
	LastName string `json:"last_name,omitempty"`
	// @protobuf(4,optional)
	Email string `json:"email,omitempty"`
	// @protobuf(5,optional)
	Phone string `json:"phone,omitempty"`
	// @protobuf(6,optional)
	Address Address `json:"address,omitempty"`
	// @protobuf(7,optional)
	Verified bool `json:"verified,omitempty"`
}

// @protobuf
type GetAdopterRequest struct {
	// @protobuf(1)
	ID string `json:"id"`
}

// @protobuf
type DeleteAdopterRequest struct {
	// @protobuf(1)
	ID string `json:"id"`
}

// @protobuf
type ListAdoptersRequest struct {
	// @protobuf(1,optional)
	PageInfo PageInfo `json:"page_info,omitempty"`
	// @protobuf(2,optional)
	Verified bool `json:"verified,omitempty"`
	// @protobuf(3,optional)
	Email string `json:"email,omitempty"`
}

// @protobuf
type ListAdoptersResponse struct {
	// @protobuf(1,repeated)
	Adopters []Adopter `json:"adopters"`
	// @protobuf(2)
	PageInfo PageInfo `json:"page_info"`
}
