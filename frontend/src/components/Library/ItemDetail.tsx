import React, { useEffect, useState } from 'react';
import { Link, useHistory, useParams } from 'react-router-dom';
import { ChevronLeftIcon, ExternalLinkIcon, XIcon } from '@heroicons/react/solid';
import Navigation from '@components/App/Navigation';
import { api, LibraryItem } from '../../api/client';

const label = 'block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1';
const input =
  'block w-full border border-gray-300 rounded-md shadow-sm px-3 py-2 text-sm focus:ring-indigo-500 focus:border-indigo-500';

export const ItemDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const history = useHistory();

  const [item, setItem] = useState<LibraryItem | null>(null);
  const [newTag, setNewTag] = useState('');
  const [saving, setSaving] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState('');

  const load = () => {
    api.getItem(id).then(setItem);
  };

  useEffect(() => {
    load();
  }, [id]);

  if (!item) {
    return null;
  }

  const field = (key: keyof LibraryItem) => (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setItem({ ...item, [key]: e.target.value } as LibraryItem);
  };

  const save = async () => {
    setSaving(true);
    try {
      const updated = await api.updateItem(item.id, {
        title: item.title,
        authors: item.authors,
        doi: item.doi,
        isbn: item.isbn,
        year: item.year,
        notes: item.notes,
        source_url: item.source_url,
        item_type: item.item_type,
        venue: item.venue,
        volume: item.volume,
        number: item.number,
        pages: item.pages,
        publisher: item.publisher,
      });
      setItem(updated);
    } finally {
      setSaving(false);
    }
  };

  const uploadFile = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setUploading(true);
    setUploadError('');
    try {
      const updated = await api.uploadItemFile(item.id, file);
      setItem(updated);
    } catch (err) {
      setUploadError((err as Error).message);
    } finally {
      setUploading(false);
      e.target.value = '';
    }
  };

  const addTag = async () => {
    if (!newTag.trim()) return;
    const updated = await api.addTag(item.id, newTag.trim());
    setItem(updated);
    setNewTag('');
  };

  const removeTag = async (tagId: string) => {
    const updated = await api.removeTag(item.id, tagId);
    setItem(updated);
  };

  const remove = async () => {
    if (!window.confirm(`Delete "${item.title}"?`)) return;
    await api.deleteItem(item.id);
    history.push('/library');
  };

  return (
    <>
      <Navigation loggedIn={true} />
      <div className="max-w-3xl mx-auto sm:px-6 lg:px-8 py-8">
        <Link to="/library" className="inline-flex items-center gap-1 text-sm text-gray-500 hover:text-indigo-600 mb-4">
          <ChevronLeftIcon className="h-4 w-4" aria-hidden="true" />
          Back to library
        </Link>

        <div className="space-y-5 bg-white border border-gray-200 rounded-lg p-6">
          <div>
            <label className={label}>Title</label>
            <input className={input} value={item.title} onChange={field('title')} />
          </div>

          <div>
            <label className={label}>Authors</label>
            <input
              className={input}
              value={item.authors ?? ''}
              onChange={field('authors')}
              placeholder="Last, First; Last, First"
            />
          </div>

          <div>
            <label className={label}>Source URL</label>
            <div className="flex gap-2">
              <input
                className={input}
                value={item.source_url ?? ''}
                onChange={field('source_url')}
                placeholder="https://..."
              />
              {item.source_url && (
                <a
                  href={item.source_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="px-3 py-2 border border-gray-300 rounded-md text-sm text-gray-500 hover:text-indigo-600 hover:border-indigo-300 flex-shrink-0"
                  title="Open link"
                >
                  <ExternalLinkIcon className="h-4 w-4" aria-hidden="true" />
                </a>
              )}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className={label}>DOI</label>
              <input className={input} value={item.doi ?? ''} onChange={field('doi')} />
            </div>
            <div>
              <label className={label}>Year</label>
              <input className={input} value={item.year ?? ''} onChange={field('year')} />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className={label}>Item Type</label>
              <input
                className={input}
                value={item.item_type ?? ''}
                onChange={field('item_type')}
                placeholder="journalArticle, book, thesis..."
              />
            </div>
            <div>
              <label className={label}>Venue</label>
              <input
                className={input}
                value={item.venue ?? ''}
                onChange={field('venue')}
                placeholder="Journal or conference/book title"
              />
            </div>
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div>
              <label className={label}>Volume</label>
              <input className={input} value={item.volume ?? ''} onChange={field('volume')} />
            </div>
            <div>
              <label className={label}>Number</label>
              <input className={input} value={item.number ?? ''} onChange={field('number')} />
            </div>
            <div>
              <label className={label}>Pages</label>
              <input className={input} value={item.pages ?? ''} onChange={field('pages')} placeholder="12-34" />
            </div>
          </div>

          <div>
            <label className={label}>Publisher</label>
            <input className={input} value={item.publisher ?? ''} onChange={field('publisher')} />
          </div>

          <div>
            <label className={label}>Notes</label>
            <textarea className={input} rows={4} value={item.notes ?? ''} onChange={field('notes')} />
          </div>

          {!item.file_type && (
            <div>
              <label className={label}>Attach a file</label>
              <p className="text-xs text-gray-500 mb-1">
                This entry has no PDF/EPUB attached yet - add one you've downloaded by hand.
              </p>
              <input
                type="file"
                accept=".pdf,.epub,application/pdf,application/epub+zip"
                onChange={uploadFile}
                disabled={uploading}
                className="text-sm"
              />
              {uploading && <p className="text-xs text-gray-500 mt-1">Uploading...</p>}
              {uploadError && <p className="text-xs text-red-600 mt-1">{uploadError}</p>}
            </div>
          )}

          <div>
            <label className={label}>Projects / Tags</label>
            {item.tags.length > 0 && (
              <div className="flex flex-wrap gap-1.5 mb-2">
                {item.tags.map((t) => (
                  <span
                    key={t.id}
                    className="inline-flex items-center gap-1 pl-2 pr-1 py-0.5 text-xs rounded-full bg-indigo-100 text-indigo-800"
                  >
                    {t.name}
                    <button
                      onClick={() => removeTag(t.id)}
                      className="rounded-full hover:bg-indigo-200 p-0.5"
                      title={`Remove ${t.name}`}
                    >
                      <XIcon className="h-3 w-3" aria-hidden="true" />
                    </button>
                  </span>
                ))}
              </div>
            )}
            <div className="flex gap-2">
              <input
                className={`flex-1 ${input}`}
                value={newTag}
                onChange={(e) => setNewTag(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && addTag()}
                placeholder="add to a project"
              />
              <button
                onClick={addTag}
                className="px-3 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
              >
                Add
              </button>
            </div>
          </div>

          <div className="flex flex-wrap gap-3 pt-4 border-t border-gray-200">
            <button
              onClick={save}
              disabled={saving}
              className="px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50"
            >
              Save
            </button>

            {item.file_type && (
              <a
                href={api.fileUrl(item.id)}
                className="px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
              >
                Download
              </a>
            )}

            <a
              href={api.exportUrl(item.id)}
              className="px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
            >
              Export to Markdown
            </a>

            <a
              href={api.bibtexUrl(item.id)}
              className="px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
            >
              Export to BibTeX
            </a>

            <button
              onClick={remove}
              className="px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-red-600 hover:bg-red-700 ml-auto"
            >
              Delete
            </button>
          </div>
        </div>
      </div>
    </>
  );
};

export default ItemDetail;
