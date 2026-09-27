package bot

import (
	"encoding/json"
	"os"
	"sync"
	"testing"
)

func TestHoneypotStateManagement(t *testing.T) {
	// Verify HoneypotState structure field types
	state := &HoneypotState{
		ChannelID: "123",
		MessageID: "456",
		BanCount:  0,
	}

	if state.ChannelID != "123" {
		t.Errorf("expected ChannelID to be '123', got '%s'", state.ChannelID)
	}

	// Test thread-safe read/write concurrency to Honeypots map
	honeypots := make(map[string]*HoneypotState)
	var mu sync.RWMutex

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(id int) {
			defer wg.Done()
			mu.Lock()
			honeypots["guild-1"] = &HoneypotState{
				ChannelID: "channel-1",
				MessageID: "msg-1",
				BanCount:  id,
			}
			mu.Unlock()
		}(i)

		go func() {
			defer wg.Done()
			mu.RLock()
			_ = honeypots["guild-1"]
			mu.RUnlock()
		}()
	}
	wg.Wait()
}

func TestHoneypotPersistence(t *testing.T) {
	// Clean up environment and files
	os.Unsetenv(HoneypotsEnvVar)
	os.Remove(".env")
	os.Remove(legacyHoneypotsFile)
	defer func() {
		os.Unsetenv(HoneypotsEnvVar)
		os.Remove(".env")
		os.Remove(legacyHoneypotsFile)
	}()

	// Create a dummy BotService
	bs := &BotService{
		Honeypots: make(map[string]*HoneypotState),
	}

	// Set some test honeypots
	bs.Honeypots["guild-123"] = &HoneypotState{
		ChannelID: "channel-abc",
		MessageID: "msg-xyz",
		BanCount:  42,
	}

	// Save
	bs.SaveHoneypots()

	// Verify env var was updated
	if os.Getenv(HoneypotsEnvVar) == "" {
		t.Fatalf("expected HONEYPOTS env var to be set")
	}

	// Create another dummy BotService and load
	bs2 := &BotService{
		Honeypots: make(map[string]*HoneypotState),
	}
	bs2.LoadHoneypots()

	state, ok := bs2.Honeypots["guild-123"]
	if !ok {
		t.Fatalf("expected guild-123 to be loaded")
	}

	if state.ChannelID != "channel-abc" || state.MessageID != "msg-xyz" || state.BanCount != 42 {
		t.Errorf("loaded honeypot state is incorrect: %+v", state)
	}
}

func TestHoneypotLegacyMigration(t *testing.T) {
	os.Unsetenv(HoneypotsEnvVar)
	os.Remove(".env")
	os.Remove(legacyHoneypotsFile)
	defer func() {
		os.Unsetenv(HoneypotsEnvVar)
		os.Remove(".env")
		os.Remove(legacyHoneypotsFile)
	}()

	// Create a legacy honeypots.json file
	legacyData := map[string]*HoneypotState{
		"guild-legacy": {
			ChannelID: "channel-old",
			MessageID: "msg-old",
			BanCount:  10,
		},
	}
	bytes, _ := json.Marshal(legacyData)
	os.WriteFile(legacyHoneypotsFile, bytes, 0644)

	// Load should migrate it to env & .env
	bs := &BotService{}
	bs.LoadHoneypots()

	state, ok := bs.Honeypots["guild-legacy"]
	if !ok {
		t.Fatalf("expected legacy guild-legacy to be loaded")
	}
	if state.ChannelID != "channel-old" || state.BanCount != 10 {
		t.Errorf("migrated state is incorrect: %+v", state)
	}

	// legacy file should be removed
	if _, err := os.Stat(legacyHoneypotsFile); !os.IsNotExist(err) {
		t.Errorf("expected legacy file to be removed after migration")
	}

	// env variable should now be set
	if os.Getenv(HoneypotsEnvVar) == "" {
		t.Errorf("expected HONEYPOTS env var to be set after migration")
	}
}
