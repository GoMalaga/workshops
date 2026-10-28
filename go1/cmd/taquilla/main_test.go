package main

import (
	"bytes"
	"strings"
	"testing"

	"reservations"
)

// A scripted session is the test of a console program: feed the input,
// compare the output line by line.
func TestRun_scriptedSession(t *testing.T) {
	store := reservations.NewStore()
	store.AddProduct(&reservations.Product{ID: "p1", Name: "Ticket", Stock: 10})

	in := strings.NewReader(strings.Join([]string{
		"reserve p1 3",
		"stock p1",
		"confirm r1",
		"cancel r1",
		"reserve p1 8",
		"reserve p1 7",
		"cancel r2",
		"stock p1",
		"reserve p1 -3",
		"reserve p1 many",
		"stock p9",
		"confirm",
		"",
		"dance",
		"quit",
		"stock p1",
	}, "\n"))
	want := strings.Join([]string{
		"r1 Held",
		"7",
		"ok",
		"error: illegal state transition",
		"error: out of stock",
		"r2 Held",
		"ok",
		"7",
		"error: invalid quantity",
		`error: "many" is not a number`,
		"error: product not found",
		"error: usage: confirm <reservation>",
		`error: unknown command "dance"`,
	}, "\n") + "\n"

	var out bytes.Buffer
	if err := run(in, &out, store); err != nil {
		t.Fatalf("run returned an error: %v", err)
	}
	if got := out.String(); got != want {
		t.Errorf("session output mismatch\n--- want ---\n%s--- got ---\n%s", want, got)
	}
}
