package models

import "time"

// @protobuf
type Metadata struct {
	// @protobuf(1)
	CreatedAt time.Time `json:"created_at"`
	// @protobuf(2)
	UpdatedAt time.Time `json:"updated_at"`
	// @protobuf(3,optional)
	CreatedBy string `json:"created_by,omitempty"`
}

// @protobuf
type PageInfo struct {
	// @protobuf(1)
	Page int32 `json:"page"`
	// @protobuf(2)
	PageSize int32 `json:"page_size"`
	// @protobuf(3)
	Total int64 `json:"total"`
}

// @protobuf
type PetStatus int32

const (
	// @protobuf(0,name="PET_STATUS_AVAILABLE")
	PetStatusAvailable PetStatus = iota
	// @protobuf(1,name="PET_STATUS_PENDING")
	PetStatusPending
	// @protobuf(2,name="PET_STATUS_ADOPTED")
	PetStatusAdopted
)

// @protobuf
type PetCategory int32

const (
	// @protobuf(0,name="PET_CATEGORY_DOG")
	PetCategoryDog PetCategory = iota
	// @protobuf(1,name="PET_CATEGORY_CAT")
	PetCategoryCat
	// @protobuf(2,name="PET_CATEGORY_BIRD")
	PetCategoryBird
	// @protobuf(3,name="PET_CATEGORY_FISH")
	PetCategoryFish
	// @protobuf(4,name="PET_CATEGORY_OTHER")
	PetCategoryOther
)
