package cache

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	testKey      string = "cache:test"
	testKeyEmpty string = "cache:empty"
)

// AppCache создание нового кэша
var AppCache = New(1*time.Minute, 2*time.Minute)
var AppCache1 = New(1*time.Minute, 2*time.Minute)

// TestGet получить контент по ключу
func TestGet(t *testing.T) {

	testValue := map[string]string{
		"test":  "test",
		"test2": "test2",
		"test3": "test3",
	}

	AppCache.Set(testKey, true, testValue)

	value, _, ok := AppCache.Get(testKey)

	if ok != true {
		t.Error("Ошибка: ", "не получили нужный value")
	}

	assert.Equal(t, value, testValue)

	value, _, ok = AppCache.Get(testKeyEmpty)

	if value != nil || ok != false {
		t.Error("Ошибка: ", "value не должно быть и мы его должны были не найти", value)
	}
}

// Проверка выстеснения
func TestCache(t *testing.T) {
	testValue := map[string]int{
		"a": 1,
		"b": 1,
		"c": 1,
		"d": 1,
		"e": 1,
		"f": 1,
		"g": 1,
		"h": 1,
		"i": 1,
		"j": 1,
		"k": 1,
		"l": 1,
		"m": 1,
		"n": 1,
		"o": 1,
		"p": 1,
		"q": 50,
		"r": 50,
		"s": 50,
		"t": 50,
	}
	testMap := map[string]string{
		"test":  "test",
		"test2": "test2",
		"test3": "test3",
	}
	for key := range testValue {
		AppCache1.Set(key, true, testMap)
	}

	for key := range testValue {
		_, _, ok := AppCache1.Get(key)
		if ok {
			t.Error("Ошибка вытеснения")
		}
	}
}
