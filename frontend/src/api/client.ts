// API_BASE is empty in production (frontend is served by the same Go
// process as the API) and points at the backend's dev port during
// `npm start`, where webpack-dev-server and the Go API run on different
// ports.
const API_BASE = process.env.RIGORMINING_API_BASE || '';

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const isFormData = options.body instanceof FormData;
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    credentials: 'include',
    headers: {
      ...(!isFormData && options.body ? { 'Content-Type': 'application/json' } : {}),
      ...options.headers,
    },
  });

  if (!res.ok) {
    const text = await res.text().catch(() => '');
    throw new Error(`${res.status} ${res.statusText}${text ? `: ${text}` : ''}`);
  }
  if (res.status === 204) {
    return undefined as unknown as T;
  }
  return res.json();
}

export type Tag = {
  id: string;
  name: string;
};

export type SyncState = 'request_sync' | 'synced' | 'request_remove';

export type LibraryItem = {
  id: string;
  title: string;
  authors: string;
  doi?: string;
  isbn?: string;
  year?: number;
  source_url?: string;
  file_type?: string;
  added_date: string;
  sync_state?: SyncState;
  notes?: string;
  item_type?: string;
  venue?: string;
  volume?: string;
  number?: string;
  pages?: string;
  publisher?: string;
  tags: Tag[];
};

export type WhoAmI = {
  id: string;
  email: string;
  name?: string;
  picture?: string;
};

export type ApiToken = {
  id: string;
  name: string;
  value?: string; // only present in the create response
  created_at: string;
};

export type ListItemsParams = {
  q?: string;
  sync_state?: SyncState;
};

export type ItemMetadataInput = {
  title: string;
  authors?: string;
  doi?: string;
  isbn?: string;
  year?: number;
  source_url?: string;
  notes?: string;
  tags?: string[];
  item_type?: string;
  venue?: string;
  volume?: string;
  number?: string;
  pages?: string;
  publisher?: string;
};

export const api = {
  whoami: () => request<WhoAmI>('/api/v1/whoami'),

  listItems: (params: ListItemsParams = {}) => {
    const qs = new URLSearchParams();
    if (params.q) qs.set('q', params.q);
    if (params.sync_state) qs.set('sync_state', params.sync_state);
    const suffix = qs.toString() ? `?${qs.toString()}` : '';
    return request<LibraryItem[]>(`/api/v1/items${suffix}`);
  },

  getItem: (id: string) => request<LibraryItem>(`/api/v1/items/${id}`),

  uploadItemFile: (id: string, file: File) => {
    const form = new FormData();
    form.set('file', file);
    return request<LibraryItem>(`/api/v1/items/${id}/file`, { method: 'POST', body: form });
  },

  createItem: (metadata: ItemMetadataInput, file?: File) => {
    const form = new FormData();
    form.set('metadata', JSON.stringify(metadata));
    if (file) form.set('file', file);
    return request<LibraryItem>('/api/v1/items', { method: 'POST', body: form });
  },

  updateItem: (id: string, metadata: ItemMetadataInput) =>
    request<LibraryItem>(`/api/v1/items/${id}`, { method: 'PATCH', body: JSON.stringify(metadata) }),

  deleteItem: (id: string) => request<void>(`/api/v1/items/${id}`, { method: 'DELETE' }),

  addTag: (id: string, name: string) =>
    request<LibraryItem>(`/api/v1/items/${id}/tags`, { method: 'POST', body: JSON.stringify({ name }) }),

  removeTag: (id: string, tagId: string) =>
    request<LibraryItem>(`/api/v1/items/${id}/tags/${tagId}`, { method: 'DELETE' }),

  requestSync: (id: string) => request<LibraryItem>(`/api/v1/items/${id}/sync`, { method: 'POST' }),
  cancelSync: (id: string) => request<LibraryItem>(`/api/v1/items/${id}/sync`, { method: 'DELETE' }),
  ackSync: (id: string, result: 'synced' | 'removed') =>
    request<LibraryItem>(`/api/v1/items/${id}/sync/ack`, { method: 'POST', body: JSON.stringify({ result }) }),

  listTokens: () => request<ApiToken[]>('/api/v1/tokens'),
  createToken: (name: string) =>
    request<ApiToken>('/api/v1/tokens', { method: 'POST', body: JSON.stringify({ name }) }),
  deleteToken: (id: string) => request<void>(`/api/v1/tokens/${id}`, { method: 'DELETE' }),

  fileUrl: (id: string) => `${API_BASE}/api/v1/items/${id}/file`,
  exportUrl: (id: string) => `${API_BASE}/api/v1/items/${id}/export.md`,
  bibtexUrl: (id: string) => `${API_BASE}/api/v1/items/${id}/export.bib`,
  extensionUrl: (browser: 'chrome' | 'firefox') => `${API_BASE}/api/v1/extension/${browser}`,
  skillUrl: (name: string) => `${API_BASE}/api/v1/skills/${name}`,

  fetchFileBlob: async (id: string): Promise<Blob> => {
    const res = await fetch(`${API_BASE}/api/v1/items/${id}/file`, { credentials: 'include' });
    if (!res.ok) throw new Error(`could not fetch file: ${res.status}`);
    return res.blob();
  },
};

export const loginUrl = `${API_BASE}/login`;
export const logoutUrl = `${API_BASE}/-/logout`;
export const devLoginUrl = `${API_BASE}/dev-login`;
