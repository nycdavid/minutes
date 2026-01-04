//go:build darwin && cgo

package main

/*
#cgo darwin CFLAGS: -x objective-c
#cgo darwin LDFLAGS: -framework AppKit -framework ApplicationServices -framework CoreFoundation
#import <AppKit/AppKit.h>
#import <ApplicationServices/ApplicationServices.h>
#import <CoreFoundation/CoreFoundation.h>
#import <stdlib.h>

static pid_t FrontmostNormalWindowPID() {
    CFArrayRef list = CGWindowListCopyWindowInfo(kCGWindowListOptionOnScreenOnly, kCGNullWindowID);
    if (!list) return 0;

    CFIndex count = CFArrayGetCount(list);
    for (CFIndex i = 0; i < count; i++) {
        CFDictionaryRef w = CFArrayGetValueAtIndex(list, i);
        if (!w) continue;

        CFNumberRef layerRef = CFDictionaryGetValue(w, kCGWindowLayer);
        int layer = -1;
        if (layerRef) CFNumberGetValue(layerRef, kCFNumberIntType, &layer);
        if (layer != 0) continue;

        CFNumberRef pidRef = CFDictionaryGetValue(w, kCGWindowOwnerPID);
        int pid = 0;
        if (pidRef && CFNumberGetValue(pidRef, kCFNumberIntType, &pid)) {
            CFRelease(list);
            return (pid_t)pid;
        }
    }

    CFRelease(list);
    return 0;
}

// Returns "AppName — Title" (title may be omitted if not available).
char* FrontmostAppAndAXTitle() {
    @autoreleasepool {
        // 1) Frontmost app WITHOUT screen-recording / window-list APIs
        NSRunningApplication *front = [[NSWorkspace sharedWorkspace] frontmostApplication];
        if (!front) return NULL;

        pid_t pid = front.processIdentifier;
        NSString *appName = front.localizedName ?: @"(unknown app)";

        // 2) AX for the focused window title (requires Accessibility permission)
        AXUIElementRef appRef = AXUIElementCreateApplication(pid);
        if (!appRef) return strdup([appName UTF8String]);

        CFTypeRef winValue = NULL;
        AXError winErr = AXUIElementCopyAttributeValue(appRef, kAXFocusedWindowAttribute, &winValue);

        NSString *out = appName;

        if (winErr == kAXErrorSuccess && winValue) {
            AXUIElementRef winRef = (AXUIElementRef)winValue;

            CFTypeRef titleValue = NULL;
            AXError titleErr = AXUIElementCopyAttributeValue(winRef, kAXTitleAttribute, &titleValue);

            if (titleErr == kAXErrorSuccess && titleValue &&
                CFGetTypeID(titleValue) == CFStringGetTypeID()) {

                NSString *title = (__bridge NSString*)titleValue;
                if (title.length > 0) {
                    out = [NSString stringWithFormat:@"%@ — %@", appName, title];
                }
                CFRelease(titleValue);
            }

            CFRelease(winRef); // releases winValue
        }

        CFRelease(appRef);
        return strdup([out UTF8String]);
    }
}
*/
import "C"

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
	"unsafe"

	"github.com/nycdavid/minutes/internal/idle"
)

func frontmostAppAndTitle() string {
	s := C.FrontmostAppAndAXTitle()
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

type (
	session struct {
		app      string
		duration time.Duration
		metadata map[string]string
	}
)

func toSession(line string) *session {
	fmt.Println(line)
	return &session{}
}

func (s *session) durationString() string {
	return fmt.Sprintf("%vh%vm%vs", s.duration.Hours(), s.duration.Minutes(), s.duration.Seconds())
}

var (
	idleThreshold = 30 * time.Second
	loopInterval  = 500 * time.Millisecond
)

func main() {
	var lastPrinted string
	var lastChromeURL string
	var lastChromeFetch time.Time

	for {
		if idle.IsIdle(idleThreshold) {
			fmt.Println("Idle...")
			time.Sleep(loopInterval)
			continue
		}

		info := frontmostAppAndTitle()
		if info == "" {
			time.Sleep(loopInterval)
			continue
		}

		// Only print on change (reduces spam).
		if info != lastPrinted {
			lastPrinted = info
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
							fmt.Println(msg)
						}
					} else if url != "" && url != lastChromeURL {
						lastChromeURL = url
						info = fmt.Sprintf("%s - URL: %s", info, url)
					}
				}
			} else {
				// Reset chrome URL state when leaving Chrome.
				fmt.Println(info)
				lastChromeURL = ""
			}

			s := toSession(info)
			fmt.Println(fmt.Sprintf("[app]: %s, [duration]: %s", s.app, s.duration))
		}

		time.Sleep(loopInterval)
	}
}
