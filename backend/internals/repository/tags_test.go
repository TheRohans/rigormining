package repository

import (
	"errors"
	"reflect"
	"testing"

	"gitlab.com/robrohan/rigormining/internals/models"
)

func TestRenameTag(t *testing.T) {
	r := testRepo(t)
	alice := testUser(t, r, "alice@test")
	bob := testUser(t, r, "bob@test")

	a1 := testItem(t, r, alice, "music")
	a2 := testItem(t, r, alice, "music", "Music") // already has both
	a3 := testItem(t, r, alice, "physics")
	b1 := testItem(t, r, bob, "music")

	n, err := r.RenameTag(alice, "music", "Music")
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("renamed %d items, want 2", n)
	}

	for id, want := range map[string][]string{
		a1: {"Music"},
		a2: {"Music"}, // merged, not duplicated
		a3: {"physics"},
		b1: {"music"}, // another user's tag is untouched
	} {
		if got := tagNames(t, r, id); !reflect.DeepEqual(got, want) {
			t.Errorf("item tags = %v, want %v", got, want)
		}
	}

	tags, err := r.ListTags(alice)
	if err != nil {
		t.Fatal(err)
	}
	want := []models.TagSummary{{Name: "Music", Count: 2}, {Name: "physics", Count: 1}}
	if !reflect.DeepEqual(tags, want) {
		t.Errorf("ListTags(alice) = %+v, want %+v", tags, want)
	}

	// Bob still uses "music", so its shared row must survive...
	var count int
	r.Db.Get(&count, `SELECT COUNT(*) FROM tag WHERE name = 'music'`)
	if count != 1 {
		t.Errorf("shared 'music' row deleted while bob still uses it")
	}
	// ...until bob renames his too, when it's cleaned up.
	if _, err := r.RenameTag(bob, "music", "Music"); err != nil {
		t.Fatal(err)
	}
	r.Db.Get(&count, `SELECT COUNT(*) FROM tag WHERE name = 'music'`)
	if count != 0 {
		t.Errorf("unused 'music' row left behind")
	}

	if _, err := r.RenameTag(alice, "nope", "x"); !errors.Is(err, ErrTagNotFound) {
		t.Errorf("unknown tag: err = %v, want ErrTagNotFound", err)
	}
	// Bob's tag exists, but alice has no items with it.
	testItem(t, r, bob, "bobs-only")
	if _, err := r.RenameTag(alice, "bobs-only", "x"); !errors.Is(err, ErrTagNotFound) {
		t.Errorf("another user's tag: err = %v, want ErrTagNotFound", err)
	}
}
