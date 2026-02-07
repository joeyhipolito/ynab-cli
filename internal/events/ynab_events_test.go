package events

import (
	"sync"
	"testing"
	"time"

	"github.com/joeyhipolito/via/features/events/schema"
	viaevents "github.com/joeyhipolito/via/features/events"
)

// ============================================================================
// Event Publishing Tests
// ============================================================================

// TestYNABEventBudgetTransactionAdded tests publishing transaction added events.
func TestYNABEventBudgetTransactionAdded(t *testing.T) {
	bus := viaevents.NewBus()
	defer bus.Close()

	var received schema.Event
	var wg sync.WaitGroup
	wg.Add(1)

	// Subscribe to transaction added events
	bus.Subscribe("budget:transaction:added", func(e schema.Event) {
		received = e
		wg.Done()
	})

	// Publish event
	event := schema.NewEvent(
		"budget:transaction:added",
		map[string]interface{}{
			"budget_id":   "test-budget",
			"transaction": map[string]interface{}{
				"id":          "tx-123",
				"amount":      -50.00,
				"payee":       "Coffee Shop",
				"category":    "Dining Out",
				"date":        "2026-02-02",
				"account_id":  "acct-456",
				"category_id": "cat-789",
			},
		},
		"corr-123",
	)

	bus.Publish(event)

	// Wait for event with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for event")
	}

	// Verify event
	if received.Type != "budget:transaction:added" {
		t.Errorf("Expected event type 'budget:transaction:added', got %s", received.Type)
	}

	payload, ok := received.Payload.(map[string]interface{})
	if !ok {
		t.Fatal("Event payload is not a map")
	}

	if payload["budget_id"] != "test-budget" {
		t.Errorf("Expected budget_id 'test-budget', got %v", payload["budget_id"])
	}
}

// TestYNABEventBudgetLimitExceeded tests publishing budget limit exceeded events.
func TestYNABEventBudgetLimitExceeded(t *testing.T) {
	bus := viaevents.NewBus()
	defer bus.Close()

	var received schema.Event
	var wg sync.WaitGroup
	wg.Add(1)

	bus.Subscribe("budget:limit:exceeded", func(e schema.Event) {
		received = e
		wg.Done()
	})

	event := schema.NewEvent(
		"budget:limit:exceeded",
		map[string]interface{}{
			"budget_id":   "test-budget",
			"category_id": "groceries",
			"category":    "Groceries",
			"budgeted":    400.00,
			"spent":       425.50,
			"overspent":   25.50,
			"month":       "2026-02",
		},
		"corr-456",
	)

	bus.Publish(event)

	// Wait for event
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for event")
	}

	if received.Type != "budget:limit:exceeded" {
		t.Errorf("Expected event type 'budget:limit:exceeded', got %s", received.Type)
	}

	payload := received.Payload.(map[string]interface{})
	if payload["category"] != "Groceries" {
		t.Errorf("Expected category 'Groceries', got %v", payload["category"])
	}

	if payload["overspent"] != 25.50 {
		t.Errorf("Expected overspent 25.50, got %v", payload["overspent"])
	}
}

// TestYNABEventSyncCompleted tests publishing sync completed events.
func TestYNABEventSyncCompleted(t *testing.T) {
	bus := viaevents.NewBus()
	defer bus.Close()

	var received schema.Event
	var wg sync.WaitGroup
	wg.Add(1)

	bus.Subscribe("budget:sync:completed", func(e schema.Event) {
		received = e
		wg.Done()
	})

	event := schema.NewEvent(
		"budget:sync:completed",
		map[string]interface{}{
			"budget_id":          "test-budget",
			"sync_type":          "incremental",
			"duration_ms":        1250,
			"transactions_added": 15,
			"transactions_updated": 3,
			"server_knowledge":   12345,
		},
		"corr-sync-1",
	)

	bus.Publish(event)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for event")
	}

	if received.Type != "budget:sync:completed" {
		t.Errorf("Expected 'budget:sync:completed', got %s", received.Type)
	}

	payload := received.Payload.(map[string]interface{})
	if payload["transactions_added"] != 15 {
		t.Errorf("Expected 15 transactions added, got %v", payload["transactions_added"])
	}
}

// TestYNABEventSyncFailed tests publishing sync failed events.
func TestYNABEventSyncFailed(t *testing.T) {
	bus := viaevents.NewBus()
	defer bus.Close()

	var received schema.Event
	var wg sync.WaitGroup
	wg.Add(1)

	bus.Subscribe("budget:sync:failed", func(e schema.Event) {
		received = e
		wg.Done()
	})

	event := schema.NewEvent(
		"budget:sync:failed",
		map[string]interface{}{
			"budget_id": "test-budget",
			"error":     "rate_limit_exceeded",
			"message":   "YNAB API rate limit exceeded, will retry in 60 seconds",
			"retry_after": 60,
		},
		"corr-sync-fail",
	)

	bus.Publish(event)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for event")
	}

	if received.Type != "budget:sync:failed" {
		t.Errorf("Expected 'budget:sync:failed', got %s", received.Type)
	}

	payload := received.Payload.(map[string]interface{})
	if payload["error"] != "rate_limit_exceeded" {
		t.Errorf("Expected error 'rate_limit_exceeded', got %v", payload["error"])
	}
}

// ============================================================================
// Event Subscription Tests
// ============================================================================

// TestYNABEventWildcardSubscription tests wildcard subscriptions for budget events.
func TestYNABEventWildcardSubscription(t *testing.T) {
	bus := viaevents.NewBus()
	defer bus.Close()

	receivedEvents := []schema.Event{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(3) // Expecting 3 events

	// Subscribe to all budget events with wildcard
	bus.Subscribe("budget:*", func(e schema.Event) {
		mu.Lock()
		receivedEvents = append(receivedEvents, e)
		mu.Unlock()
		wg.Done()
	})

	// Publish different budget events
	events := []schema.Event{
		schema.NewEvent("budget:transaction:added", map[string]interface{}{"id": "1"}, "corr-1"),
		schema.NewEvent("budget:sync:completed", map[string]interface{}{"id": "2"}, "corr-2"),
		schema.NewEvent("budget:limit:exceeded", map[string]interface{}{"id": "3"}, "corr-3"),
	}

	for _, event := range events {
		bus.Publish(event)
	}

	// Wait for all events
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for events")
	}

	mu.Lock()
	defer mu.Unlock()

	if len(receivedEvents) != 3 {
		t.Errorf("Expected 3 events, got %d", len(receivedEvents))
	}

	// Verify all event types were received
	eventTypes := make(map[string]bool)
	for _, e := range receivedEvents {
		eventTypes[e.Type] = true
	}

	expectedTypes := []string{
		"budget:transaction:added",
		"budget:sync:completed",
		"budget:limit:exceeded",
	}

	for _, expected := range expectedTypes {
		if !eventTypes[expected] {
			t.Errorf("Expected event type %s not received", expected)
		}
	}
}

// TestYNABEventMultipleSubscribers tests multiple subscribers to same event.
func TestYNABEventMultipleSubscribers(t *testing.T) {
	bus := viaevents.NewBus()
	defer bus.Close()

	var wg sync.WaitGroup
	wg.Add(3) // 3 subscribers

	received1 := false
	received2 := false
	received3 := false

	// Subscribe multiple handlers to same event
	bus.Subscribe("budget:transaction:added", func(e schema.Event) {
		received1 = true
		wg.Done()
	})

	bus.Subscribe("budget:transaction:added", func(e schema.Event) {
		received2 = true
		wg.Done()
	})

	bus.Subscribe("budget:transaction:added", func(e schema.Event) {
		received3 = true
		wg.Done()
	})

	// Publish single event
	event := schema.NewEvent(
		"budget:transaction:added",
		map[string]interface{}{"id": "test"},
		"corr-multi",
	)

	bus.Publish(event)

	// Wait for all subscribers
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for subscribers")
	}

	if !received1 || !received2 || !received3 {
		t.Error("Not all subscribers received the event")
	}
}

// TestYNABEventUnsubscribe tests unsubscribing from events.
func TestYNABEventUnsubscribe(t *testing.T) {
	bus := viaevents.NewBus()
	defer bus.Close()

	callCount := 0
	var mu sync.Mutex

	// Subscribe and get subscription ID
	subID := bus.Subscribe("budget:transaction:added", func(e schema.Event) {
		mu.Lock()
		callCount++
		mu.Unlock()
	})

	// Publish first event
	event1 := schema.NewEvent("budget:transaction:added", map[string]interface{}{"id": "1"}, "corr-1")
	bus.Publish(event1)
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	firstCount := callCount
	mu.Unlock()

	// Unsubscribe
	bus.Unsubscribe(subID)

	// Publish second event
	event2 := schema.NewEvent("budget:transaction:added", map[string]interface{}{"id": "2"}, "corr-2")
	bus.Publish(event2)
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	secondCount := callCount
	mu.Unlock()

	// Verify first event was received
	if firstCount != 1 {
		t.Errorf("Expected 1 call before unsubscribe, got %d", firstCount)
	}

	// Verify second event was NOT received
	if secondCount != 1 {
		t.Errorf("Expected 1 call after unsubscribe (no change), got %d", secondCount)
	}
}

// ============================================================================
// Store Integration Tests
// ============================================================================

// TestYNABEventStoreIntegration tests event storage and retrieval.
func TestYNABEventStoreIntegration(t *testing.T) {
	bus := viaevents.NewBus()
	defer bus.Close()

	// Enable event storage
	bus.EnablePersistence(t.TempDir())

	// Publish multiple events
	for i := 0; i < 10; i++ {
		event := schema.NewEvent(
			"budget:transaction:added",
			map[string]interface{}{
				"id":     i,
				"amount": -50.00 * float64(i),
			},
			fmt.Sprintf("corr-%d", i),
		)
		bus.Publish(event)
	}

	// Allow time for persistence
	time.Sleep(500 * time.Millisecond)

	// Retrieve recent events
	events, err := bus.GetRecentEvents("budget:transaction:added", 5)
	if err != nil {
		t.Fatalf("GetRecentEvents failed: %v", err)
	}

	if len(events) != 5 {
		t.Errorf("Expected 5 recent events, got %d", len(events))
	}

	// Verify events are in reverse chronological order (most recent first)
	for i := 0; i < len(events)-1; i++ {
		if events[i].Timestamp.Before(events[i+1].Timestamp) {
			t.Error("Events not in reverse chronological order")
		}
	}
}

// TestYNABEventRecentEventsQuery tests querying recent events with filters.
func TestYNABEventRecentEventsQuery(t *testing.T) {
	bus := viaevents.NewBus()
	defer bus.Close()

	bus.EnablePersistence(t.TempDir())

	// Publish mix of events
	eventTypes := []string{
		"budget:transaction:added",
		"budget:sync:completed",
		"budget:limit:exceeded",
	}

	for i := 0; i < 15; i++ {
		eventType := eventTypes[i%3]
		event := schema.NewEvent(
			eventType,
			map[string]interface{}{"index": i},
			fmt.Sprintf("corr-%d", i),
		)
		bus.Publish(event)
	}

	time.Sleep(500 * time.Millisecond)

	// Query only transaction events
	txEvents, err := bus.GetRecentEvents("budget:transaction:added", 10)
	if err != nil {
		t.Fatalf("GetRecentEvents failed: %v", err)
	}

	// Should get 5 transaction events (every 3rd event)
	if len(txEvents) != 5 {
		t.Errorf("Expected 5 transaction events, got %d", len(txEvents))
	}

	// Verify all are transaction events
	for _, e := range txEvents {
		if e.Type != "budget:transaction:added" {
			t.Errorf("Expected 'budget:transaction:added', got %s", e.Type)
		}
	}

	// Query with wildcard
	allBudgetEvents, err := bus.GetRecentEvents("budget:*", 20)
	if err != nil {
		t.Fatalf("GetRecentEvents with wildcard failed: %v", err)
	}

	// Should get all 15 events
	if len(allBudgetEvents) != 15 {
		t.Errorf("Expected 15 budget events, got %d", len(allBudgetEvents))
	}
}

// ============================================================================
// Event Correlation Tests
// ============================================================================

// TestYNABEventCorrelation tests event correlation IDs.
func TestYNABEventCorrelation(t *testing.T) {
	bus := viaevents.NewBus()
	defer bus.Close()

	correlationID := "sync-workflow-123"
	var receivedEvents []schema.Event
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Subscribe to all budget events
	bus.Subscribe("budget:*", func(e schema.Event) {
		mu.Lock()
		receivedEvents = append(receivedEvents, e)
		mu.Unlock()
		wg.Done()
	})

	// Publish related events with same correlation ID
	events := []schema.Event{
		schema.NewEvent("budget:sync:started", map[string]interface{}{"step": 1}, correlationID),
		schema.NewEvent("budget:sync:progress", map[string]interface{}{"step": 2}, correlationID),
		schema.NewEvent("budget:sync:completed", map[string]interface{}{"step": 3}, correlationID),
	}

	wg.Add(len(events))
	for _, event := range events {
		bus.Publish(event)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for correlated events")
	}

	mu.Lock()
	defer mu.Unlock()

	// Verify all events have same correlation ID
	for _, e := range receivedEvents {
		if e.CorrelationID != correlationID {
			t.Errorf("Expected correlation ID %s, got %s", correlationID, e.CorrelationID)
		}
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func fmt.Sprintf(format string, args ...interface{}) string {
	return ""
}
