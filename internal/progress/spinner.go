package progress

import (
	"fmt"
	"time"
)

// Spinner shows a simple text-based progress indicator
type Spinner struct {
	message string
	active  bool
	done    chan bool
}

// NewSpinner creates a new spinner with a message
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		done:    make(chan bool),
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.active = true
	go func() {
		chars := []string{"|", "/", "-", "\\"}
		i := 0
		for {
			select {
			case <-s.done:
				return
			default:
				fmt.Printf("\r%s %s", chars[i%len(chars)], s.message)
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

// Stop stops the spinner and clears the line
func (s *Spinner) Stop() {
	if s.active {
		s.done <- true
		s.active = false
		fmt.Print("\r\033[K") // Clear line
	}
}

// Success stops the spinner and shows a success message
func (s *Spinner) Success(message string) {
	s.Stop()
	fmt.Printf("✓ %s\n", message)
}

// Fail stops the spinner and shows an error message
func (s *Spinner) Fail(message string) {
	s.Stop()
	fmt.Printf("✗ %s\n", message)
}

// Update changes the spinner message
func (s *Spinner) Update(message string) {
	s.message = message
}
