import React, { useState, useRef, useEffect, ChangeEvent, KeyboardEvent } from 'react';
import ItemList from './ItemList';
import Book, { BookFormats } from './Book';
import Navigation from '@components/App/Navigation';
import TagTree, { TagSelection, UNTAGGED } from '@components/App/TagTree';
import { CloudUploadIcon, SearchIcon } from '@heroicons/react/solid';
import { api, LibraryItem } from '../../api/client';
import log from '@components/App/log';

export const Library: React.FC = () => {
  const [items, setItems] = useState<LibraryItem[]>([]);
  const [currentBook, setCurrentBook] = useState<string>('');
  const [currentBookType, setCurrentBookType] = useState<BookFormats>(undefined);
  const [currentBookTitle, setCurrentBookTitle] = useState<string>('');

  const [selectedTag, setSelectedTag] = useState<TagSelection>(null);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [search, setSearch] = useState('');
  const [newProjectName, setNewProjectName] = useState('');

  const dropArea = useRef<HTMLDivElement>();

  const loadItems = () => {
    api
      .listItems()
      .then(setItems)
      .catch((e) => log(e.message));
  };

  useEffect(() => {
    loadItems();
  }, []);

  const viewItem = async (item: LibraryItem) => {
    log('View item: ' + item.title);
    const blob = await api.fetchFileBlob(item.id);
    setCurrentBookType(item.file_type === 'epub' ? 'epub' : 'pdf');
    setCurrentBookTitle(item.title);
    setCurrentBook(URL.createObjectURL(blob));
  };

  const closeBook = () => {
    if (currentBook) URL.revokeObjectURL(currentBook);
    setCurrentBook('');
    setCurrentBookType(undefined);
    setCurrentBookTitle('');
  };

  const downloadBook = (url: string) => {
    window.open(url, '_blank');
  };

  const uploadFile = async (file: File) => {
    const title = file.name.replace(/\.[^/.]+$/, '');
    try {
      await api.createItem({ title }, file);
      loadItems();
    } catch (e) {
      log('upload failed: ' + (e as Error).message);
    }
  };

  const uploadFiles = (e: ChangeEvent<HTMLInputElement>) => {
    Array.from(e.target?.files ?? []).forEach(uploadFile);
  };

  const highlight = (e: DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const unhighlight = (e: DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleDrop = (e: DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    // Item rows being dragged onto the tag tree set this - ignore that case
    // here so dropping a row back onto the page doesn't try to re-upload it.
    if (e.dataTransfer.types.includes('text/rigormining-item-id')) return;
    Array.from(e.dataTransfer.files).forEach(uploadFile);
  };

  useEffect(() => {
    const listeners = new Map<string, EventListener>();

    if (dropArea?.current) {
      ['dragenter', 'dragover'].forEach((eventName) => {
        dropArea.current.addEventListener(eventName, highlight, false);
        listeners.set(eventName, highlight);
      });
      ['dragleave'].forEach((eventName) => {
        dropArea.current.addEventListener(eventName, unhighlight, false);
        listeners.set(eventName, unhighlight);
      });
      ['drop'].forEach((eventName) => {
        dropArea.current.addEventListener(eventName, handleDrop, false);
        listeners.set(eventName, handleDrop);
      });

      return () => {
        listeners.forEach((l, s) => {
          dropArea?.current?.removeEventListener(s, l);
        });
      };
    }
  });

  const toggleSelected = (id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const addTagToItem = async (itemId: string, tagName: string) => {
    await api.addTag(itemId, tagName);
    loadItems();
  };

  const addSelectedToProject = async () => {
    const name = newProjectName.trim();
    if (!name || selectedIds.size === 0) return;
    await Promise.all(Array.from(selectedIds).map((id) => api.addTag(id, name)));
    setNewProjectName('');
    setSelectedIds(new Set());
    loadItems();
  };

  const onNewProjectKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') addSelectedToProject();
  };

  const allTagNames = Array.from(new Set(items.flatMap((i) => i.tags.map((t) => t.name)))).sort();

  const visibleItems = items
    .filter((item) => {
      if (selectedTag === null) return true;
      if (selectedTag === UNTAGGED) return item.tags.length === 0;
      return item.tags.some((t) => t.name === selectedTag);
    })
    .filter((item) => {
      if (!search.trim()) return true;
      const q = search.trim().toLowerCase();
      return item.title.toLowerCase().includes(q) || (item.authors ?? '').toLowerCase().includes(q);
    });

  return (
    <>
      {currentBook.length <= 0 && (
        <>
          <Navigation loggedIn={true} />

          <div className="flex" style={{ height: 'calc(100vh - 4rem)' }}>
            <TagTree items={items} selected={selectedTag} onSelect={setSelectedTag} onDropItem={addTagToItem} />

            <div ref={dropArea} className="flex-1 overflow-y-auto">
              <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
                <label
                  htmlFor="fileElem"
                  className="flex items-center gap-2 text-sm text-gray-500 border border-dashed border-gray-300 rounded-md px-3 py-2 mb-4 cursor-pointer hover:border-indigo-400 hover:text-indigo-600"
                >
                  <CloudUploadIcon className="h-4 w-4 flex-shrink-0" aria-hidden="true" />
                  Drop a PDF or EPUB here, or click to browse
                  <input
                    type="file"
                    id="fileElem"
                    multiple
                    onChange={uploadFiles}
                    className="hidden"
                    accept=".pdf,.epub,application/pdf,application/epub,application/epub+zip,application/x-pdf"
                  />
                </label>

                <div className="flex items-center gap-3 mb-3">
                  <div className="relative flex-1 max-w-xs">
                    <SearchIcon className="h-4 w-4 text-gray-400 absolute left-2.5 top-2.5" aria-hidden="true" />
                    <input
                      value={search}
                      onChange={(e) => setSearch(e.target.value)}
                      placeholder="Search title or author"
                      className="w-full pl-8 pr-3 py-1.5 text-sm border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
                    />
                  </div>

                  {selectedIds.size > 0 && (
                    <div className="flex items-center gap-2 ml-auto">
                      <span className="text-xs text-gray-500">{selectedIds.size} selected</span>
                      <input
                        list="rigormining-project-names"
                        value={newProjectName}
                        onChange={(e) => setNewProjectName(e.target.value)}
                        onKeyDown={onNewProjectKeyDown}
                        placeholder="Add to project..."
                        className="text-sm border border-gray-300 rounded-md shadow-sm px-2 py-1.5 focus:ring-indigo-500 focus:border-indigo-500"
                      />
                      <datalist id="rigormining-project-names">
                        {allTagNames.map((t) => (
                          <option key={t} value={t} />
                        ))}
                      </datalist>
                      <button
                        onClick={addSelectedToProject}
                        disabled={!newProjectName.trim()}
                        className="px-3 py-1.5 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50"
                      >
                        Add
                      </button>
                    </div>
                  )}
                </div>

                <ItemList
                  items={visibleItems}
                  viewItem={viewItem}
                  selectedIds={selectedIds}
                  onToggleSelected={toggleSelected}
                />
              </div>
            </div>
          </div>
        </>
      )}
      {currentBook.length > 0 && (
        <Book
          type={currentBookType}
          url={currentBook}
          title={currentBookTitle}
          closeBook={closeBook}
          downloadBook={downloadBook}
        />
      )}
    </>
  );
};

export default Library;
