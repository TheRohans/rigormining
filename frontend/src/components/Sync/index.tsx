import React, { useEffect, useState } from 'react';
import Navigation from '@components/App/Navigation';
import TagTree, { TagSelection, UNTAGGED } from '@components/App/TagTree';
import { api, LibraryItem } from '../../api/client';

const DB_NAME = 'rigormining-sync';
const STORE_NAME = 'handles';
const HANDLE_KEY = 'kobo-directory';

function openHandleStore(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, 1);
    req.onupgradeneeded = () => req.result.createObjectStore(STORE_NAME);
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error);
  });
}

async function idbGet(key: string): Promise<FileSystemDirectoryHandle | undefined> {
  const db = await openHandleStore();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, 'readonly');
    const req = tx.objectStore(STORE_NAME).get(key);
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error);
  });
}

async function idbSet(key: string, value: FileSystemDirectoryHandle): Promise<void> {
  const db = await openHandleStore();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, 'readwrite');
    tx.objectStore(STORE_NAME).put(value, key);
    tx.oncomplete = () => resolve();
    tx.onerror = () => reject(tx.error);
  });
}

const sanitize = (name: string) => name.replace(/[\\/:]/g, '-');

// Checked means "the server-side sync_state wants this on the device" -
// request_sync (queued, not copied yet) or synced (already there). Not
// checked covers both "never requested" and request_remove (unticked a
// synced item - it's still physically on the device until the next sync
// run actually deletes it, but the target state is "don't want it").
const isWanted = (item: LibraryItem) => item.sync_state === 'request_sync' || item.sync_state === 'synced';

const STATE_BADGE: Record<string, { label: string; className: string }> = {
  synced: { label: 'on device', className: 'bg-green-100 text-green-800' },
  request_sync: { label: 'queued to sync', className: 'bg-indigo-100 text-indigo-800' },
  request_remove: { label: 'queued to remove', className: 'bg-amber-100 text-amber-800' },
};
const NOT_QUEUED_BADGE = { label: 'not queued', className: 'bg-gray-100 text-gray-600' };

export const Sync: React.FC = () => {
  const [items, setItems] = useState<LibraryItem[]>([]);
  const [status, setStatus] = useState('');
  const [selectedTag, setSelectedTag] = useState<TagSelection>(null);
  const supported = typeof window.showDirectoryPicker === 'function';

  // Every item with a file, whatever its sync_state - so you can see
  // what's already on the device (and untick it) as well as what's queued,
  // not just what's still pending.
  const loadAll = () => {
    api
      .listItems()
      .then((all) => setItems(all.filter((i) => i.file_type)))
      .catch((e) => setStatus(e.message));
  };

  useEffect(() => {
    loadAll();
  }, []);

  const toggleWanted = async (item: LibraryItem) => {
    try {
      const updated = isWanted(item) ? await api.cancelSync(item.id) : await api.requestSync(item.id);
      setItems((prev) => prev.map((i) => (i.id === updated.id ? updated : i)));
    } catch (e) {
      setStatus((e as Error).message);
    }
  };

  const getDirectoryHandle = async (): Promise<FileSystemDirectoryHandle | null> => {
    if (!window.showDirectoryPicker) return null;

    const saved = await idbGet(HANDLE_KEY).catch(() => undefined);
    if (saved) {
      const granted =
        (await saved.queryPermission({ mode: 'readwrite' })) === 'granted' ||
        (await saved.requestPermission({ mode: 'readwrite' })) === 'granted';
      if (granted) return saved;
    }

    const handle = await window.showDirectoryPicker({ mode: 'readwrite' });
    await idbSet(HANDLE_KEY, handle);
    return handle;
  };

  // sync_state="synced" only ever records "we copied this file at some
  // point" - there's no ongoing link to what's actually still on the
  // device, so if the user deletes a file directly on the Kobo, the app
  // has no way to know unless it actually looks. This checks every item we
  // think is synced against the real directory contents and acks any
  // whose file is genuinely gone as "removed", so it goes back to
  // not-queued instead of the app trusting a stale state forever.
  const reconcileSynced = async (koboDir: FileSystemDirectoryHandle): Promise<number> => {
    const synced = await api.listItems({ sync_state: 'synced' }).then((all) => all.filter((i) => i.file_type));
    let cleared = 0;
    for (const item of synced) {
      const filename = sanitize(`${item.title}.${item.file_type}`);
      try {
        await koboDir.getFileHandle(filename);
      } catch (e) {
        await api.ackSync(item.id, 'removed');
        cleared += 1;
      }
    }
    return cleared;
  };

  const syncToKobo = async () => {
    setStatus('Requesting access to your Kobo drive...');
    let dir: FileSystemDirectoryHandle | null;
    try {
      dir = await getDirectoryHandle();
    } catch (e) {
      setStatus(`Could not access the drive: ${(e as Error).message}`);
      return;
    }
    if (!dir) {
      setStatus('This browser does not support the File System Access API. Use Chrome or Edge.');
      return;
    }

    const koboDir = await dir.getDirectoryHandle('rigormining', { create: true });

    setStatus("Checking what's actually still on the device...");
    const reconciled = await reconcileSynced(koboDir);
    if (reconciled > 0) {
      setStatus(`${reconciled} item${reconciled === 1 ? '' : 's'} no longer on the device - un-queued.`);
    }

    const all = await api.listItems().then((res) => res.filter((i) => i.file_type));
    const toSync = all.filter((i) => i.sync_state === 'request_sync');
    const toRemove = all.filter((i) => i.sync_state === 'request_remove');
    setItems(all);

    if (toSync.length === 0 && toRemove.length === 0) {
      setStatus('Nothing queued to sync or remove.');
      return;
    }

    for (const item of toRemove) {
      setStatus(`Removing "${item.title}"...`);
      try {
        const filename = sanitize(`${item.title}.${item.file_type}`);
        await koboDir.removeEntry(filename).catch(() => undefined); // already gone is fine
        await api.ackSync(item.id, 'removed');
      } catch (e) {
        setStatus(`Failed removing "${item.title}": ${(e as Error).message}`);
        return;
      }
    }

    for (const item of toSync) {
      setStatus(`Copying "${item.title}"...`);
      try {
        const blob = await api.fetchFileBlob(item.id);
        const filename = sanitize(`${item.title}.${item.file_type}`);
        const fileHandle = await koboDir.getFileHandle(filename, { create: true });
        const writable = await fileHandle.createWritable();
        await writable.write(blob);
        await writable.close();
        await api.ackSync(item.id, 'synced');
      } catch (e) {
        setStatus(`Failed on "${item.title}": ${(e as Error).message}`);
        return;
      }
    }

    setStatus(
      `Done - copied ${toSync.length} item${toSync.length === 1 ? '' : 's'} and removed ${toRemove.length} item${
        toRemove.length === 1 ? '' : 's'
      }.`,
    );
    loadAll();
  };

  const visibleItems = items.filter((item) => {
    if (selectedTag === null) return true;
    if (selectedTag === UNTAGGED) return item.tags.length === 0;
    return item.tags.some((t) => t.name === selectedTag);
  });

  const queuedSync = items.filter((i) => i.sync_state === 'request_sync').length;
  const queuedRemove = items.filter((i) => i.sync_state === 'request_remove').length;
  const onDevice = items.filter((i) => i.sync_state === 'synced').length;

  return (
    <>
      <Navigation loggedIn={true} />
      <div className="flex" style={{ height: 'calc(100vh - 4rem)' }}>
        <TagTree items={items} selected={selectedTag} onSelect={setSelectedTag} />

        <div className="flex-1 overflow-y-auto">
          <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            <h1 className="text-lg font-semibold text-gray-900 mb-4">Sync to Kobo</h1>

            {!supported && (
              <p className="text-sm text-red-700 mb-4">
                This browser doesn&apos;t support one-click sync (Chrome or Edge required). Download items individually
                from their detail page and copy them onto your Kobo over USB instead.
              </p>
            )}

            <p className="text-sm text-gray-600 mb-4">
              {onDevice} on device, {queuedSync} queued to sync, {queuedRemove} queued to remove.
            </p>

            {supported && (
              <button
                onClick={syncToKobo}
                disabled={queuedSync === 0 && queuedRemove === 0}
                className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50"
              >
                Plug in your Kobo, then click to sync
              </button>
            )}

            {status && <p className="mt-4 text-sm text-gray-700">{status}</p>}

            <ul className="mt-6 divide-y divide-gray-100 border border-gray-200 rounded-md bg-white">
              {visibleItems.map((item) => {
                const badge = (item.sync_state && STATE_BADGE[item.sync_state]) || NOT_QUEUED_BADGE;
                return (
                  <li key={item.id} className="px-3 py-2 flex justify-between items-center gap-3">
                    <label className="flex items-center gap-3 flex-1 min-w-0 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={isWanted(item)}
                        onChange={() => toggleWanted(item)}
                        className="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                      />
                      <span className="truncate text-sm text-gray-900">{item.title}</span>
                    </label>
                    <span className={`text-xs px-2 py-0.5 rounded-full flex-shrink-0 ${badge.className}`}>
                      {badge.label}
                    </span>
                  </li>
                );
              })}
              {visibleItems.length === 0 && (
                <li className="px-3 py-6 text-center text-sm text-gray-500">No items here.</li>
              )}
            </ul>
          </div>
        </div>
      </div>
    </>
  );
};

export default Sync;
