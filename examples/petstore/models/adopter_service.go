package models

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AdopterService is a concrete struct that provides CRUD operations for adopters
// @service name:"AdopterService"
type AdopterService struct {
	mu       sync.RWMutex
	adopters map[string]Adopter
}

// NewAdopterService creates a new instance of AdopterService
func NewAdopterService() *AdopterService {
	return &AdopterService{
		adopters: make(map[string]Adopter),
	}
}

// CreateAdopter creates a new adopter
// @rpc name:"CreateAdopter" input:"CreateAdopterRequest" output:"Adopter"
func (s *AdopterService) CreateAdopter(ctx context.Context, req CreateAdopterRequest) (Adopter, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("adopter_%d", time.Now().UnixNano())
	now := time.Now()

	adopter := Adopter{
		ID:        id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     req.Phone,
		Address:   req.Address,
		Verified:  false, // New adopters start unverified
		Metadata: Metadata{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	s.adopters[id] = adopter
	return adopter, nil
}

// GetAdopter retrieves an adopter by ID
// @rpc name:"GetAdopter" input:"GetAdopterRequest" output:"Adopter"
func (s *AdopterService) GetAdopter(ctx context.Context, req GetAdopterRequest) (Adopter, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	adopter, exists := s.adopters[req.ID]
	if !exists {
		return Adopter{}, fmt.Errorf("adopter with ID %s not found", req.ID)
	}

	return adopter, nil
}

// UpdateAdopter updates an existing adopter
// @rpc name:"UpdateAdopter" input:"UpdateAdopterRequest" output:"Adopter"
func (s *AdopterService) UpdateAdopter(ctx context.Context, req UpdateAdopterRequest) (Adopter, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	adopter, exists := s.adopters[req.ID]
	if !exists {
		return Adopter{}, fmt.Errorf("adopter with ID %s not found", req.ID)
	}

	// Update only non-empty fields
	if req.FirstName != "" {
		adopter.FirstName = req.FirstName
	}
	if req.LastName != "" {
		adopter.LastName = req.LastName
	}
	if req.Email != "" {
		adopter.Email = req.Email
	}
	if req.Phone != "" {
		adopter.Phone = req.Phone
	}
	// Address struct - check if it has any fields set
	if req.Address.Street != "" || req.Address.City != "" || req.Address.State != "" ||
		req.Address.Country != "" || req.Address.ZipCode != "" {
		adopter.Address = req.Address
	}
	adopter.Verified = req.Verified

	adopter.Metadata.UpdatedAt = time.Now()
	s.adopters[req.ID] = adopter

	return adopter, nil
}

// DeleteAdopter removes an adopter
// @rpc name:"DeleteAdopter" input:"DeleteAdopterRequest" output:"Adopter"
func (s *AdopterService) DeleteAdopter(ctx context.Context, req DeleteAdopterRequest) (Adopter, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	adopter, exists := s.adopters[req.ID]
	if !exists {
		return Adopter{}, fmt.Errorf("adopter with ID %s not found", req.ID)
	}

	delete(s.adopters, req.ID)
	return adopter, nil
}

// ListAdopters returns a list of adopters with optional filtering and pagination
// @rpc name:"ListAdopters" input:"ListAdoptersRequest" output:"ListAdoptersResponse"
func (s *AdopterService) ListAdopters(ctx context.Context, req ListAdoptersRequest) (ListAdoptersResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filteredAdopters []Adopter
	for _, adopter := range s.adopters {
		// Apply filters
		if req.Email != "" && adopter.Email != req.Email {
			continue
		}
		filteredAdopters = append(filteredAdopters, adopter)
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

	if start > int32(len(filteredAdopters)) {
		start = int32(len(filteredAdopters))
	}
	if end > int32(len(filteredAdopters)) {
		end = int32(len(filteredAdopters))
	}

	result := filteredAdopters[start:end]

	return ListAdoptersResponse{
		Adopters: result,
		PageInfo: PageInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    int64(len(filteredAdopters)),
		},
	}, nil
}
