import React, { useEffect, useState } from 'react';
import Navigation from '@components/App/Navigation';
import { api, ApiToken } from '../../api/client';

export const Settings: React.FC = () => {
  const [tokens, setTokens] = useState<ApiToken[]>([]);
  const [name, setName] = useState('');
  const [justCreated, setJustCreated] = useState<ApiToken | null>(null);

  const load = () => {
    api.listTokens().then(setTokens);
  };

  useEffect(() => {
    load();
  }, []);

  const create = async () => {
    if (!name.trim()) return;
    const token = await api.createToken(name.trim());
    setJustCreated(token);
    setName('');
    load();
  };

  const remove = async (id: string) => {
    await api.deleteToken(id);
    load();
  };

  return (
    <>
      <Navigation loggedIn={true} />
      <div className="max-w-2xl mx-auto sm:px-6 lg:px-8 py-10">
        <h1 className="text-lg font-medium text-gray-900 mb-2">API tokens</h1>
        <p className="text-sm text-gray-600 mb-6">
          Used by the browser extension (and any other client that can't do a browser OAuth login) to authenticate as
          you. Paste one into the extension's options page.
        </p>

        {justCreated && (
          <div className="mb-6 p-4 bg-yellow-50 border border-yellow-200 rounded-md text-sm">
            <p className="font-medium">Copy this now - it won't be shown again:</p>
            <code className="block mt-2 break-all">{justCreated.value}</code>
          </div>
        )}

        <div className="flex gap-2 mb-6">
          <input
            className="flex-1 border border-gray-300 rounded-md shadow-sm p-2"
            placeholder="e.g. browser extension"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
          <button
            onClick={create}
            className="px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
          >
            Create token
          </button>
        </div>

        <ul className="divide-y divide-gray-200">
          {tokens.map((t) => (
            <li key={t.id} className="py-3 flex justify-between items-center">
              <div>
                <p className="text-sm font-medium text-gray-900">{t.name}</p>
                <p className="text-xs text-gray-500">created {t.created_at}</p>
              </div>
              <button onClick={() => remove(t.id)} className="text-sm text-red-600 hover:text-red-900">
                Revoke
              </button>
            </li>
          ))}
        </ul>
      </div>
    </>
  );
};

export default Settings;
