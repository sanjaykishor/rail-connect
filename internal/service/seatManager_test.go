package service

import (
	"github.com/sanjaykishor/rail-connect/internal/config"
	"github.com/stretchr/testify/assert"
	"testing"

	"go.uber.org/zap"
)

func CreateSeatManager() *SeatManager {
	sectionConfigs := []config.SectionConfig{
		{Name: "A", MaxSeats: 20},
		{Name: "B", MaxSeats: 20},
	}

	logger:= zap.NewNop()
	return NewSeatManager(sectionConfigs, logger)
}

func TestNewSeatManager(t *testing.T) {

	seatManager := CreateSeatManager()

	assert.NotNil(t, seatManager, "SeatManager should be initialized")
	assert.Contains(t, seatManager.Sections, "A", "Section A should be present")
	assert.Contains(t, seatManager.Sections, "B", "Section B should be present")
	assert.Equal(t, seatManager.Sections["A"].MaxSeats, 20, "Section A should have 20 seats")
	assert.Equal(t, seatManager.Sections["B"].MaxSeats, 20, "Section B should have 20 seats")
	assert.Equal(t, seatManager.Sections["A"].VacantSeats, 20, "Section A should have 20 vacant seats")
	assert.Equal(t, seatManager.Sections["B"].VacantSeats, 20, "Section B should have 20 vacant seats")
	assert.Equal(t, seatManager.Sections["A"].FirstVacant, 1, "Section A should have first vacant seat as 1")
	assert.Equal(t, seatManager.Sections["B"].FirstVacant, 1, "Section B should have first vacant seat as 1")
	assert.Equal(t, seatManager.SectionOrder[0], "A", "First section in order should be A")
	assert.Equal(t, seatManager.SectionOrder[1], "B", "Second section in order should be B")
	assert.Equal(t, seatManager.nextSectionIdx, 0, "Next section index should be 0")
	assert.Equal(t, seatManager.Sections["A"].Seats[1].Available, true, "First seat in section A should be available")
	assert.Equal(t, seatManager.Sections["B"].Seats[1].Available, true, "First seat in section B should be available")
}

func TestAssignSeat(t *testing.T) {

	seatManager := CreateSeatManager()

	// Assign a seat
	sectionName, seatNumber, err := seatManager.AssignSeat()
	assert.NoError(t, err, "Should not return an error when assigning a seat")
	assert.Equal(t, sectionName, "A", "First section in order should be A")
	assert.Equal(t, seatNumber, 1, "First seat in section A should be assigned")
	assert.Equal(t, seatManager.Sections["A"].VacantSeats, 19, "Section A should have 19 vacant seats after assignment")
	assert.Equal(t, seatManager.Sections["A"].FirstVacant, 2, "Section A should have first vacant seat as 2 after assignment")
	assert.Equal(t, seatManager.Sections["A"].Seats[1].Available, false, "First seat in section A should not be available after assignment")

	// Assign another seat
	sectionName, seatNumber, err = seatManager.AssignSeat()
	assert.NoError(t, err, "Should not return an error when assigning a seat")
	assert.Equal(t, sectionName, "B", "First section in order should be B")
	assert.Equal(t, seatNumber, 1, "Second seat in section A should be assigned")
	assert.Equal(t, seatManager.Sections["B"].VacantSeats, 19, "Section B should have 19 vacant seats after assignment")
	assert.Equal(t, seatManager.Sections["B"].FirstVacant, 2, "Section B should have first vacant seat as 2 after assignment")
	assert.Equal(t, seatManager.Sections["B"].Seats[1].Available, false, "Second seat in section B should not be available after assignment")

	// Assign another seat
	sectionName, seatNumber, err = seatManager.AssignSeat()
	assert.NoError(t, err, "Should not return an error when assigning a seat")
	assert.Equal(t, sectionName, "A", "First section in order should be A")
	assert.Equal(t, seatNumber, 2, "Second seat in section A should be assigned")
	assert.Equal(t, seatManager.Sections["A"].VacantSeats, 18, "Section A should have 18 vacant seats after assignment")
	assert.Equal(t, seatManager.Sections["A"].FirstVacant, 3, "Section A should have first vacant seat as 3 after assignment")
	assert.Equal(t, seatManager.Sections["A"].Seats[2].Available, false, "Second seat in section A should not be available after assignment")

	// No vacant seats
	// Fill up all seats in section A
	for i := 3; i <= 20; i++ {
		seatManager.Sections["A"].Seats[i].Available = false
		seatManager.Sections["A"].VacantSeats--
	}
	seatManager.Sections["A"].FirstVacant = 21
	// Assign a seat
	sectionName, seatNumber, err = seatManager.AssignSeat()
	assert.NoError(t, err, "Should not return an error when assigning a seat")
	assert.Equal(t, sectionName, "B", "First section in order should be B")
	assert.Equal(t, seatNumber, 2, "Second seat in section B should be assigned")
	assert.Equal(t, seatManager.Sections["B"].VacantSeats, 18, "Section B should have 18 vacant seats after assignment")
	assert.Equal(t, seatManager.Sections["B"].FirstVacant, 3, "Section B should have first vacant seat as 3 after assignment")
	assert.Equal(t, seatManager.Sections["B"].Seats[2].Available, false, "Second seat in section B should not be available after assignment")

	// Fill up all seats in section B
	for i := 3; i <= 20; i++ {
		seatManager.Sections["B"].Seats[i].Available = false
		seatManager.Sections["B"].VacantSeats--
	}
	seatManager.Sections["B"].FirstVacant = 21
	// Assign a seat
	sectionName, seatNumber, err = seatManager.AssignSeat()
	assert.Error(t, err, "Should return an error when no seats are available")
	assert.Equal(t, sectionName, "", "Section name should be empty when no seats are available")
	assert.Equal(t, seatNumber, -1, "Seat number should be -1 when no seats are available")
	assert.Equal(t, seatManager.Sections["A"].VacantSeats, 0, "Section A should have 0 vacant seats after assignment")
	assert.Equal(t, seatManager.Sections["B"].VacantSeats, 0, "Section B should have 0 vacant seats after assignment")
	assert.Equal(t, seatManager.Sections["A"].FirstVacant, 21, "Section A should have first vacant seat as 21 after assignment")
	assert.Equal(t, seatManager.Sections["B"].FirstVacant, 21, "Section B should have first vacant seat as 21 after assignment")
	assert.Equal(t, seatManager.Sections["A"].Seats[20].Available, false, "Last seat in section A should not be available after assignment")
	assert.Equal(t, seatManager.Sections["B"].Seats[20].Available, false, "Last seat in section B should not be available after assignment")
}

func TestReleaseSeat(t *testing.T) {
	seatManager := CreateSeatManager()

	tests := []struct {
		sectionName         string
		seatNumber          int
		expectedVacantSeats int
		expectedFirstVacant int
		expectedAvailable   bool
	}{
		{"A", 1, 20, 1, true},
		{"B", 1, 20, 1, true},
	}

	for _, test := range tests {
		// Assign a seat
		seatManager.Sections[test.sectionName].Seats[test.seatNumber].Available = false
		seatManager.Sections[test.sectionName].VacantSeats--
		seatManager.Sections[test.sectionName].FirstVacant++

		// Release the seat
		err := seatManager.ReleaseSeat(test.sectionName, test.seatNumber)
		assert.NoError(t, err, "Should not return an error when releasing a seat")

		// Check the expected values
		assert.Equal(t, test.expectedVacantSeats, seatManager.Sections[test.sectionName].VacantSeats, "Vacant seats should match")
		assert.Equal(t, test.expectedFirstVacant, seatManager.Sections[test.sectionName].FirstVacant, "First vacant seat should match")
		assert.Equal(t, test.expectedAvailable, seatManager.Sections[test.sectionName].Seats[test.seatNumber].Available, "Seat availability should match")
	}

	// Test releasing a seat that is already available
	err := seatManager.ReleaseSeat("A", 1)
	assert.Error(t, err, "Should return an error when releasing an already available seat")

	// Test releasing a seat that does not exist
	err = seatManager.ReleaseSeat("A", 100)
	assert.Error(t, err, "Should return an error when releasing a seat that does not exist")

	// Test releasing a seat in a section that does not exist
	err = seatManager.ReleaseSeat("C", 1)
	assert.Error(t, err, "Should return an error when releasing a seat in a section that does not exist")
}

func TestUpdateSeat(t *testing.T) {
	seatManager := CreateSeatManager()

	tests := []struct {
		sectionName                string
		seatNumber                 int
		newSectionName             string
		newSeatNumber              int
		expectedVacantSeats        int
		expectedFirstVacant        int
		expectedAvailableOfNewSeat bool
	}{
		{"A", 1, "B", 1, 19, 2, false},
		{"B", 1, "A", 1, 19, 2, false},
	}
	for _, test := range tests {
		// Assign a seat
		seatManager.Sections[test.sectionName].Seats[test.seatNumber].Available = false
		seatManager.Sections[test.sectionName].VacantSeats--
		seatManager.Sections[test.sectionName].FirstVacant++

		// Update the seat
		err := seatManager.UpdateSeat(test.seatNumber, test.sectionName, test.newSeatNumber, test.newSectionName)
		assert.NoError(t, err, "Should not return an error when updating a seat")

		// Check the expected values
		assert.Equal(t, test.expectedVacantSeats, seatManager.Sections[test.newSectionName].VacantSeats, "Vacant seats should match")
		assert.Equal(t, test.expectedFirstVacant, seatManager.Sections[test.newSectionName].FirstVacant, "First vacant seat should match")
		assert.Equal(t, test.expectedAvailableOfNewSeat, seatManager.Sections[test.newSectionName].Seats[test.newSeatNumber].Available, "Seat availability should match")
	}

	// Test updating a seat that is already available
	seatManager.Sections["A"].Seats[1].Available = true
	err := seatManager.UpdateSeat(1, "A", 1, "B")
	assert.Error(t, err, "Should return an error when updating an already available seat")

	// Test updating a seat that does not exist
	err = seatManager.UpdateSeat(100, "A", 1, "B")
	assert.Error(t, err, "Should return an error when updating a seat that does not exist")

	// Test updating a seat in a section that does not exist
	err = seatManager.UpdateSeat(1, "C", 1, "B")
	assert.Error(t, err, "Should return an error when updating a seat in a section that does not exist")
}
