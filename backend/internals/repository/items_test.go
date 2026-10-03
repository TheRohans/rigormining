package repository

import (
	"testing"
)

func TestDeleteItemRemovesLinks(t *testing.T) {
	r := testRepo(t)
	alice := testUser(t, r, "alice@test")
	bob := testUser(t, r, "bob@test")

	a1 := testItem(t, r, alice, "music", "physics")
	b1 := testItem(t, r, bob, "music")

	// Postgres rejects this outright if the tag links are left behind.
	if err := r.DeleteItem(a1, alice); err != nil {
		t.Fatal(err)
	}
	if _, err := r.GetItemById(a1); err == nil {
		t.Error("item still exists after delete")
	}
	var links int
	r.Db.Get(&links, r.Db.Rebind(`SELECT COUNT(*) FROM item_tag WHERE item_uuid = ?`), a1)
	if links != 0 {
		t.Errorf("%d tag links left behind", links)
	}

	// Deleting someone else's item is a no-op, links included.
	if err := r.DeleteItem(b1, alice); err != nil {
		t.Fatal(err)
	}
	if _, err := r.GetItemById(b1); err != nil {
		t.Error("alice deleted bob's item")
	}
	r.Db.Get(&links, r.Db.Rebind(`SELECT COUNT(*) FROM item_tag WHERE item_uuid = ?`), b1)
	if links != 1 {
		t.Errorf("bob's item has %d tag links, want 1", links)
	}
}

func TestListItemsSearch(t *testing.T) {
	r := testRepo(t)
	alice := testUser(t, r, "alice@test")

	titled := func(title, authors string) {
		id := testItem(t, r, alice)
		item, _ := r.GetItemById(id)
		item.Title, item.Authors = title, authors
		if err := r.UpdateItem(item); err != nil {
			t.Fatal(err)
		}
	}
	titled("Attention Is All You Need", "Vaswani, Ashish")
	titled("100% Accuracy Is a Myth", "Smith, Jane")
	titled("snake_case Considered Harmful", "Jones, Bob")
	titled("Snakes and Ladders", "Brown, Al")

	for q, want := range map[string]int{
		"attention": 1, // case-insensitive (Postgres LIKE isn't by default)
		"ATTENTION": 1,
		"vaswani":   1, // authors too
		"100%":      1, // % is literal, not "anything"
		"%":         1,
		"snake_":    1, // _ is literal, not "any one character"
		"snake":     2,
		"nothing":   0,
	} {
		items, err := r.ListItems(alice, ItemFilter{Query: q})
		if err != nil {
			t.Fatalf("q=%q: %v", q, err)
		}
		if len(items) != want {
			t.Errorf("q=%q: %d items, want %d", q, len(items), want)
		}
	}
}
