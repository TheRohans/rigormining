import React from 'react';
import { LibraryIcon, TagIcon } from '@heroicons/react/solid';
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
};

export const DRAG_ITEM_ID_TYPE = 'text/rigormining-item-id';

const rowBase = 'flex items-center justify-between gap-2 px-3 py-1.5 rounded-md text-sm cursor-pointer select-none';
const rowIdle = 'text-gray-600 hover:bg-gray-100';
const rowActive = 'bg-indigo-50 text-indigo-700 font-medium';

export const TagTree: React.FC<TagTreeProps> = ({ items, selected, onSelect, onDropItem }) => {
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
              <li key={name}>
                <button
                  onClick={() => onSelect(name)}
                  onDragOver={allowDrop}
                  onDrop={handleDrop(name)}
                  className={`${rowBase} w-full ${selected === name ? rowActive : rowIdle}`}
                  title={name}
                >
                  <span className="flex items-center gap-2 truncate">
                    <TagIcon className="h-4 w-4 flex-shrink-0" aria-hidden="true" />
                    <span className="truncate">{name}</span>
                  </span>
                  <span className="text-xs text-gray-400 flex-shrink-0">{count}</span>
                </button>
              </li>
            ))}
          </ul>
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
