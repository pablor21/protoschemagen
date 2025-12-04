package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/example/proto/starwars/models"
)

// InMemoryStarWarsService implements the StarWarsService interface with in-memory data
type InMemoryStarWarsService struct {
	humans map[string]models.Human
	droids map[string]models.Droid
	mu     sync.RWMutex
}

// NewInMemoryStarWarsService creates a new service with sample data
func NewInMemoryStarWarsService() *InMemoryStarWarsService {
	service := &InMemoryStarWarsService{
		humans: make(map[string]models.Human),
		droids: make(map[string]models.Droid),
	}

	service.loadSampleData()
	return service
}

// loadSampleData populates the service with Star Wars characters
func (s *InMemoryStarWarsService) loadSampleData() {
	// Sample Humans
	tatooine := "Tatooine"
	alderaan := "Alderaan"

	luke := models.Human{
		ID:         "luke-skywalker",
		Name:       "Luke Skywalker",
		HomePlanet: &tatooine,
		Height:     1.72,
		Mass:       &[]float64{77.0}[0],
		Starships:  []string{"X-wing", "Imperial shuttle"},
		AppearsIn:  []models.Episode{models.NEWHOPE, models.EMPIRE, models.JEDI},
		Friends:    make(map[string]models.Human),
		CreatedAt:  time.Now().AddDate(0, -1, 0), // 1 month ago
		UpdatedAt:  time.Now(),
	}

	leia := models.Human{
		ID:         "princess-leia",
		Name:       "Princess Leia Organa",
		HomePlanet: &alderaan,
		Height:     1.50,
		Mass:       &[]float64{49.0}[0],
		Starships:  []string{"Rebel transport"},
		AppearsIn:  []models.Episode{models.NEWHOPE, models.EMPIRE, models.JEDI},
		Friends:    make(map[string]models.Human),
		CreatedAt:  time.Now().AddDate(0, -1, 0),
		UpdatedAt:  time.Now(),
	}

	han := models.Human{
		ID:         "han-solo",
		Name:       "Han Solo",
		HomePlanet: nil, // Unknown
		Height:     1.80,
		Mass:       &[]float64{80.0}[0],
		Starships:  []string{"Millennium Falcon", "Imperial shuttle"},
		AppearsIn:  []models.Episode{models.NEWHOPE, models.EMPIRE, models.JEDI},
		Friends:    make(map[string]models.Human),
		CreatedAt:  time.Now().AddDate(0, -1, 0),
		UpdatedAt:  time.Now(),
	}

	// Set up friendships (simplified - using IDs only to avoid circular references in maps)
	luke.Friends = make(map[string]models.Human)
	leia.Friends = make(map[string]models.Human)
	han.Friends = make(map[string]models.Human)

	// Instead of storing full objects, store minimal references to prevent recursion
	lukeFriend := models.Human{ID: leia.ID, Name: leia.Name}
	leiaFriend := models.Human{ID: luke.ID, Name: luke.Name}
	hanFriend := models.Human{ID: luke.ID, Name: luke.Name}

	luke.Friends[leia.ID] = lukeFriend
	luke.Friends[han.ID] = hanFriend
	leia.Friends[luke.ID] = leiaFriend
	leia.Friends[han.ID] = hanFriend
	han.Friends[luke.ID] = hanFriend
	han.Friends[leia.ID] = lukeFriend

	s.humans[luke.ID] = luke
	s.humans[leia.ID] = leia
	s.humans[han.ID] = han

	// Sample Droids
	c3po := models.Droid{
		ID:              "c-3po",
		Name:            "C-3PO",
		PrimaryFunction: "Protocol",
		FriendsMap:      make(map[string]*models.Human),
		// AppearsIn:       []models.Episode{models.NEWHOPE, models.EMPIRE, models.JEDI},
		// CreatedAt:       time.Now().AddDate(0, -2, 0), // 2 months ago
		// UpdatedAt:       time.Now(),
	}

	r2d2 := models.Droid{
		ID:              "r2-d2",
		Name:            "R2-D2",
		PrimaryFunction: "Astromech",
		FriendsMap:      make(map[string]*models.Human),
		// AppearsIn:       []models.Episode{models.NEWHOPE, models.EMPIRE, models.JEDI},
		// CreatedAt:       time.Now().AddDate(0, -2, 0),
		// UpdatedAt:       time.Now(),
	}

	// Set up droid friendships with humans
	c3po.FriendsMap[luke.ID] = &luke
	c3po.FriendsMap[leia.ID] = &leia
	r2d2.FriendsMap[luke.ID] = &luke
	r2d2.FriendsMap[leia.ID] = &leia

	s.droids[c3po.ID] = c3po
	s.droids[r2d2.ID] = r2d2
}

// GetHuman retrieves a human by ID
func (s *InMemoryStarWarsService) GetHuman(ctx context.Context, req models.GetHumanRequest) (models.Human, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	human, exists := s.humans[req.ID]
	if !exists {
		return models.Human{}, fmt.Errorf("human with ID %s not found", req.ID)
	}

	return human, nil
}

// GetDroid retrieves a droid by ID
func (s *InMemoryStarWarsService) GetDroid(ctx context.Context, req models.GetDroidRequest) (models.Droid, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	droid, exists := s.droids[req.ID]
	if !exists {
		return models.Droid{}, fmt.Errorf("droid with ID %s not found", req.ID)
	}

	return droid, nil
}

// StreamHumans returns a slice of humans with pagination
func (s *InMemoryStarWarsService) StreamHumans(ctx context.Context, req models.StreamHumansRequest) ([]models.Human, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Convert map to slice for pagination
	var allHumans []models.Human
	for _, human := range s.humans {
		allHumans = append(allHumans, human)
	}

	// Apply pagination
	start := int(req.Offset)
	end := start + int(req.Limit)

	if start >= len(allHumans) {
		return []models.Human{}, nil
	}

	if end > len(allHumans) {
		end = len(allHumans)
	}

	// Return slice directly
	return allHumans[start:end], nil
} // UploadHumans adds multiple humans from a slice to the database
func (s *InMemoryStarWarsService) UploadHumans(ctx context.Context, humans []models.Human) (models.UploadHumansResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for _, human := range humans {
		if human.ID == "" {
			continue // Skip humans without IDs
		}

		// Set timestamps
		if human.CreatedAt.IsZero() {
			human.CreatedAt = time.Now()
		}
		human.UpdatedAt = time.Now()

		s.humans[human.ID] = human
		count++
	}

	return models.UploadHumansResponse{
		Count:  int32(count),
		Status: fmt.Sprintf("Successfully uploaded %d humans", count),
	}, nil
} // ChatWithCharacters simulates a chat system using slices
func (s *InMemoryStarWarsService) ChatWithCharacters(ctx context.Context, messages []models.ChatMessage) ([]models.ChatMessage, error) {
	var responses []models.ChatMessage

	for _, msg := range messages {
		// Generate automated responses based on the recipient
		var response models.ChatMessage

		switch msg.To {
		case "luke-skywalker":
			response = models.ChatMessage{
				ID:        fmt.Sprintf("resp-%s", msg.ID),
				From:      "luke-skywalker",
				To:        msg.From,
				Message:   fmt.Sprintf("Luke here! Thanks for your message: '%s'. May the Force be with you!", msg.Message),
				Timestamp: time.Now().Unix(),
			}
		case "princess-leia":
			response = models.ChatMessage{
				ID:        fmt.Sprintf("resp-%s", msg.ID),
				From:      "princess-leia",
				To:        msg.From,
				Message:   fmt.Sprintf("Princess Leia responding: '%s'. Help us, you're our only hope!", msg.Message),
				Timestamp: time.Now().Unix(),
			}
		case "c-3po":
			response = models.ChatMessage{
				ID:        fmt.Sprintf("resp-%s", msg.ID),
				From:      "c-3po",
				To:        msg.From,
				Message:   fmt.Sprintf("C-3PO here! The odds of understanding your message '%s' are approximately 3,720 to 1!", msg.Message),
				Timestamp: time.Now().Unix(),
			}
		default:
			response = models.ChatMessage{
				ID:        fmt.Sprintf("resp-%s", msg.ID),
				From:      "system",
				To:        msg.From,
				Message:   fmt.Sprintf("Character '%s' is not available. Your message: '%s'", msg.To, msg.Message),
				Timestamp: time.Now().Unix(),
			}
		}

		responses = append(responses, response)
	}

	return responses, nil
} // Additional helper methods for the service

// ListAllHumans returns all humans in the system
func (s *InMemoryStarWarsService) ListAllHumans() []models.Human {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var humans []models.Human
	for _, human := range s.humans {
		humans = append(humans, human)
	}
	return humans
}

// ListAllDroids returns all droids in the system
func (s *InMemoryStarWarsService) ListAllDroids() []models.Droid {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var droids []models.Droid
	for _, droid := range s.droids {
		droids = append(droids, droid)
	}
	return droids
}

// GetHumanByName finds a human by name (case-insensitive)
func (s *InMemoryStarWarsService) GetHumanByName(name string) (models.Human, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, human := range s.humans {
		if human.Name == name {
			return human, nil
		}
	}

	return models.Human{}, fmt.Errorf("human with name %s not found", name)
}

// GetStats returns service statistics
func (s *InMemoryStarWarsService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"total_humans":     len(s.humans),
		"total_droids":     len(s.droids),
		"total_characters": len(s.humans) + len(s.droids),
		"last_updated":     time.Now(),
	}
}

// Helper methods for slice-based operations (for testing and demos)

// StreamHumansSlice returns a slice of humans (helper method for demos)
func (s *InMemoryStarWarsService) StreamHumansSlice(req models.StreamHumansRequest) ([]models.Human, error) {
	return s.StreamHumans(context.Background(), req)
}

// UploadHumansSlice uploads humans from a slice (helper method for demos)
func (s *InMemoryStarWarsService) UploadHumansSlice(humans []models.Human) (models.UploadHumansResponse, error) {
	return s.UploadHumans(context.Background(), humans)
}

// ChatWithCharactersSlice handles slice-based chat (helper method for demos)
func (s *InMemoryStarWarsService) ChatWithCharactersSlice(messages []models.ChatMessage) ([]models.ChatMessage, error) {
	return s.ChatWithCharacters(context.Background(), messages)
}
