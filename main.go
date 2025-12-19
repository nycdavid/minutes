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

        // layer 0 => normal app windows (filters out overlays like BetterTouchTool)
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

char* FrontmostAppAndAXTitle() {
    @autoreleasepool {
        pid_t pid = FrontmostNormalWindowPID();
        if (pid == 0) return NULL;

        NSRunningApplication *app = [NSRunningApplication runningApplicationWithProcessIdentifier:pid];
        NSString *appName = app ? [app localizedName] : @"(unknown app)";

        AXUIElementRef appRef = AXUIElementCreateApplication(pid);
        if (!appRef) return strdup([[appName description] UTF8String]);

        CFTypeRef winValue = NULL;
        AXError winErr = AXUIElementCopyAttributeValue(appRef, kAXFocusedWindowAttribute, &winValue);

        NSString *out = appName;

        if (winErr == kAXErrorSuccess && winValue) {
            AXUIElementRef winRef = (AXUIElementRef)winValue;

            CFTypeRef titleValue = NULL;
            AXError titleErr = AXUIElementCopyAttributeValue(winRef, kAXTitleAttribute, &titleValue);

            if (titleErr == kAXErrorSuccess && titleValue && CFGetTypeID(titleValue) == CFStringGetTypeID()) {
                NSString *title = (__bridge NSString*)titleValue;
                out = [NSString stringWithFormat:@"%@ — %@", appName, title];
                CFRelease(titleValue);
            }

            CFRelease(winRef);
        }

        CFRelease(appRef);
        return strdup([out UTF8String]);
    }
}
*/
import "C"

import (
	"fmt"
	"time"
	"unsafe"
)

func frontmostAppAndTitle() string {
	s := C.FrontmostAppAndAXTitle()
	if s == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(s))
	return C.GoString(s)
}

func main() {
	for {
		fmt.Println(frontmostAppAndTitle())
		time.Sleep(1 * time.Second)
	}
}
