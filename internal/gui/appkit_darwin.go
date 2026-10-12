//go:build darwin && !ios

package gui

// What the window says to AppKit: cairn's menus in the system's menu bar and
// the system's file panels. Written here, in a cgo preamble, rather than in
// an Objective-C file of its own: cmd/go links -lobjc once for every package
// holding one, Gio's is already one, and a second has the linker say so on
// every build. The preamble holds the definitions, so this file exports
// nothing: a file that does has its preamble compiled twice. See
// menu_darwin.go and files_darwin.go.

/*
#cgo CFLAGS: -Werror -fobjc-arc -x objective-c
#cgo LDFLAGS: -framework AppKit -framework UniformTypeIdentifiers

#import <AppKit/AppKit.h>
#import <UniformTypeIdentifiers/UniformTypeIdentifiers.h>

extern void cairnMenuPicked(int tag);
extern void cairnFileChosen(char *path);

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

// cairnAnswered hands the panel's file back, its path or nothing when it was
// closed without one.
static void cairnAnswered(NSSavePanel *panel, NSInteger result) {
	if (result == NSModalResponseOK && [panel URL] != nil) {
		cairnFileChosen((char *)[[[panel URL] path] UTF8String]);
	} else {
		cairnFileChosen((char *)"");
	}
}

// cairn_save_panel asks, over the window, where a file named so goes. Called
// on the main thread.
void cairn_save_panel(uintptr_t view, const char *name) {
	NSView *over = (__bridge NSView *)(void *)view;
	NSSavePanel *panel = [NSSavePanel savePanel];
	[panel setNameFieldStringValue:[NSString stringWithUTF8String:name]];
	[panel beginSheetModalForWindow:[over window] completionHandler:^(NSInteger result) {
		cairnAnswered(panel, result);
	}];
}

// cairn_open_panel asks, over the window, for a file of one of the extensions
// given, separated by commas; any file when there are none. Called on the
// main thread.
void cairn_open_panel(uintptr_t view, const char *extensions) {
	NSView *over = (__bridge NSView *)(void *)view;
	NSOpenPanel *panel = [NSOpenPanel openPanel];
	NSString *listed = [NSString stringWithUTF8String:extensions];
	if ([listed length] > 0) {
		NSMutableArray<UTType *> *types = [NSMutableArray new];
		for (NSString *ext in [listed componentsSeparatedByString:@","]) {
			UTType *type = [UTType typeWithFilenameExtension:ext];
			if (type != nil) {
				[types addObject:type];
			}
		}
		[panel setAllowedContentTypes:types];
	}
	[panel beginSheetModalForWindow:[over window] completionHandler:^(NSInteger result) {
		cairnAnswered(panel, result);
	}];
}

// Nothing else of the window is said to AppKit.
*/
import "C"
