package htmlparse

import (
	"strings"
	"testing"
)

func TestRemoveByTag(t *testing.T) {
	d, _ := Parse(`<html><body><p>keep</p><script>bad</script><script>bad2</script></body></html>`)
	n := RemoveByTag(d, "script")
	if n != 2 {
		t.Fatalf("removed = %d, want 2", n)
	}
	out, _ := d.Render()
	if strings.Contains(out, "script") {
		t.Fatalf("script survived: %q", out)
	}
}

func TestRemoveByID(t *testing.T) {
	d, _ := Parse(`<html><body><div id="a">x</div><div id="b">y</div></body></html>`)
	ok := RemoveByID(d, "a")
	if !ok {
		t.Fatal("expected removal")
	}
	out, _ := d.Render()
	if strings.Contains(out, "id=\"a\"") {
		t.Fatalf("div#a survived: %q", out)
	}
}

func TestAddRemoveClass(t *testing.T) {
	d, _ := Parse(`<html><body><p class="a">x</p></body></html>`)
	ps := FindByTag(d, "p")
	if len(ps) == 0 {
		t.Fatal("no p")
	}
	AddClass(ps[0], "b")
	if !HasClass(ps[0], "b") {
		t.Fatal("class b not added")
	}
	RemoveClass(ps[0], "a")
	if HasClass(ps[0], "a") {
		t.Fatal("class a not removed")
	}
}

func TestSetAttr(t *testing.T) {
	d, _ := Parse(`<html><body><p>x</p></body></html>`)
	ps := FindByTag(d, "p")
	SetAttr(ps[0], "id", "test")
	if Attr(ps[0], "id") != "test" {
		t.Fatal("attr not set")
	}
}

func TestInnerText(t *testing.T) {
	d, _ := Parse(`<html><body><p>Hello <b>world</b> foo</p></body></html>`)
	ps := FindByTag(d, "p")
	text := InnerText(ps[0])
	if text != "Hello world foo" {
		t.Fatalf("inner text = %q", text)
	}
}

func TestChildren(t *testing.T) {
	d, _ := Parse(`<html><body><div><p>1</p><p>2</p></div></body></html>`)
	divs := FindByTag(d, "div")
	kids := Children(divs[0])
	if len(kids) != 2 {
		t.Fatalf("children = %d, want 2", len(kids))
	}
}
