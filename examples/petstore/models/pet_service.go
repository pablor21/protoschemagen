package models

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PetService is a concrete struct that provides CRUD operations for pets
// @service name:"PetService"
type PetService struct {
	mu   sync.RWMutex
	pets map[string]Pet
}

// NewPetService creates a new instance of PetService
func NewPetService() *PetService {
	return &PetService{
		pets: make(map[string]Pet),
	}
}

// CreatePet creates a new pet
// @rpc name:"CreatePet" input:"CreatePetRequest" output:"Pet"
func (s *PetService) CreatePet(ctx context.Context, req CreatePetRequest) (Pet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("pet_%d", time.Now().UnixNano())
	now := time.Now()

	pet := Pet{
		ID:          id,
		Name:        req.Name,
		Category:    req.Category,
		Breed:       req.Breed,
		Age:         req.Age,
		Status:      PetStatusAvailable,
		Description: req.Description,
		PhotoUrls:   req.PhotoUrls,
		Metadata: Metadata{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	s.pets[id] = pet
	return pet, nil
}

// GetPet retrieves a pet by ID
// @rpc name:"GetPet" input:"GetPetRequest" output:"Pet"
func (s *PetService) GetPet(ctx context.Context, req GetPetRequest) (Pet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pet, exists := s.pets[req.ID]
	if !exists {
		return Pet{}, fmt.Errorf("pet with ID %s not found", req.ID)
	}

	return pet, nil
}

// UpdatePet updates an existing pet
// @rpc name:"UpdatePet" input:"UpdatePetRequest" output:"Pet"
func (s *PetService) UpdatePet(ctx context.Context, req UpdatePetRequest) (Pet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pet, exists := s.pets[req.ID]
	if !exists {
		return Pet{}, fmt.Errorf("pet with ID %s not found", req.ID)
	}

	// Update only non-empty fields
	if req.Name != "" {
		pet.Name = req.Name
	}
	if req.Breed != "" {
		pet.Breed = req.Breed
	}
	if req.Age > 0 {
		pet.Age = req.Age
	}
	if req.Description != "" {
		pet.Description = req.Description
	}
	if req.PhotoUrls != nil {
		pet.PhotoUrls = req.PhotoUrls
	}
	if req.AdopterID != "" {
		pet.AdopterID = req.AdopterID
	}

	pet.Metadata.UpdatedAt = time.Now()
	s.pets[req.ID] = pet

	return pet, nil
}

// DeletePet removes a pet
// @rpc name:"DeletePet" input:"DeletePetRequest" output:"Pet"
func (s *PetService) DeletePet(ctx context.Context, req DeletePetRequest) (Pet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pet, exists := s.pets[req.ID]
	if !exists {
		return Pet{}, fmt.Errorf("pet with ID %s not found", req.ID)
	}

	delete(s.pets, req.ID)
	return pet, nil
}

// ListPets returns a list of pets with optional filtering and pagination
// @rpc name:"ListPets" input:"ListPetsRequest" output:"ListPetsResponse"
func (s *PetService) ListPets(ctx context.Context, req ListPetsRequest) (ListPetsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filteredPets []Pet
	for _, pet := range s.pets {
		// Apply filters
		if req.Status != 0 && pet.Status != req.Status {
			continue
		}
		if req.Category != 0 && pet.Category != req.Category {
			continue
		}
		if req.AdopterID != "" && pet.AdopterID != req.AdopterID {
			continue
		}
		filteredPets = append(filteredPets, pet)
	}

	// Simple pagination
	pageSize := req.PageInfo.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	page := req.PageInfo.Page
	if page <= 0 {
		page = 1
	}

	start := (page - 1) * pageSize
	end := start + pageSize

	if start > int32(len(filteredPets)) {
		start = int32(len(filteredPets))
	}
	if end > int32(len(filteredPets)) {
		end = int32(len(filteredPets))
	}

	result := filteredPets[start:end]

	return ListPetsResponse{
		Pets: result,
		PageInfo: PageInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    int64(len(filteredPets)),
		},
	}, nil
}
