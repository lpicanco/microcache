package microcache

import "testing"

func TestPutGet(t *testing.T) {
	cache := NewCache(DefaultConfiguration(100))

	structValue := struct {
		key   int32
		value string
	}{
		42, "answer",
	}

	cases := []struct {
		in   string
		want interface{}
	}{
		{"Integer value", 432},
		{"String value", "string key"},
		{"String value", "string key 2"},
		{"Array value", [3]string{"Val01", "Val02", "Val03"}},
		{"Struct value", structValue},
		{"Struct reference value", &structValue},
	}
	for _, c := range cases {
		cache.Put(c.in, c.want)

		got, found := cache.Get(c.in)
		if !found {
			t.Errorf("Cache.Get(%q) not found", c.in)
		}

		if got != c.want {
			t.Errorf("Cache.Get(%q) == %v, want %v", c.in, got, c.want)
		}
	}
}

func TestNotFound(t *testing.T) {
	cache := NewCache(DefaultConfiguration(100))
	got, found := cache.Get("key")

	if found {
		t.Error("Cache.Get(key) returned true")
	}

	if got != nil {
		t.Errorf("Cache.Get(key) == %v, want %v", got, nil)
	}

}
