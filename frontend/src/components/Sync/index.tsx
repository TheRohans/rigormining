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

export const Sync: React.FC = () => {
  const [items, setItems] = useState<LibraryItem[]>([]);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [status, setStatus] = useState('');
  const [selectedTag, setSelectedTag] = useState<TagSelection>(null);
  const supported = typeof window.showDirectoryPicker === 'function';

  // Every item with a file, delivered or not - not just pending ones, so
  // you can see what's already on the device (and reset it) as well as
  // pick individual items to sync rather than always doing the whole
  // library at once.
  const loadAll = () => {
    api
      .listItems()
      .then((all) => {
        const withFiles = all.filter((i) => i.file_type);
        setItems(withFiles);
        setSelected(new Set(withFiles.filter((i) => !i.delivered_at).map((i) => i.id)));
      })
      .catch((e) => setStatus(e.message));
  };

  useEffect(() => {
    loadAll();
  }, []);

  const toggleSelected = (id: string) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const resetItem = async (id: string) => {
    await api.clearDelivered(id);
    loadAll();
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

  // delivered_at only ever records "we copied this file at some point" -
  // there's no ongoing link to what's actually still on the device, so if
  // the user deletes a file directly on the Kobo, the app has no way to
  // know unless it actually looks. This checks every item we think is
  // delivered against the real directory contents and un-marks any whose
  // file is genuinely gone, so it becomes pending again instead of the app
  // trusting a stale flag forever.
  const reconcileDelivered = async (koboDir: FileSystemDirectoryHandle): Promise<number> => {
    const delivered = await api.listItems({ delivered: true }).then((all) => all.filter((i) => i.file_type));
    let cleared = 0;
    for (const item of delivered) {
      const filename = sanitize(`${item.title}.${item.file_type}`);
      try {
        await koboDir.getFileHandle(filename);
      } catch (e) {
        await api.clearDelivered(item.id);
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
    const reconciled = await reconcileDelivered(koboDir);
    if (reconciled > 0) {
      setStatus(`${reconciled} item${reconciled === 1 ? '' : 's'} no longer on the device - re-adding to sync.`);
    }

    const all = await api.listItems().then((res) => res.filter((i) => i.file_type));
    const toSync = all.filter((i) => selected.has(i.id));
    setItems(all);

    if (toSync.length === 0) {
      setStatus('Nothing selected to sync.');
      return;
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
        await api.markDelivered(item.id);
      } catch (e) {
        setStatus(`Failed on "${item.title}": ${(e as Error).message}`);
        return;
      }
    }

    setStatus(`Done - copied ${toSync.length} item${toSync.length === 1 ? '' : 's'} to the Kobo.`);
    loadAll();
  };

  const visibleItems = items.filter((item) => {
    if (selectedTag === null) return true;
    if (selectedTag === UNTAGGED) return item.tags.length === 0;
    return item.tags.some((t) => t.name === selectedTag);
  });

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
                This browser doesn't support one-click sync (Chrome or Edge required). Download items individually from
                their detail page and copy them onto your Kobo over USB instead.
              </p>
            )}

            <p className="text-sm text-gray-600 mb-4">
              {selected.size} of {items.length} item{items.length === 1 ? '' : 's'} selected to sync.
            </p>

            {supported && (
              <button
                onClick={syncToKobo}
                disabled={selected.size === 0}
                className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50"
              >
                Plug in your Kobo, then click to sync selected
              </button>
            )}

            {status && <p className="mt-4 text-sm text-gray-700">{status}</p>}

            <ul className="mt-6 divide-y divide-gray-100 border border-gray-200 rounded-md bg-white">
              {visibleItems.map((item) => (
                <li key={item.id} className="px-3 py-2 flex justify-between items-center gap-3">
                  <label className="flex items-center gap-3 flex-1 min-w-0 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={selected.has(item.id)}
                      onChange={() => toggleSelected(item.id)}
                      className="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                    />
                    <span className="truncate text-sm text-gray-900">{item.title}</span>
                  </label>
                  <span
                    className={`text-xs px-2 py-0.5 rounded-full flex-shrink-0 ${
                      item.delivered_at ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-600'
                    }`}
                  >
                    {item.delivered_at ? 'on device' : 'pending'}
                  </span>
                  {item.delivered_at && (
                    <button
                      onClick={() => resetItem(item.id)}
                      className="text-xs text-red-600 hover:text-red-900 flex-shrink-0"
                    >
                      Reset
                    </button>
                  )}
                </li>
              ))}
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
