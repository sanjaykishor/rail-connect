package service

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
	"github.com/sanjaykishor/rail-connect/internal/config"
)

// SeatManager handles the assignment, release, and modification of seats.
// It uses a round-robin strategy to assign seats across multiple sections.
type Section struct {
	Name         string
	MaxSeats     int
	Seats        map[int]*Seat
	VacantSeats  int  // Track number of vacant seats
	FirstVacant  int  // Track first vacant seat for faster lookup
}

// Seat represents an individual seat within a section
type Seat struct {
	Number    int
	Available bool
}

// SeatManager manages seat assignments across multiple sections
type SeatManager struct {
	Sections       map[string]*Section
	SectionOrder   []string           // Maintains section order for round robin
	nextSectionIdx int                // Next section index for round-robin assignments
	mu             sync.Mutex        
	Logger         *zap.Logger
}

// NewSeatManager creates a new SeatManager with the specified sections
func NewSeatManager(sections []config.SectionConfig, logger *zap.Logger) *SeatManager {
	seatManager := &SeatManager{
		Sections:       make(map[string]*Section),
		SectionOrder:   make([]string, len(sections)),
		nextSectionIdx: 0,
		Logger:         logger,
	}

	for i, sectionConfig := range sections {
		section := &Section{
			Name:        sectionConfig.Name,
			MaxSeats:    sectionConfig.MaxSeats,
			Seats:       make(map[int]*Seat),
			VacantSeats: sectionConfig.MaxSeats,
			FirstVacant: 1, // Initially, the first seat is vacant
		}

		for j := 1; j <= sectionConfig.MaxSeats; j++ {
			section.Seats[j] = &Seat{
				Number:    j,
				Available: true,
			}
		}

		seatManager.Sections[sectionConfig.Name] = section
		seatManager.SectionOrder[i] = sectionConfig.Name
	}

	seatManager.Logger.Info("SeatManager initialized", 
		zap.Int("sections", len(sections)),
		zap.Strings("sectionNames", seatManager.SectionOrder))
	
	return seatManager
}

// AssignSeat assigns a seat using round-robin algorithm across sections
func (sm *SeatManager) AssignSeat() (string, int, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	// Try each section once, starting from nextSectionIdx
	totalSections := len(sm.SectionOrder)
	if totalSections == 0 {
		return "", -1, fmt.Errorf("no available sections")
	}
	
	// Try sections in round-robin order
	for i := 0; i < totalSections; i++ {
		currentIdx := (sm.nextSectionIdx + i) % totalSections
		sectionName := sm.SectionOrder[currentIdx]
		section := sm.Sections[sectionName]
		
		// Skip if no vacant seats
		if section.VacantSeats <= 0 {
			continue
		}
		
		// Find the first available seat
		seatNum := section.FirstVacant
		for seatNum <= section.MaxSeats {
			seat, exists := section.Seats[seatNum]
			if exists && seat.Available {
				// Found a seat - assign it
				seat.Available = false
				section.VacantSeats--
				
				// Update first vacant seat pointer
				section.FirstVacant = seatNum + 1
				for section.FirstVacant <= section.MaxSeats {
					if s, ex := section.Seats[section.FirstVacant]; ex && s.Available {
						break
					}
					section.FirstVacant++
				}
				
				// Update next section for round-robin
				sm.nextSectionIdx = (currentIdx + 1) % totalSections
				
				sm.Logger.Info("Seat assigned via round-robin",
					zap.String("section", section.Name),
					zap.Int("seat_number", seat.Number),
					zap.Int("remaining_vacant", section.VacantSeats))
					
				return section.Name, seat.Number, nil
			}
			seatNum++
		}
		
		// there was an inconsistency - fix the count
		section.VacantSeats = 0
	}
	
	sm.Logger.Warn("No available seats in any section")
	return "", -1, fmt.Errorf("no available seats")
}

// ReleaseSeat releases a previously assigned seat
func (sm *SeatManager) ReleaseSeat(sectionName string, seatNumber int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	section, exists := sm.Sections[sectionName]
	if !exists {
		return fmt.Errorf("section %s does not exist", sectionName)
	}
	
	seat, exists := section.Seats[seatNumber]
	if !exists {
		return fmt.Errorf("seat %d does not exist in section %s", seatNumber, sectionName)
	}
	
	if seat.Available {
		return fmt.Errorf("seat %d is already available in section %s", seatNumber, sectionName)
	}
	
	// Update seat status
	seat.Available = true
	section.VacantSeats++
	
	// Update first vacant pointer if this is now earlier than current pointer
	if seatNumber < section.FirstVacant {
		section.FirstVacant = seatNumber
	}
	
	sm.Logger.Info("Seat released",
		zap.String("section", section.Name),
		zap.Int("seat_number", seatNumber),
		zap.Int("vacant_seats", section.VacantSeats))
		
	return nil
}

// UpdateSeat changes a user's seat from one to another
func (sm *SeatManager) UpdateSeat(currSeat int, currSection string, reqSeat int, reqSection string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	oldSectionObj, oldExists := sm.Sections[currSection]
	if !oldExists {
		return fmt.Errorf("section %s does not exist", currSection)
	}
	
	newSectionObj, newExists := sm.Sections[reqSection]
	if !newExists {
		return fmt.Errorf("section %s does not exist", reqSection)
	}
	
	oldSeat, oldSeatExists := oldSectionObj.Seats[currSeat]
	if !oldSeatExists {
		return fmt.Errorf("seat %d does not exist in section %s", currSeat, currSection)
	}
	
	if oldSeat.Available {
		return fmt.Errorf("current seat %d in section %s is not occupied", currSeat, currSection)
	}
	
	newSeat, newSeatExists := newSectionObj.Seats[reqSeat]
	if !newSeatExists {
		return fmt.Errorf("requested seat %d does not exist in section %s", reqSeat, reqSection)
	}
	
	if !newSeat.Available {
		return fmt.Errorf("requested seat %d in section %s is not available", reqSeat, reqSection)
	}
	
	// Update seats
	oldSeat.Available = true
	newSeat.Available = false
	
	// Update vacancy counts
	oldSectionObj.VacantSeats++
	newSectionObj.VacantSeats--
	
	// Update FirstVacant pointers if needed
	if currSeat < oldSectionObj.FirstVacant {
		oldSectionObj.FirstVacant = currSeat
	}
	if reqSeat == newSectionObj.FirstVacant {
		// Find next vacant seat
		newSectionObj.FirstVacant = reqSeat + 1
		for newSectionObj.FirstVacant <= newSectionObj.MaxSeats {
			if s, ex := newSectionObj.Seats[newSectionObj.FirstVacant]; ex && s.Available {
				break
			}
			newSectionObj.FirstVacant++
		}
	}
	
	sm.Logger.Info("Seat updated",
		zap.String("old_section", oldSectionObj.Name),
		zap.Int("old_seat", currSeat),
		zap.String("new_section", newSectionObj.Name),
		zap.Int("new_seat", reqSeat))
		
	return nil
}
