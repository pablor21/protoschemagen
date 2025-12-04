package models

// @protobuf
type Pet struct {
	// @protobuf(1)
	ID string `json:"id"`
	// @protobuf(2)
	Name string `json:"name"`
	// @protobuf(3)
	Category PetCategory `json:"category"`
	// @protobuf(4)
	Breed string `json:"breed"`
	// @protobuf(5)
	Age int32 `json:"age"`
	// @protobuf(6)
	Status PetStatus `json:"status"`
	// @protobuf(7,optional)
	Description string `json:"description,omitempty"`
	// @protobuf(8,repeated)
	PhotoUrls []string `json:"photo_urls,omitempty"`
	// @protobuf(9,optional)
	AdopterID string `json:"adopter_id,omitempty"`
	// @protobuf(10)
	Metadata Metadata `json:"metadata"`
}

// @protobuf
type CreatePetRequest struct {
	// @protobuf(1)
	Name string `json:"name"`
	// @protobuf(2)
	Category PetCategory `json:"category"`
	// @protobuf(3)
	Breed string `json:"breed"`
	// @protobuf(4)
	Age int32 `json:"age"`
	// @protobuf(5,optional)
	Description string `json:"description,omitempty"`
	// @protobuf(6,repeated)
	PhotoUrls []string `json:"photo_urls,omitempty"`
}

// @protobuf
type UpdatePetRequest struct {
	// @protobuf(1)
	ID string `json:"id"`
	// @protobuf(2,optional)
	Name string `json:"name,omitempty"`
	// @protobuf(3,optional)
	Category PetCategory `json:"category,omitempty"`
	// @protobuf(4,optional)
	Breed string `json:"breed,omitempty"`
	// @protobuf(5,optional)
	Age int32 `json:"age,omitempty"`
	// @protobuf(6,optional)
	Status PetStatus `json:"status,omitempty"`
	// @protobuf(7,optional)
	Description string `json:"description,omitempty"`
	// @protobuf(8,repeated)
	PhotoUrls []string `json:"photo_urls,omitempty"`
	// @protobuf(9,optional)
	AdopterID string `json:"adopter_id,omitempty"`
}

// @protobuf
type GetPetRequest struct {
	// @protobuf(1)
	ID string `json:"id"`
}

// @protobuf
type DeletePetRequest struct {
	// @protobuf(1)
	ID string `json:"id"`
}

// @protobuf
type ListPetsRequest struct {
	// @protobuf(1,optional)
	PageInfo PageInfo `json:"page_info,omitempty"`
	// @protobuf(2,optional)
	Status PetStatus `json:"status,omitempty"`
	// @protobuf(3,optional)
	Category PetCategory `json:"category,omitempty"`
	// @protobuf(4,optional)
	AdopterID string `json:"adopter_id,omitempty"`
}

// @protobuf
type ListPetsResponse struct {
	// @protobuf(1,repeated)
	Pets []Pet `json:"pets"`
	// @protobuf(2)
	PageInfo PageInfo `json:"page_info"`
}
