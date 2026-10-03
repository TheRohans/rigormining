import React, { KeyboardEvent, useRef, useState } from 'react';
import { LibraryIcon, PencilIcon, TagIcon } from '@heroicons/react/solid';
import { LibraryItem } from '../../api/client';

// Sentinel for the "Untagged" pseudo-tag - distinct from `null` (which means
// "All items"), and not a value any real tag name could collide with.
export const UNTAGGED = '\0untagged';

export type TagSelection = string | typeof UNTAGGED | null;

type TagTreeProps = {
  items: LibraryItem[];
  selected: TagSelection;
  onSelect: (tag: TagSelection) => void;
  onDropItem?: (itemId: string, tagName: string) => void;
  // Called with the old and new name when a tag is renamed in place.
  onRenameTag?: (from: string, to: string) => Promise<void>;
};

export const DRAG_ITEM_ID_TYPE = 'text/rigormining-item-id';

const rowBase = 'flex items-center justify-between gap-2 px-3 py-1.5 rounded-md text-sm cursor-pointer select-none';
const rowIdle = 'text-gray-600 hover:bg-gray-100';
const rowActive = 'bg-indigo-50 text-indigo-700 font-medium';

export const TagTree: React.FC<TagTreeProps> = ({ items, selected, onSelect, onDropItem, onRenameTag }) => {
  // The tag being renamed, and the text typed so far.
  const [editing, setEditing] = useState<string | null>(null);
  const [draft, setDraft] = useState('');
  const [renameError, setRenameError] = useState('');
  // Set by Escape, so the blur that follows doesn't save the edit.
  const cancelled = useRef(false);

  const counts = new Map<string, number>();
  let untaggedCount = 0;
  items.forEach((item) => {
    if (item.tags.length === 0) {
      untaggedCount += 1;
      return;
    }
    item.tags.forEach((t) => counts.set(t.name, (counts.get(t.name) ?? 0) + 1));
  });
  const tags = Array.from(counts.entries()).sort((a, b) => a[0].localeCompare(b[0]));

  const allowDrop = (e: React.DragEvent) => {
    if (!onDropItem) return;
    e.preventDefault();
  };

  const handleDrop = (tagName: string) => (e: React.DragEvent) => {
    if (!onDropItem) return;
    e.preventDefault();
    const itemId = e.dataTransfer.getData(DRAG_ITEM_ID_TYPE);
    if (itemId) onDropItem(itemId, tagName);
  };

  const startRename = (name: string) => {
    if (!onRenameTag) return;
    cancelled.current = false;
    setEditing(name);
    setDraft(name);
    setRenameError('');
  };

  const finishRename = async () => {
    const from = editing;
    const to = draft.trim();
    setEditing(null);
    if (cancelled.current || !from || !onRenameTag || !to || to === from) return;
    try {
      await onRenameTag(from, to);
    } catch (err) {
      setRenameError(`Couldn't rename "${from}": ${(err as Error).message}`);
    }
  };

  const onRenameKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') e.currentTarget.blur();
    if (e.key === 'Escape') {
      cancelled.current = true;
      setEditing(null);
    }
  };

  return (
    <nav className="w-56 shrink-0 border-r border-gray-200 bg-gray-50 px-2 py-4 overflow-y-auto">
      <button onClick={() => onSelect(null)} className={`${rowBase} w-full ${selected === null ? rowActive : rowIdle}`}>
        <span className="flex items-center gap-2 truncate">
          <LibraryIcon className="h-4 w-4 flex-shrink-0" aria-hidden="true" />
          All Items
        </span>
        <span className="text-xs text-gray-400">{items.length}</span>
      </button>

      {tags.length > 0 && (
        <div className="mt-3">
          <p className="px-3 text-xs font-semibold text-gray-400 uppercase tracking-wide">Projects</p>
          <ul className="mt-1 space-y-0.5">
            {tags.map(([name, count]) => (
              <li
                key={name}
                onDragOver={allowDrop}
                onDrop={handleDrop(name)}
                className={`flex items-center rounded-md text-sm ${selected === name ? rowActive : rowIdle}`}
              >
                {/* Clicking the tag icon renames the tag. */}
                <button
                  type="button"
                  onClick={() => startRename(name)}
                  disabled={!onRenameTag}
                  className="group relative flex-shrink-0 pl-3 pr-2 py-1.5 hover:text-indigo-600 disabled:cursor-default"
                  title={`Rename "${name}"`}
                  aria-label={`Rename tag ${name}`}
                >
                  <TagIcon className={`h-4 w-4 ${onRenameTag ? 'group-hover:opacity-0' : ''}`} aria-hidden="true" />
                  {onRenameTag && (
                    <PencilIcon
                      className="h-4 w-4 absolute left-3 top-1.5 opacity-0 group-hover:opacity-100"
                      aria-hidden="true"
                    />
                  )}
                </button>

                {editing === name ? (
                  <input
                    autoFocus
                    value={draft}
                    onChange={(e) => setDraft(e.target.value)}
                    onBlur={finishRename}
                    onKeyDown={onRenameKeyDown}
                    onFocus={(e) => e.currentTarget.select()}
                    className="flex-1 min-w-0 mr-2 my-0.5 px-1.5 py-0.5 text-sm border border-indigo-300 rounded focus:ring-indigo-500 focus:border-indigo-500"
                    aria-label={`New name for ${name}`}
                  />
                ) : (
                  <button
                    onClick={() => onSelect(name)}
                    onDoubleClick={() => startRename(name)}
                    className="flex flex-1 min-w-0 items-center justify-between gap-2 pr-3 py-1.5 cursor-pointer select-none text-left"
                    title={name}
                  >
                    <span className="truncate">{name}</span>
                    <span className="text-xs text-gray-400 flex-shrink-0">{count}</span>
                  </button>
                )}
              </li>
            ))}
          </ul>
          {renameError && <p className="px-3 mt-1 text-xs text-red-600">{renameError}</p>}
        </div>
      )}

      {untaggedCount > 0 && (
        <button
          onClick={() => onSelect(UNTAGGED)}
          className={`${rowBase} w-full mt-3 ${selected === UNTAGGED ? rowActive : rowIdle}`}
        >
          <span className="truncate">Untagged</span>
          <span className="text-xs text-gray-400">{untaggedCount}</span>
        </button>
      )}
    </nav>
  );
};

export default TagTree;
