import React from 'react';
import { Link } from 'react-router-dom';
import { EyeIcon, ExternalLinkIcon } from '@heroicons/react/solid';
import { LibraryItem } from '../../api/client';
import { DRAG_ITEM_ID_TYPE } from '@components/App/TagTree';

type ItemListProps = {
  items: LibraryItem[];
  viewItem: (item: LibraryItem) => void;
  selectedIds: Set<string>;
  onToggleSelected: (id: string) => void;
};

const typeBadge = (type: string | undefined, sourceUrl: string | undefined): JSX.Element => {
  if (type === 'pdf') {
    return <span className="px-1.5 py-0.5 text-[11px] font-medium rounded bg-green-100 text-green-700">pdf</span>;
  }
  if (type === 'epub') {
    return <span className="px-1.5 py-0.5 text-[11px] font-medium rounded bg-amber-100 text-amber-700">epub</span>;
  }
  if (sourceUrl) {
    return <span className="px-1.5 py-0.5 text-[11px] font-medium rounded bg-blue-100 text-blue-700">link</span>;
  }
  return <span className="px-1.5 py-0.5 text-[11px] font-medium rounded bg-gray-100 text-gray-500">no file</span>;
};

export const ItemList: React.FC<ItemListProps> = ({ items, viewItem, selectedIds, onToggleSelected }) => {
  if (items.length === 0) {
    return (
      <p className="py-10 text-center text-sm text-gray-500">No items here - drop a PDF or EPUB above to add one.</p>
    );
  }

  return (
    <ul className="divide-y divide-gray-100 border border-gray-200 rounded-md bg-white">
      {items.map((item) => (
        <li
          key={item.id}
          draggable
          onDragStart={(e) => e.dataTransfer.setData(DRAG_ITEM_ID_TYPE, item.id)}
          className="flex items-center gap-3 px-3 py-2 hover:bg-gray-50 cursor-grab"
        >
          <input
            type="checkbox"
            checked={selectedIds.has(item.id)}
            onChange={() => onToggleSelected(item.id)}
            onClick={(e) => e.stopPropagation()}
            className="flex-shrink-0 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
          />

          <div className="min-w-0 flex-1">
            <Link
              to={`/library/${item.id}`}
              className="block text-sm font-medium text-gray-900 truncate hover:text-indigo-600"
            >
              {item.title}
            </Link>
            <p className="text-xs text-gray-500 truncate">
              {[item.authors, item.year].filter(Boolean).join(' · ') || '—'}
            </p>
          </div>

          <div className="flex items-center gap-2 flex-shrink-0">
            {typeBadge(item.file_type, item.source_url)}
            <span
              className={`h-1.5 w-1.5 rounded-full ${item.sync_state === 'synced' ? 'bg-green-500' : 'bg-gray-300'}`}
              title={item.sync_state === 'synced' ? 'On Kobo' : 'Not synced'}
            />
            {item.file_type && (
              <button
                onClick={() => viewItem(item)}
                className="p-1 text-gray-400 hover:text-indigo-600 rounded"
                title="Open"
              >
                <EyeIcon className="h-4 w-4" aria-hidden="true" />
              </button>
            )}
            {!item.file_type && item.source_url && (
              <a
                href={item.source_url}
                target="_blank"
                rel="noopener noreferrer"
                onClick={(e) => e.stopPropagation()}
                className="p-1 text-gray-400 hover:text-indigo-600 rounded"
                title="Open link"
              >
                <ExternalLinkIcon className="h-4 w-4" aria-hidden="true" />
              </a>
            )}
          </div>
        </li>
      ))}
    </ul>
  );
};

export default ItemList;
