//go:build darwin && cgo

package idle

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation
#include <IOKit/IOKitLib.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdint.h>

uint64_t hidIdleTimeNanos() {
    io_iterator_t iter = 0;
    io_registry_entry_t entry = 0;
    kern_return_t kr = IOServiceGetMatchingServices(kIOMainPortDefault,
        IOServiceMatching("IOHIDSystem"), &iter);
    if (kr != KERN_SUCCESS || iter == 0) return 0;

    entry = IOIteratorNext(iter);
    IOObjectRelease(iter);
    if (entry == 0) return 0;

    CFTypeRef prop = IORegistryEntryCreateCFProperty(entry, CFSTR("HIDIdleTime"), kCFAllocatorDefault, 0);
    IOObjectRelease(entry);
    if (prop == NULL) return 0;

    uint64_t nanos = 0;
    if (CFGetTypeID(prop) == CFDataGetTypeID()) {
        CFDataRef data = (CFDataRef)prop;
        if (CFDataGetLength(data) == (CFIndex)sizeof(uint64_t)) {
            CFDataGetBytes(data, CFRangeMake(0, sizeof(uint64_t)), (UInt8*)&nanos);
        }
    } else if (CFGetTypeID(prop) == CFNumberGetTypeID()) {
        CFNumberGetValue((CFNumberRef)prop, kCFNumberSInt64Type, &nanos);
    }
    CFRelease(prop);
    return nanos;
}
*/
import "C"

import "time"

func IdleDuration() time.Duration {
	nanos := uint64(C.hidIdleTimeNanos())
	return time.Duration(nanos) * time.Nanosecond
}

func IsIdle(threshold time.Duration) bool {
	return IdleDuration() >= threshold
}
