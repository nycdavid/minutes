//go:build darwin && cgo

package main

/*
#cgo darwin CFLAGS: -x objective-c
#cgo darwin LDFLAGS: -framework AppKit -framework ApplicationServices -framework CoreFoundation
#import <AppKit/AppKit.h>
#import <ApplicationServices/ApplicationServices.h>
#import <CoreFoundation/CoreFoundation.h>
#import <stdlib.h>

static NSString *AXCopyStringAttr(AXUIElementRef ref, CFStringRef attr) {
   CFTypeRef v = NULL;
   AXError e = AXUIElementCopyAttributeValue(ref, attr, &v);
   if (e != kAXErrorSuccess || !v) return nil;

   NSString *s = nil;
   if (CFGetTypeID(v) == CFStringGetTypeID()) {
       s = (__bridge NSString *)v;
   }
   CFRelease(v);
   return s;
}

char* FrontmostAppAndHeaderTitle(void) {
   @autoreleasepool {
       // If AX isn't trusted, you will not get stable focused element/window/title.
       if (!AXIsProcessTrusted()) {
           NSRunningApplication *front = [[NSWorkspace sharedWorkspace] frontmostApplication];
           if (!front) return NULL;
           NSString *appName = front.localizedName ?: @"(unknown app)";
           return strdup(appName.UTF8String);
       }

       AXUIElementRef sys = AXUIElementCreateSystemWide();
       if (!sys) return NULL;

       // 1) Focused UI element (more reliable than focused window)
       AXUIElementRef focusedElem = NULL;
       AXError fe = AXUIElementCopyAttributeValue(
           sys,
           kAXFocusedUIElementAttribute,
           (CFTypeRef *)&focusedElem
       );
       CFRelease(sys);

       if (fe != kAXErrorSuccess || !focusedElem) return NULL;

       // 2) Get its window
       AXUIElementRef winRef = NULL;
       AXError we = AXUIElementCopyAttributeValue(
           focusedElem,
           kAXWindowAttribute,
           (CFTypeRef *)&winRef
       );
       CFRelease(focusedElem);

       // If we can't get a window, fall back to frontmost app name
       if (we != kAXErrorSuccess || !winRef) {
           NSRunningApplication *front = [[NSWorkspace sharedWorkspace] frontmostApplication];
           if (!front) return NULL;
           NSString *appName = front.localizedName ?: @"(unknown app)";
           return strdup(appName.UTF8String);
       }

       // 3) Derive app name from the window's PID (keeps app+title consistent)
       pid_t pid = 0;
       AXUIElementGetPid(winRef, &pid);
       NSRunningApplication *app = [NSRunningApplication runningApplicationWithProcessIdentifier:pid];
       NSString *appName = app.localizedName ?: @"(unknown app)";

       // 4) Title, with fallback to Document (often better for editors/browsers)
       NSString *title = AXCopyStringAttr(winRef, kAXTitleAttribute);
       if (title.length == 0) {
           // Some apps populate Document when Title is empty/stale.
           title = AXCopyStringAttr(winRef, kAXDocumentAttribute);
       }

       NSString *out = appName;
       if (title.length > 0) {
           out = [NSString stringWithFormat:@"%@ — %@", appName, title];
       }

       CFRelease(winRef);
       return strdup(out.UTF8String);
   }
}
*/
import "C"

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
	"unsafe"

	"github.com/nycdavid/minutes/internal/idle"
	"github.com/nycdavid/minutes/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func frontmostAppAndTitle() string {
	s := C.FrontmostAppAndHeaderTitle()
	if s == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(s))
	return C.GoString(s)
}

// Returns active tab URL from Chrome using osascript.
// Requires Automation permission: your app controlling "Google Chrome".
func chromeActiveURL() (string, error) {
	// Keep it single-line output.
	script := `tell application "Google Chrome" to get URL of active tab of front window`
	cmd := exec.Command("/usr/bin/osascript", "-e", script)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// If permission is missing, stderr usually contains a helpful message.
		return "", fmt.Errorf("osascript failed: %v: %s", err, strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(out.String()), nil
}

func toSession(line string) *models.Heartbeat {
	parts := strings.Split(line, " — ")
	app := parts[0]
	return &models.Heartbeat{Application: app, Timestamp: time.Now().UTC().UnixMilli(), Metadata: parts[1]}
}

var (
	idleThreshold = 30 * time.Second
	loopInterval  = 30 * time.Second
)

func main() {
	var lastChromeURL string
	var lastChromeFetch time.Time

	for {
		if idle.IsIdle(idleThreshold) {
			slog.Info("Idle...")
			time.Sleep(loopInterval)
			continue
		}

		info := frontmostAppAndTitle()
		if info == "" {
			time.Sleep(loopInterval)
			continue
		}

		// If Chrome is focused, fetch URL (throttled).
		// "Google Chrome" is the localized app name; adjust if you use Chrome Beta/Canary.
		if strings.HasPrefix(info, "Google Chrome") {
			if time.Since(lastChromeFetch) >= 2*time.Second {
				lastChromeFetch = time.Now()

				url, err := chromeActiveURL()
				if err != nil {
					// Print once if it changes (or on first failure).
					msg := "Chrome URL error: " + err.Error()
					if msg != lastChromeURL {
						lastChromeURL = msg
						slog.Error("getting Chrome URL", "msg", msg)
					}
				} else if url != "" && url != lastChromeURL {
					lastChromeURL = url
					info = fmt.Sprintf("%s - URL: %s", info, url)
				}
			}
		} else {
			// Reset chrome URL state when leaving Chrome.
			lastChromeURL = ""
		}

		sqliteFpath := os.Getenv("DB_URL")
		d, err := gorm.Open(sqlite.Open(sqliteFpath), &gorm.Config{})
		if err != nil {
			log.Fatal(err)
		}

		s := toSession(info)
		err = gorm.G[models.Heartbeat](d).Create(context.Background(), s)
		if err != nil {
			log.Fatal(err)
		}

		slog.Info(
			"heartbeat sent",
			"app", s.Application,
			"timestamp", s.Timestamp,
			"metadata", s.Metadata,
		)

		time.Sleep(loopInterval)
	}
}
