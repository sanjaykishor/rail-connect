package service

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type Section struct {
	Name string
	MaxSeats int
	Seats map[int]*Seat
	hasVacantSeat bool
	occupiedSeats int
}

type Seat struct {
	Number int
	Available bool
}

type SeatManager struct {
	Sections map[string]*Section
	mu sync.RWMutex
	Logger *zap.Logger
}

type SectionConfig struct {
	SectionName string 
	MaxSeats int
}

func NewSeatManager(sections []SectionConfig, logger *zap.Logger) *SeatManager {
	seatManager := &SeatManager{
		Sections: make(map[string]*Section),
		Logger: logger,
	}

	for _, sectionConfig := range sections {
		section := &Section{
			Name: sectionConfig.SectionName,
			MaxSeats: sectionConfig.MaxSeats,
			Seats: make(map[int]*Seat),
			hasVacantSeat: true,
			occupiedSeats: 0,
		}
		for i := 1; i <= sectionConfig.MaxSeats; i++ {
			section.Seats[i] = &Seat{
				Number: i,
				Available: true,
			}
		}
		seatManager.Sections[sectionConfig.SectionName] = section
	}
	seatManager.Logger.Info("SeatManager initialized", zap.Int("sections", len(sections)))
	return seatManager
}

func (sm *SeatManager) AssignSeat() (string, int, error) {
	sm.mu.RLock()
	defer sm.mu.Unlock()

	for _, section := range sm.Sections { 
		for _, seat := range section.Seats {
			if seat.Available {
				seat.Available = false
				section.occupiedSeats++
				section.hasVacantSeat = section.occupiedSeats < section.MaxSeats
				sm.Logger.Info("Seat assigned", zap.String("section", section.Name), zap.Int("seat_number", seat.Number))
				return section.Name, seat.Number, nil
			}
		}
	}
	sm.Logger.Warn("No available seats")
	return "", -1, fmt.Errorf("no available seats")
}

func (sm *SeatManager) ReleaseSeat(sectionName string, seatNumber int) error {
	sm.mu.RLock()
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

	seat.Available = true
	section.occupiedSeats--
	if section.occupiedSeats == 0 {
		section.hasVacantSeat = true
	}
	sm.Logger.Info("Seat released", zap.String("section", section.Name), zap.Int("seat_number", seat.Number))
	return nil
}

func (sm *SeatManager) UpdateSeat(currSeat int, currSection string, reqSeat int, reqSection string) error {
	sm.mu.RLock()
	defer sm.mu.Unlock()

	oldSection, exists := sm.Sections[currSection]
	if !exists {
		return fmt.Errorf("section %s does not exist", currSection)
	}

	oldSeat, exists := oldSection.Seats[currSeat]
	if !exists {
		return fmt.Errorf("currSeat %d does not exist in section %s", currSeat, currSection)
	}

	if oldSeat.Available {
		return fmt.Errorf("seat %d is already available in section %s", currSeat, currSection)
	}

	newSection, exists := sm.Sections[reqSection]
	if !exists {
		return fmt.Errorf("section %s does not exist", reqSection)
	}

	newSeat, exists := newSection.Seats[reqSeat]
	if !exists {
		return fmt.Errorf("seat %d does not exist in section %s", reqSeat, reqSection)
	}

	if newSeat.Available {
		newSeat.Available = false
		oldSeat.Available = true
		oldSection.occupiedSeats--
		newSection.occupiedSeats++
		if oldSection.occupiedSeats == 0 {
			oldSection.hasVacantSeat = true
		}
		if newSection.occupiedSeats == newSection.MaxSeats {
			newSection.hasVacantSeat = false
		}
		sm.Logger.Info("Seat updated", zap.String("old_section", oldSection.Name), zap.Int("old_seat_number", currSeat), zap.String("new_section", newSection.Name), zap.Int("new_seat_number", reqSeat))
		return nil
	} else {
		return fmt.Errorf("new seat %d in section %s is not available", reqSeat, reqSection)
	}
}



