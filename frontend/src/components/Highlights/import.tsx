// Deprecated: this screen expected a native `window.ks_getHighlights`/
// `window.ks_deviceCheck` bridge that was never implemented. Kobo delivery
// is now handled by @components/Sync using the File System Access API;
// syncing highlights back from the device is deferred (see the project
// plan). Kept only because this session couldn't get permission to delete
// the file outright - safe to remove, along with its route in App/index.tsx
// (already removed) and Highlights/index.tsx.
export {};
