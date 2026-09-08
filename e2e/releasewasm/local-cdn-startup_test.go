//go:build !skip_e2e && !js

package releasewasm

import (
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"

	playwright "github.com/mxschmitt/playwright-go"
	"github.com/s4wave/spacewave/core/cdn"
)

// TestLocalCDNDriveStartup measures the production landing-to-Drive operation
// over localhost distribution metadata and KVF packs in a fresh browser context.
func TestLocalCDNDriveStartup(t *testing.T) {
	if os.Getenv(localCDNEnv) != "1" {
		t.Skip("set " + localCDNEnv + "=1 to build and serve the local startup CDN")
	}

	// Observe delivery without request interception, which disables HTTP caching.
	page := testHarness.newPage(t)
	var mtx sync.Mutex
	var external []string
	var distribution, root, ranges int
	page.Context().OnRequest(func(req playwright.Request) {
		// The optional public-content catalog is absent from this fixture.
		// Browser proxy isolation rejects its connection before any remote I/O.
		if req.URL() == cdn.DefaultBaseURL+"/"+cdn.ProvisionedSpaceID+"/root.packedmsg" {
			return
		}
		u, err := url.Parse(req.URL())
		if err != nil || u.Hostname() != "127.0.0.1" {
			mtx.Lock()
			external = append(external, req.URL())
			mtx.Unlock()
			return
		}
		mtx.Lock()
		switch {
		case strings.HasSuffix(u.Path, "/distribution.packedmsg"):
			distribution++
		case strings.HasSuffix(u.Path, "/root.packedmsg"):
			root++
		case strings.HasSuffix(u.Path, ".kvf"):
			ranges++
		}
		mtx.Unlock()
	})

	// Start exactly where a new visitor starts, before the runtime is booted.
	if _, err := page.Goto(testHarness.baseURL + "/"); err != nil {
		t.Fatal(err)
	}
	clickMS := browserNowMs(t, page)
	if err := page.Locator("a").Filter(playwright.LocatorFilterOptions{HasText: "Create a Drive"}).First().Click(); err != nil {
		t.Fatal(err)
	}
	waitForLiveApp(t, page)
	waitForQuickstartAppRoute(t, page)
	readyMS, failure := waitForQuickstartDriveContentReady(t, page)
	if failure != "" {
		t.Fatal(failure)
	}
	completeQuickstartDriveIntroIfPresent(t, page)
	if _, failure := exerciseQuickstartDriveGoldenPath(t, page); failure != "" {
		t.Fatal(failure)
	}
	logQuickstartTiming(t, page)

	// A successful timing must include the remote-loading contract it measures.
	mtx.Lock()
	defer mtx.Unlock()
	t.Logf("local CDN startup: navigation=%dms click-to-file=%dms distribution=%d roots=%d pack-requests=%d", *readyMS, *readyMS-clickMS, distribution, root, ranges)
	if len(external) != 0 {
		t.Errorf("startup attempted external requests: %v", external)
	}
	if distribution == 0 || root == 0 || ranges == 0 {
		t.Error("startup did not exercise distribution, CDN root, and pack loading")
	}
}
