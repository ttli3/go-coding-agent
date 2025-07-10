package ui

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

type LoadingIndicator struct {
	message   string
	startTime time.Time
	stopChan  chan bool
	wg        sync.WaitGroup
	isRunning bool
	mu        sync.Mutex
}

var (
	thinkingPhrases = []string{
		"Thinking", "Pondering", "Analyzing", "Processing", "Considering",
		"Examining", "Evaluating", "Reflecting", "Contemplating", "Studying",
	}

	codingPhrases = []string{
		"Cooking", "Crafting", "Building", "Implementing", "Creating",
		"Developing", "Constructing", "Assembling", "Engineering", "Coding",
	}

	readingPhrases = []string{
		"Reading", "Scanning", "Parsing", "Reviewing", "Inspecting",
		"Exploring", "Investigating", "Browsing", "Examining", "Studying",
	}

	searchingPhrases = []string{
		"Searching", "Looking", "Hunting", "Seeking", "Finding",
		"Locating", "Discovering", "Exploring", "Investigating", "Scouting",
	}
)

func NewLoadingIndicator(activityType string) *LoadingIndicator {
	var phrases []string

	switch strings.ToLower(activityType) {
	case "coding", "implementing", "creating", "building":
		phrases = codingPhrases
	case "reading", "viewing", "examining":
		phrases = readingPhrases
	case "searching", "finding", "locating":
		phrases = searchingPhrases
	default:
		phrases = thinkingPhrases
	}

	// Pick a random phrase
	phrase := phrases[rand.Intn(len(phrases))]

	return &LoadingIndicator{
		message:   phrase,
		startTime: time.Now(),
		stopChan:  make(chan bool),
	}
}

func (l *LoadingIndicator) Start() {
	l.mu.Lock()
	if l.isRunning {
		l.mu.Unlock()
		return
	}
	l.isRunning = true
	l.mu.Unlock()

	l.wg.Add(1)
	go l.animate()
}

func (l *LoadingIndicator) Stop() {
	l.mu.Lock()
	if !l.isRunning {
		l.mu.Unlock()
		return
	}
	l.isRunning = false
	l.mu.Unlock()

	close(l.stopChan)
	l.wg.Wait()

	// Clear the loading line
	fmt.Print("\r" + strings.Repeat(" ", 50) + "\r")
}

func (l *LoadingIndicator) animate() {
	defer l.wg.Done()

	// Clean, minimal spinner characters
	spinner := []string{"|", "/", "-", "\\"}
	spinnerIndex := 0

	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-l.stopChan:
			return
		case <-ticker.C:
			elapsed := time.Since(l.startTime)
			elapsedStr := formatDuration(elapsed)

			// Create elegant loading message with proper spacing
			loadingMsg := fmt.Sprintf("  %s  %s  %s",
				color.New(color.FgCyan).Sprint(spinner[spinnerIndex]),
				color.New(color.FgWhite, color.Bold).Sprint(l.message),
				color.New(color.FgHiBlack).Sprint(elapsedStr))

			// Print with carriage return to overwrite
			fmt.Print("\r" + loadingMsg)

			spinnerIndex = (spinnerIndex + 1) % len(spinner)
		}
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
}
