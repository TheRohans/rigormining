import React, { useEffect, useRef, useState } from 'react';
import { API, graphqlOperation } from 'aws-amplify';
import { createHighlight, updateHighlight } from '../../graphql/mutations';
// import { listHighlights } from '../../graphql/queries';
// import { DeviceType, HighlightType } from '../../API';
import { Highlight, HighlightList, DeviceList } from './types';
import Navigation from '@components/App/Navigation';
import { useUser } from '@components/App';

declare const window: any;

export const Import: React.FC = () => {
  const [highlights, setHighlights] = useState<HighlightList>();
  const [devices, setDevices] = useState<DeviceList>();

  const kindleBtnRef = useRef<HTMLButtonElement>();
  const koboBtnRef = useRef<HTMLButtonElement>();

  const checkKindle = () => check('kindle');
  const checkKobo = () => check('kobo');

  const check = (key: string) => {
    if (devices && (devices as any)[key] === 1) {
      window
        .ks_getHighlights(key)
        .then((h: HighlightList) => {
          setHighlights(h);
        })
        .catch((err: Error) => {
          // alert incase we are running in the desktop
          // need global notification
          alert(err.message);
        });
    }
  };

  const importHighlights = () => {
    highlights.forEach(async (h) => _import(h));
  };

  const _import = async (h: Highlight) => {
    const user = await useUser();
    // Change the ID to include our username
    // const uploadHighlight = Object.assign({}, h, { id: `${user.username}:${h.id}` });
    h.id = `${user.username}:${h.id}`;

    // This is pretty hammer-esq...
    // Try to create the new highlight
    try {
      await API.graphql(graphqlOperation(createHighlight, { input: h }));
    } catch (e) {
      console.warn(e);
      alert(JSON.stringify(e));
      // if it fails, try to update...
      try {
        await API.graphql(graphqlOperation(updateHighlight, { input: h }));
      } catch (err) {
        // if that fails, the meh
        console.error(err);
        alert(JSON.stringify(err));
      }
    }
  };

  const importHighlight = async (idx: number) => {
    _import(highlights[idx]);
  };

  useEffect(() => {
    if (window.ks_deviceCheck) {
      const interval = setInterval(
        () => {
          window.ks_deviceCheck().then((d: DeviceList) => {
            setDevices(d);
          });
        },
        devices ? 5000 : 1000,
      );
      return () => clearInterval(interval);
    }
  }, []);

  return (
    <>
      <Navigation loggedIn={true} />
      <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
        <div className="py-10">
          <main>
            <div className="bg-white px-4 py-5 border border-gray-200 sm:px-6 mb-2 ">
              <div className="-ml-4 -mt-4 flex justify-between items-center flex-wrap sm:flex-nowrap">
                <div className="ml-4 mt-4">
                  <div className="flex items-center">
                    <div className="ml-4">
                      <h3 className="text-lg leading-6 font-medium text-gray-900">Highlight Import</h3>
                    </div>
                  </div>
                </div>
                <div className="ml-4 mt-4 flex-shrink-0 flex gap-1">
                  {devices?.kindle === 1 && (
                    <button
                      id="kindle"
                      onClick={() => checkKindle()}
                      ref={kindleBtnRef}
                      disabled={false}
                      className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                    >
                      Kindle
                    </button>
                  )}
                  {devices?.kobo === 1 && (
                    <button
                      id="kobo"
                      onClick={() => checkKobo()}
                      ref={koboBtnRef}
                      disabled={false}
                      className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                    >
                      Kobo
                    </button>
                  )}
                  {highlights && (
                    <button
                      id="kobo"
                      onClick={() => importHighlights()}
                      className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                    >
                      Import All
                    </button>
                  )}
                </div>
              </div>
            </div>

            <ul role="list" className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {highlights &&
                highlights.map((highlight, i) => {
                  return (
                    <li
                      key={`${highlight.id}${highlight.type}`}
                      className="col-span-1 bg-white rounded-lg shadow divide-y divide-gray-200"
                    >
                      <div className="w-full flex items-center justify-between p-6 space-x-6">
                        <div className="flex-1">
                          <div className="flex items-center space-x-3 truncate">
                            <h3 className="text-gray-900 text-sm font-medium truncate overflow-ellipsis">
                              {highlight.title}
                            </h3>
                          </div>
                          <p className="mt-1 text-gray-500 text-sm overflow-ellipsis overflow-hidden whitespace-pre-line">
                            {highlight.text}
                          </p>
                        </div>
                      </div>
                      <div>
                        <div className="-mt-px flex divide-x divide-gray-200">
                          <div className="w-0 flex-1 flex">
                            <button
                              onClick={() => importHighlight(i)}
                              className="relative -mr-px w-0 flex-1 inline-flex items-center justify-center py-4 text-sm text-gray-700 font-medium border border-transparent rounded-bl-lg hover:text-gray-500"
                            >
                              <span className="ml-3">Import</span>
                            </button>
                          </div>
                        </div>
                      </div>
                    </li>
                  );
                })}
            </ul>
          </main>
        </div>
      </div>
    </>
  );
};

export default Import;
