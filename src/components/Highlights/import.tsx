import React, { useEffect, useRef, useState } from 'react';
import { API, graphqlOperation } from 'aws-amplify';
import { createHighlight, updateHighlight } from '../../graphql/mutations';
// import { listHighlights } from '../../graphql/queries';
// import { DeviceType, HighlightType } from '../../API';
import { Highlight, HighlightList, DeviceList } from './types';
import Navigation from '@components/App/Navigation';

declare const window: any;

export const Import: React.FC = () => {
  const [highlights, setHighlights] =
    useState<HighlightList>(/*[
    {
      deviceType: 'Kindle',
      type: 'Highlight',
      id: 'a5a2b761',
      author: 'Rob Rohan',
      page: 1,
      title: 'Faster after 50 and things',
      text: `[GN] main (entry.c:232) Starting checker thread
    [GN] check_for_device (entry.c:152) Current working dir: /home/rob/Projects/knotset-desk/KnotSet
    [GN] fn_getHighlights (entry.c:91) Going to check a kindle.
    [GN] fn_getHighlights (entry.c:101) File not loaded, going to try to get the highlights from the device
    [GN] load_file_as_json (entry.c:52) Trying to read into file: kindle.json. Got a read size of 77077 and allocation size 77077
    [GN] load_file_as_json (entry.c:55) Reading file into json array...
    [GN] fn_getHighlights (entry.c:103) Back. Returning highlight data...
    [GN] main (entry.c:239) Finishing checker thread`,
      // date: '12345',
    },
    {
      deviceType: 'Kindle',
      type: 'Highlight',
      id: 'a5a2b762',
      author: 'Roab Rohan',
      page: 1,
      title: 'Another book with things in it',
      text: ` If you do not wish to use the wide options, experiment with the following:
      
      Read and write bytes, not characters. Also known as, use binary, not text.
      
      fgetc effectively gets a byte from a file, but if the byte is greater than 127, try treating it as a int instead of a char. fputc, on the other hand, silently ignores putting a char > 127. It will work if you use an int rather than char as the input.
      
      Also, in the open mode, try using binary, so try rb & wb rather than r & w
      `,
      // date: '12345',
    },
    {
      deviceType: 'Kindle',
      type: 'Highlight',
      id: 'a5a2b763',
      author: 'wang luo',
      page: 1,
      title: 'Da Jia Hao',
      text: `这个故事强调了两个牧羊犬和主人之间的竞争，并记载了一个被困在他们之间的男孩大卫的成熟。他的母亲去世，他留给父亲亚当·玛丹（Adam M'Adam）照顾，他是一个讽刺，愤怒的酒鬼，几乎没有赎回的品质。玛丹（M'Adam）是红毛犬（Red Wull）的主人，红毛犬是一种猛烈的狗，会用蛮力将羊放牧。另一只狗是战斗之子鲍勃（Bob）。他用技巧和说服力放牧绵羊。
      他的主人是肯缪尔（Kenmuir）的主人詹姆斯·摩尔（James Moore），他是大卫的代父`,
      // date: '12345',
    },
  ]*/);
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
    // This is pretty hammer-esq...
    // Try to create the new highlight
    try {
      await API.graphql(graphqlOperation(createHighlight, { input: h }));
    } catch (e) {
      console.warn(e);
      // alert(JSON.stringify(e));
      // if it fails, try to update...
      try {
        await API.graphql(graphqlOperation(updateHighlight, { input: h }));
      } catch (err) {
        // if that fails, the meh
        console.error(err);
        // alert(JSON.stringify(err));
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
