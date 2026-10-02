package jobs

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestReadPageLive(t *testing.T) {
	if os.Getenv("LIVE") == "" {
		t.Skip("LIVE=1 to run")
	}
	for _, u := range []string{
		"https://www.ktn.gv.at/Service/Stellenausschreibungen/Details?id=224",
		"https://reliefweb.int/job/4225389/country-director-guinea-liberia",
	} {
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
		text, src := readPage(ctx, u, 6000)
		cancel()
		t.Logf("%s → %s\n%s", u, src, text)
		os.WriteFile("/tmp/rp-"+src+".txt", []byte(text), 0644)
		if text == "" {
			t.Errorf("unreadable: %s", u)
		}
	}
}
