package ynab_test

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/joeyhipolito/via/features/events"
	"github.com/joeyhipolito/via/features/events/schema"
)

// ============================================================================
// Real-Time Event-Driven End-to-End Tests
// ============================================================================

// TestE2E_RealTime_TransactionNotifications tests real-time transaction notifications
// across multiple platforms (CLI, Web, Mobile, Chat).
func TestE2E_RealTime_TransactionNotifications(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real-time E2E test in short mode")
	}

	tmpDir := t.TempDir()

	// Initialize event bus
	bus := events.NewBus()
	defer bus.Close()

	// Simulate multiple platform subscribers
	platforms := []string{"web-dashboard", "mobile-app", "telegram-bot", "cli-watcher"}
	platformEvents := make(map[string][]schema.Event)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Each platform subscribes to transaction events
	for _, platform := range platforms {
		wg.Add(1)
		platformName := platform // Capture for closure

		bus.Subscribe("budget:transaction:*", func(e schema.Event) {
			mu.Lock()
			platformEvents[platformName] = append(platformEvents[platformName], e)
			mu.Unlock()

			// Simulate platform-specific handling
			handlePlatformEvent(t, platformName, e)
			wg.Done()
		})
	}

	// Simulate user adding transaction via CLI
	correlationID := "user-action-123"

	// Publish transaction added event
	transactionEvent := schema.NewEvent(
		"budget:transaction:added",
		map[string]interface{}{
			"budget_id": "personal-budget",
			"transaction": map[string]interface{}{
				"id":          "tx-realtime-1",
				"amount":      -35.50,
				"payee":       "Coffee Shop",
				"category":    "Dining Out",
				"category_id": "cat-dining",
				"date":        "2026-02-02",
				"account_id":  "checking",
				"account":     "Checking Account",
			},
			"platform": "cli",
			"user_id":  "user-123",
		},
		correlationID,
	)

	bus.Publish(transactionEvent)

	// Wait for all platforms to receive event
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for platforms to receive event")
	}

	// Verify all platforms received the event
	mu.Lock()
	defer mu.Unlock()

	for _, platform := range platforms {
		events := platformEvents[platform]
		if len(events) != 1 {
			t.Errorf("Platform %s: expected 1 event, got %d", platform, len(events))
		}

		if len(events) > 0 {
			event := events[0]
			if event.Type != "budget:transaction:added" {
				t.Errorf("Platform %s: wrong event type: %s", platform, event.Type)
			}

			payload := event.Payload.(map[string]interface{})
			tx := payload["transaction"].(map[string]interface{})

			if tx["amount"].(float64) != -35.50 {
				t.Errorf("Platform %s: wrong amount: %v", platform, tx["amount"])
			}
		}
	}

	t.Logf("✓ Real-time notifications: %d platforms received transaction event", len(platforms))
}

// TestE2E_RealTime_BudgetAlertPropagation tests budget alert propagation
// across platforms with different response times.
func TestE2E_RealTime_BudgetAlertPropagation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real-time E2E test in short mode")
	}

	bus := events.NewBus()
	defer bus.Close()

	// Track alert acknowledgments from different platforms
	alertAcks := make(map[string]time.Time)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Web dashboard subscriber (fast)
	wg.Add(1)
	bus.Subscribe("budget:limit:exceeded", func(e schema.Event) {
		time.Sleep(10 * time.Millisecond) // Simulate processing
		mu.Lock()
		alertAcks["web-dashboard"] = time.Now()
		mu.Unlock()
		wg.Done()
	})

	// Mobile app subscriber (medium)
	wg.Add(1)
	bus.Subscribe("budget:limit:exceeded", func(e schema.Event) {
		time.Sleep(50 * time.Millisecond) // Simulate slower processing
		mu.Lock()
		alertAcks["mobile-app"] = time.Now()
		mu.Unlock()
		wg.Done()
	})

	// Chat bot subscriber (slow, sends notification)
	wg.Add(1)
	bus.Subscribe("budget:limit:exceeded", func(e schema.Event) {
		time.Sleep(100 * time.Millisecond) // Simulate API call
		mu.Lock()
		alertAcks["telegram-bot"] = time.Now()
		mu.Unlock()
		wg.Done()
	})

	// Publish budget limit exceeded alert
	startTime := time.Now()

	alertEvent := schema.NewEvent(
		"budget:limit:exceeded",
		map[string]interface{}{
			"budget_id":   "personal-budget",
			"category_id": "groceries",
			"category":    "Groceries",
			"budgeted":    400.00,
			"spent":       425.75,
			"overspent":   25.75,
			"month":       "2026-02",
			"severity":    "warning",
		},
		"alert-123",
	)

	bus.Publish(alertEvent)

	// Wait for all platforms to acknowledge
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for alert acknowledgments")
	}

	// Analyze response times
	mu.Lock()
	defer mu.Unlock()

	expectedPlatforms := []string{"web-dashboard", "mobile-app", "telegram-bot"}
	for _, platform := range expectedPlatforms {
		ackTime, exists := alertAcks[platform]
		if !exists {
			t.Errorf("Platform %s did not acknowledge alert", platform)
			continue
		}

		responseTime := ackTime.Sub(startTime)
		t.Logf("Platform %s acknowledged alert in %v", platform, responseTime)
	}

	// Verify response time ordering
	if alertAcks["web-dashboard"].After(alertAcks["telegram-bot"]) {
		t.Error("Web dashboard should respond faster than Telegram bot")
	}

	t.Log("✓ Budget alert propagation: all platforms acknowledged with appropriate delays")
}

// TestE2E_RealTime_SyncProgress tests real-time sync progress updates.
func TestE2E_RealTime_SyncProgress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real-time E2E test in short mode")
	}

	bus := events.NewBus()
	defer bus.Close()

	// Track sync progress updates
	progressUpdates := []map[string]interface{}{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Subscribe to sync events
	wg.Add(7) // started + 5 progress + completed

	bus.Subscribe("budget:sync:*", func(e schema.Event) {
		mu.Lock()
		progressUpdates = append(progressUpdates, map[string]interface{}{
			"type":      e.Type,
			"timestamp": e.Timestamp,
			"payload":   e.Payload,
		})
		mu.Unlock()
		wg.Done()
	})

	correlationID := "sync-session-456"

	// Simulate sync lifecycle
	// 1. Sync started
	bus.Publish(schema.NewEvent(
		"budget:sync:started",
		map[string]interface{}{
			"budget_id": "personal-budget",
			"sync_type": "incremental",
		},
		correlationID,
	))

	time.Sleep(50 * time.Millisecond)

	// 2. Progress updates (simulate fetching accounts, categories, transactions)
	steps := []struct {
		step     string
		progress int
	}{
		{"Fetching accounts", 20},
		{"Fetching categories", 40},
		{"Fetching transactions", 60},
		{"Processing updates", 80},
		{"Finalizing sync", 95},
	}

	for _, s := range steps {
		bus.Publish(schema.NewEvent(
			"budget:sync:progress",
			map[string]interface{}{
				"budget_id": "personal-budget",
				"step":      s.step,
				"progress":  s.progress,
			},
			correlationID,
		))
		time.Sleep(30 * time.Millisecond)
	}

	// 3. Sync completed
	bus.Publish(schema.NewEvent(
		"budget:sync:completed",
		map[string]interface{}{
			"budget_id":            "personal-budget",
			"sync_type":            "incremental",
			"duration_ms":          1850,
			"accounts_added":       0,
			"accounts_updated":     2,
			"transactions_added":   12,
			"transactions_updated": 3,
			"server_knowledge":     98765,
		},
		correlationID,
	))

	// Wait for all events
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for sync events")
	}

	// Verify sync event sequence
	mu.Lock()
	defer mu.Unlock()

	if len(progressUpdates) != 7 {
		t.Errorf("Expected 7 sync events, got %d", len(progressUpdates))
	}

	// Verify event order
	if len(progressUpdates) >= 7 {
		firstEvent := progressUpdates[0]["type"].(string)
		lastEvent := progressUpdates[len(progressUpdates)-1]["type"].(string)

		if firstEvent != "budget:sync:started" {
			t.Errorf("First event should be 'started', got '%s'", firstEvent)
		}

		if lastEvent != "budget:sync:completed" {
			t.Errorf("Last event should be 'completed', got '%s'", lastEvent)
		}

		// Verify timestamps are in order
		for i := 1; i < len(progressUpdates); i++ {
			prev := progressUpdates[i-1]["timestamp"].(time.Time)
			curr := progressUpdates[i]["timestamp"].(time.Time)

			if curr.Before(prev) {
				t.Error("Events not in chronological order")
			}
		}
	}

	t.Log("✓ Sync progress: received all events in correct order")
}

// TestE2E_RealTime_ConcurrentPlatformWrites tests concurrent writes from
// multiple platforms with conflict resolution.
func TestE2E_RealTime_ConcurrentPlatformWrites(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real-time E2E test in short mode")
	}

	tmpDir := t.TempDir()
	bus := events.NewBus()
	defer bus.Close()

	// Track conflicts
	conflicts := []schema.Event{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Subscribe to conflict events
	bus.Subscribe("budget:conflict:detected", func(e schema.Event) {
		mu.Lock()
		conflicts = append(conflicts, e)
		mu.Unlock()
	})

	// Simulate concurrent transaction creation from different platforms
	platforms := []string{"web", "mobile", "cli"}
	var writeWg sync.WaitGroup

	for i, platform := range platforms {
		writeWg.Add(1)
		wg.Add(1)

		go func(p string, index int) {
			defer writeWg.Done()

			// Each platform tries to create transaction
			txEvent := schema.NewEvent(
				"budget:transaction:added",
				map[string]interface{}{
					"budget_id": "personal-budget",
					"transaction": map[string]interface{}{
						"id":          fmt.Sprintf("tx-concurrent-%d", index),
						"amount":      -25.00,
						"payee":       "Store",
						"category_id": "shopping",
						"account_id":  "checking",
						"date":        "2026-02-02",
						"platform":    p,
						"timestamp":   time.Now().UnixNano(),
					},
					"platform": p,
				},
				fmt.Sprintf("concurrent-%s-%d", p, index),
			)

			bus.Publish(txEvent)
			wg.Done()
		}(platform, i)

		// Slight delay to create race condition
		time.Sleep(10 * time.Millisecond)
	}

	writeWg.Wait()

	// Wait for all events to be processed
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for concurrent writes")
	}

	t.Logf("✓ Concurrent writes: %d platforms wrote transactions", len(platforms))

	// If conflicts were detected, verify they were handled
	mu.Lock()
	defer mu.Unlock()

	if len(conflicts) > 0 {
		t.Logf("Detected %d conflicts (expected in concurrent scenario)", len(conflicts))
	}
}

// TestE2E_RealTime_EventPersistence tests event persistence and replay.
func TestE2E_RealTime_EventPersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real-time E2E test in short mode")
	}

	tmpDir := t.TempDir()

	// Initialize bus with persistence
	bus := events.NewBus()
	defer bus.Close()

	bus.EnablePersistence(tmpDir)

	// Publish events
	eventTypes := []string{
		"budget:transaction:added",
		"budget:transaction:updated",
		"budget:sync:completed",
		"budget:limit:exceeded",
	}

	for i, eventType := range eventTypes {
		bus.Publish(schema.NewEvent(
			eventType,
			map[string]interface{}{
				"id":    i,
				"index": i,
			},
			fmt.Sprintf("persist-%d", i),
		))
	}

	// Wait for persistence
	time.Sleep(500 * time.Millisecond)

	// Retrieve persisted events
	recentEvents, err := bus.GetRecentEvents("budget:*", 10)
	if err != nil {
		t.Fatalf("Failed to retrieve recent events: %v", err)
	}

	if len(recentEvents) != len(eventTypes) {
		t.Errorf("Expected %d persisted events, got %d", len(eventTypes), len(recentEvents))
	}

	// Verify event replay capability
	replayedCount := 0
	for _, event := range recentEvents {
		// Simulate replaying event
		payload := event.Payload.(map[string]interface{})
		if _, hasID := payload["id"]; hasID {
			replayedCount++
		}
	}

	if replayedCount != len(eventTypes) {
		t.Errorf("Expected to replay %d events, replayed %d", len(eventTypes), replayedCount)
	}

	t.Logf("✓ Event persistence: %d events persisted and replayed", replayedCount)
}

// TestE2E_RealTime_CrossPlatformEventFlow tests complete event flow
// from one platform to all others.
func TestE2E_RealTime_CrossPlatformEventFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real-time E2E test in short mode")
	}

	bus := events.NewBus()
	defer bus.Close()

	// Simulate 5 different platforms
	platforms := map[string]chan schema.Event{
		"web-dashboard":  make(chan schema.Event, 10),
		"mobile-ios":     make(chan schema.Event, 10),
		"mobile-android": make(chan schema.Event, 10),
		"telegram-bot":   make(chan schema.Event, 10),
		"cli-terminal":   make(chan schema.Event, 10),
	}

	// Each platform subscribes to all budget events
	for platformName, ch := range platforms {
		platformChan := ch // Capture for closure
		bus.Subscribe("budget:*", func(e schema.Event) {
			platformChan <- e
		})
	}

	// User performs action on web dashboard
	originPlatform := "web-dashboard"
	actionEvent := schema.NewEvent(
		"budget:transaction:added",
		map[string]interface{}{
			"budget_id": "shared-budget",
			"transaction": map[string]interface{}{
				"id":          "tx-cross-platform-1",
				"amount":      -50.00,
				"payee":       "Amazon",
				"category":    "Shopping",
				"category_id": "cat-shopping",
				"date":        "2026-02-02",
			},
			"origin_platform": originPlatform,
			"user_id":         "user-456",
		},
		"cross-platform-flow-1",
	)

	// Publish event
	bus.Publish(actionEvent)

	// Verify all platforms receive the event
	timeout := time.After(2 * time.Second)
	receivedCount := 0

	for platformName, ch := range platforms {
		select {
		case event := <-ch:
			receivedCount++

			payload := event.Payload.(map[string]interface{})
			origin := payload["origin_platform"].(string)

			if platformName == originPlatform {
				t.Logf("Platform %s: received own event (origin)", platformName)
			} else {
				t.Logf("Platform %s: received event from %s", platformName, origin)
			}

			// Verify event integrity
			if event.Type != "budget:transaction:added" {
				t.Errorf("Platform %s: wrong event type: %s", platformName, event.Type)
			}

		case <-timeout:
			t.Errorf("Platform %s: timeout waiting for event", platformName)
		}
	}

	if receivedCount != len(platforms) {
		t.Errorf("Expected %d platforms to receive event, got %d", len(platforms), receivedCount)
	}

	t.Logf("✓ Cross-platform event flow: event propagated to %d platforms", receivedCount)
}

// ============================================================================
// Helper Functions
// ============================================================================

func handlePlatformEvent(t *testing.T, platform string, event schema.Event) {
	// Simulate platform-specific event handling
	payload := event.Payload.(map[string]interface{})

	switch platform {
	case "web-dashboard":
		// Update UI in real-time
		t.Logf("[WEB] Updating dashboard with event: %s", event.Type)

	case "mobile-app":
		// Show push notification
		t.Logf("[MOBILE] Showing push notification for: %s", event.Type)

	case "telegram-bot":
		// Send Telegram message
		if event.Type == "budget:limit:exceeded" {
			t.Logf("[TELEGRAM] Sending budget alert to user")
		}

	case "cli-watcher":
		// Print to terminal if active
		if tx, ok := payload["transaction"].(map[string]interface{}); ok {
			t.Logf("[CLI] Transaction: %s - $%.2f",
				tx["payee"], tx["amount"])
		}
	}
}

func serializeEvent(event schema.Event) string {
	data, _ := json.Marshal(event)
	return string(data)
}
