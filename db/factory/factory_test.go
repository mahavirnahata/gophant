package factory

import "testing"

var UserFactory = New(func(f *Context) map[string]any {
	return map[string]any{
		"name":  f.Name(),
		"email": f.Email(),
		"role":  "member",
	}
})

func TestMake_ReturnsMap(t *testing.T) {
	row := UserFactory.Make()
	if row["name"] == "" {
		t.Fatal("expected non-empty name")
	}
	if row["role"] != "member" {
		t.Fatalf("expected role=member, got %v", row["role"])
	}
}

func TestMake_EmailIsUnique(t *testing.T) {
	UserFactory.Reset()
	r1 := UserFactory.Make()
	r2 := UserFactory.Make()
	if r1["email"] == r2["email"] {
		t.Fatal("consecutive Make() calls should produce unique emails")
	}
}

func TestMakeMany_Count(t *testing.T) {
	rows := UserFactory.MakeMany(5)
	if len(rows) != 5 {
		t.Fatalf("expected 5 rows, got %d", len(rows))
	}
}

func TestWith_Override(t *testing.T) {
	admin := UserFactory.With(map[string]any{"role": "admin"})
	row := admin.Make()
	if row["role"] != "admin" {
		t.Fatalf("expected role=admin, got %v", row["role"])
	}
}

func TestWith_DoesNotMutateOriginal(t *testing.T) {
	_ = UserFactory.With(map[string]any{"role": "admin"})
	row := UserFactory.Make()
	if row["role"] != "member" {
		t.Fatal("With() should not mutate the original factory")
	}
}

func TestSeq_Format(t *testing.T) {
	f := New(func(c *Context) map[string]any {
		return map[string]any{"label": c.Seq("item-%d")}
	})
	f.Reset()
	r1 := f.Make()
	r2 := f.Make()
	if r1["label"] == r2["label"] {
		t.Fatal("Seq should increment")
	}
}

func TestPick(t *testing.T) {
	f := New(func(c *Context) map[string]any {
		return map[string]any{"color": c.Pick("red", "green", "blue")}
	})
	row := f.Make()
	color := row["color"].(string)
	if color != "red" && color != "green" && color != "blue" {
		t.Fatalf("unexpected color: %s", color)
	}
}

func TestInt_InRange(t *testing.T) {
	f := New(func(c *Context) map[string]any {
		return map[string]any{"age": c.Int(18, 65)}
	})
	for i := 0; i < 20; i++ {
		row := f.Make()
		age := row["age"].(int)
		if age < 18 || age > 65 {
			t.Fatalf("age %d out of range [18,65]", age)
		}
	}
}

func TestSentence(t *testing.T) {
	f := New(func(c *Context) map[string]any {
		return map[string]any{"title": c.Sentence(5)}
	})
	row := f.Make()
	title := row["title"].(string)
	if len(title) == 0 {
		t.Fatal("sentence should not be empty")
	}
}

func TestReset_ResetsSeq(t *testing.T) {
	f := New(func(c *Context) map[string]any {
		return map[string]any{"email": c.Email()}
	})
	f.Reset()
	r1 := f.Make()
	f.Reset()
	r2 := f.Make()
	if r1["email"] != r2["email"] {
		t.Fatal("Reset() should reset sequence so first items are identical")
	}
}
