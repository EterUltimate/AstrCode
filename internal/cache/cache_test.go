// AstrCode - Agent orchestration engine for AstrBot
// Copyright (C) 2026 EterUltimate
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package cache

import (
	"testing"
	"time"
)

func TestMemoryCache(t *testing.T) {
	cache := NewMemoryCache()

	// Set + Get
	cache.Set("key1", "value1", 5*time.Minute)
	val, ok := cache.Get("key1")
	if !ok || val != "value1" {
		t.Errorf("expected value1, got %v, ok=%v", val, ok)
	}

	// Missing key
	_, ok = cache.Get("nonexistent")
	if ok {
		t.Error("expected not found")
	}

	// Delete
	cache.Delete("key1")
	_, ok = cache.Get("key1")
	if ok {
		t.Error("expected key1 to be deleted")
	}

	// Clear
	cache.Set("a", 1, 5*time.Minute)
	cache.Set("b", 2, 5*time.Minute)
	cache.Clear()
	_, okA := cache.Get("a")
	_, okB := cache.Get("b")
	if okA || okB {
		t.Error("expected all keys cleared")
	}
}

func TestMemoryCacheExpiry(t *testing.T) {
	cache := NewMemoryCache()
	cache.Set("expiring", "data", 50*time.Millisecond)

	val, ok := cache.Get("expiring")
	if !ok || val != "data" {
		t.Error("should exist before expiry")
	}

	time.Sleep(100 * time.Millisecond)
	_, ok = cache.Get("expiring")
	if ok {
		t.Error("should be expired")
	}
}

func TestSkillCache(t *testing.T) {
	cache := NewSkillCache()
	args := map[string]interface{}{"path": "/test"}

	// Miss
	_, ok := cache.GetSkillResult("read_code", args)
	if ok {
		t.Error("expected cache miss")
	}

	// Set + Hit
	cache.SetSkillResult("read_code", args, "file contents", 5*time.Minute)
	result, ok := cache.GetSkillResult("read_code", args)
	if !ok || result != "file contents" {
		t.Errorf("expected 'file contents', got '%s'", result)
	}
}

func TestHashKey(t *testing.T) {
	key1 := HashKey("hello")
	key2 := HashKey("hello")
	key3 := HashKey("world")

	if key1 != key2 {
		t.Error("same input should produce same hash")
	}
	if key1 == key3 {
		t.Error("different input should produce different hash")
	}
}
