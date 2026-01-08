//
//  main.m
//  focusdetector
//
//  Created by David Yoon on 1/8/26.
//

#import <Foundation/Foundation.h>

int FrontmostAppName(char* buf, int bufLen) {
    if (!buf || bufLen <= 0) {
        return -1;
    }
    
    @autoreleasepool {
        NSString *a = @"Hello from Objective-C";
        const char *utf8 = [a UTF8String];
        
        if (!utf8) {
            return -2;
        }
        
        strncpy(buf, utf8, bufLen - 1);
        buf[bufLen - 1] = '\0';
    }
    
    return 0;
}
