//go:build darwin && !ios

// The window's menu bar on macOS: cairn's menus added to the one Gio makes,
// each line calling back into Go with its number. See menu_darwin.go.

#import <AppKit/AppKit.h>

#include "_cgo_export.h"

@interface CairnMenuTarget : NSObject
- (void)picked:(NSMenuItem *)item;
@end

@implementation CairnMenuTarget
- (void)picked:(NSMenuItem *)item {
	cairnMenuPicked((int)[item tag]);
}
@end

static CairnMenuTarget *cairnTarget;
static NSMenu *cairnLast;

// cairn_menu_group adds a menu to the menu bar; the lines added after it go in
// it. AppKit is only ever touched on the main thread.
void cairn_menu_group(const char *title) {
	NSString *named = [NSString stringWithUTF8String:title];
	dispatch_async(dispatch_get_main_queue(), ^{
		if (cairnTarget == nil) {
			cairnTarget = [CairnMenuTarget new];
		}
		NSMenuItem *top = [[NSMenuItem alloc] initWithTitle:named action:nil keyEquivalent:@""];
		NSMenu *menu = [[NSMenu alloc] initWithTitle:named];
		[top setSubmenu:menu];
		[[NSApp mainMenu] addItem:top];
		cairnLast = menu;
	});
}

// cairn_menu_item adds a line to the last menu: a title, the key it answers
// with ⌘ (none when empty), and the number it calls back with.
void cairn_menu_item(const char *title, const char *key, int tag) {
	NSString *named = [NSString stringWithUTF8String:title];
	NSString *keyed = [NSString stringWithUTF8String:key];
	dispatch_async(dispatch_get_main_queue(), ^{
		NSMenuItem *line = [[NSMenuItem alloc] initWithTitle:named action:@selector(picked:) keyEquivalent:keyed];
		[line setTarget:cairnTarget];
		[line setTag:tag];
		[cairnLast addItem:line];
	});
}
