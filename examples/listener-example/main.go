package main

import (
	"fmt"
	"time"

	"github.com/click33/sa-token-go/core"
	"github.com/click33/sa-token-go/storage/memory"
	"github.com/click33/sa-token-go/stputil"
)

func main() {
	fmt.Println("=== Sa-Token-Go Event Listener Example ===\n")

	// Create event manager
	eventMgr := core.NewEventManager()

	// 1. Simple function listener
	eventMgr.RegisterFunc(core.EventLogin, func(data *core.EventData) {
		fmt.Printf("[LOGIN] User %s logged in with token %s\n", data.LoginID, data.Token[:20]+"...")
	})

	// 2. Logout listener
	eventMgr.RegisterFunc(core.EventLogout, func(data *core.EventData) {
		fmt.Printf("[LOGOUT] User %s logged out\n", data.LoginID)
	})

	// 3. Kickout listener
	eventMgr.RegisterFunc(core.EventKickout, func(data *core.EventData) {
		fmt.Printf("[KICKOUT] User %s was forcibly logged out\n", data.LoginID)
	})

	// 4. High-priority synchronous listener
	eventMgr.RegisterWithConfig(core.EventLogin,
		core.ListenerFunc(func(data *core.EventData) {
			fmt.Printf("[AUDIT] Login audit - User: %s, Time: %d\n",
				data.LoginID, data.Timestamp)
		}),
		core.ListenerConfig{
			Async:    false, // Synchronous
			Priority: 100,   // High priority
			ID:       "audit-logger",
		},
	)

	// 5. Wildcard listener (all events)
	eventMgr.RegisterFunc(core.EventAll, func(data *core.EventData) {
		fmt.Printf("[ALL EVENTS] %s\n", data.String())
	})

	// 6. Custom panic handler
	eventMgr.SetPanicHandler(func(event core.Event, data *core.EventData, recovered interface{}) {
		fmt.Printf("[PANIC RECOVERED] Event: %s, Error: %v\n", event, recovered)
	})

	// Initialize Sa-Token
	stputil.SetManager(
		core.NewBuilder().
			Storage(memory.NewStorage()).
			TokenName("Authorization").
			Timeout(7200).
			Build(),
	)

	fmt.Println("\n--- Triggering Events ---\n")

	// Trigger login event
	token1, _ := stputil.Login(1000)
	time.Sleep(100 * time.Millisecond) // Wait for async listeners

	token2, _ := stputil.Login(2000)
	time.Sleep(100 * time.Millisecond)

	// Trigger logout event
	stputil.Logout(1000)
	time.Sleep(100 * time.Millisecond)

	// Trigger kickout event
	stputil.Kickout(2000)
	time.Sleep(100 * time.Millisecond)

	// Wait for all async listeners to complete
	eventMgr.Wait()

	fmt.Println("\n--- Listener Statistics ---")
	fmt.Printf("Total listeners: %d\n", eventMgr.Count())
	fmt.Printf("Login listeners: %d\n", eventMgr.CountForEvent(core.EventLogin))
	fmt.Printf("Logout listeners: %d\n", eventMgr.CountForEvent(core.EventLogout))

	// Unregister a listener
	fmt.Println("\n--- Unregistering audit logger ---")
	if eventMgr.Unregister("audit-logger") {
		fmt.Println("Audit logger unregistered successfully")
	}

	fmt.Printf("Remaining listeners: %d\n", eventMgr.Count())

	// Disable certain events
	fmt.Println("\n--- Disabling kickout events ---")
	eventMgr.DisableEvent(core.EventKickout)

	fmt.Println("\n--- Testing event disable (this should not trigger kickout listener) ---")
	stputil.Login(3000)
	stputil.Kickout(3000)
	time.Sleep(100 * time.Millisecond)

	// Re-enable all events
	eventMgr.EnableEvent()

	fmt.Println("\n=== Example Complete ===")

	// Cleanup
	fmt.Println("\nTokens:")
	fmt.Printf("Token 1: %s\n", token1)
	fmt.Printf("Token 2: %s\n", token2)
}
