package uuid

import (
	"os"
	"testing"
)

func TestV4(t *testing.T) {
	uuid := MakeV4()
	if uuid.Version() != 4 {
		t.Fatalf("Invalid V4 UUID: version != 4")
	}
	if uuid[6]>>4 != 4 {
		t.Fatalf("Invalid V4 UUID: version != 4")
	}
	msb := uuid[8] >> 4
	if msb != 0x8 && msb != 0x9 && msb != 0xa && msb != 0xb {
		t.Fatalf("Invalid V4 UUID: some bits [0x%x] are wrong", msb)
	}
	if len(uuid.String()) != 36 {
		t.Fatalf("Invalid V4 UUID: %s", uuid.String())
	}
	if uuid.String()[8] != '-' {
		t.Fatal("Expected dash in position 8")
	}
	if uuid.String()[13] != '-' {
		t.Fatal("Expected dash in position 13")
	}
	if uuid.String()[18] != '-' {
		t.Fatal("Expected dash in position 18")
	}
	if uuid.String()[23] != '-' {
		t.Fatal("Expected dash in position 23")
	}
}

func TestParse(t *testing.T) {
	str := "9b78d54c-8cc9-46bc-ae29-efcba10e1abb"
	uuid, err := Parse(str)
	if err != nil {
		t.Fatalf("Parsing failed")
	}
	if uuid.Version() != 4 {
		t.Fatalf("UUID version should be 4")
	}
	if uuid[0] != 0x9b || uuid[len(uuid)-1] != 0xbb {
		t.Fatalf("UUID value is wrong")
	}
	if str != uuid.String() {
		t.Fatalf("UUID value is wrong")
	}
}

func TestParseV4(t *testing.T) {
	for i := 0; i < 100; i++ {
		uuid := MakeV4()
		uuid2, err := Parse(uuid.String())
		if err != nil {
			t.Fatalf("Parsing of %v failed", uuid)
		}
		if !uuid2.Equal(uuid) {
			t.Fatalf("UUIDs are not equal")
		}
	}
}

func TestParseGood(t *testing.T) {
	good := []string{
		"9ABCDEF0-8cc9-46bc-ae29-efcba10e1abb",
		"{9ABCDEF0-8cc9-46bc-ae29-efcba10e1abb}",
	}
	for _, str := range good {
		if _, err := Parse(str); err != nil {
			t.Fatalf("Parsing of %s should succeed", str)
		}
	}
}

func TestParseErrors(t *testing.T) {
	bad := []string{
		"9b78d54c-8cc9-46bc-ae29-efcba10e1ab",
		"{9b78d54c-8cc9-46bc-ae29-efcba10e1abb",
		"9b78d54c-8cc9-46bc-ae29-efcba10e1abb}",
		"{9b78d54cx8cc9-46bc-ae29-efcba10e1abb}",
		"{9b78d54c-8cc9-46bc-ae29-efcba10e1abb]",
		"[9b78d54c-8cc9-46bc-ae29-efcba10e1abb]",
		"9b78d54cx8cc9-46bc-ae29-efcba10e1abb",
		"9b78d54c-8cc9x46bc-ae29-efcba10e1abb",
		"9b78d54c-8cc9-46bcxae29-efcba10e1abb",
		"9b78d54c-8cc9-46bc-ae29xefcba10e1abb",
		"9bP8d54c-8cc9-46bc-ae29-efcba10e1abb",
		"9b78d54c-8cc9-46bc-ae29-efcba10e1abX",
	}
	for _, str := range bad {
		if _, err := Parse(str); err != os.EINVAL {
			t.Fatalf("Parsing of %s should have failed", str)
		}
	}
}

func TestKey(t *testing.T) {
	uuid := MakeV4()
	if len(uuid.Key()) != 16 {
		t.Fatal("Key for V4 UUID is wrong")
	}
}

func BenchmarkMakeV4(b *testing.B) {
	for n := b.N; n > 0; n-- {
		_ = MakeV4()
	}
}
