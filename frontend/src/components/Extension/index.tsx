import React from 'react';
import { Link } from 'react-router-dom';
import Navigation from '@components/App/Navigation';
import { api } from '../../api/client';

type BrowserCardProps = {
  name: string;
  browser: 'chrome' | 'firefox';
  steps: string[];
};

const BrowserCard: React.FC<BrowserCardProps> = ({ name, browser, steps }) => (
  <div className="border border-gray-200 rounded-lg p-6 bg-white space-y-4">
    <div className="flex items-center justify-between">
      <h2 className="text-lg font-semibold text-gray-900">{name}</h2>
      <a
        href={api.extensionUrl(browser)}
        className="px-3 py-1.5 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
      >
        Download
      </a>
    </div>
    <ol className="list-decimal list-inside text-sm text-gray-600 space-y-1">
      {steps.map((step, i) => (
        <li key={i}>{step}</li>
      ))}
    </ol>
  </div>
);

export const Extension: React.FC = () => {
  return (
    <>
      <Navigation loggedIn={true} />
      <div className="max-w-3xl mx-auto sm:px-6 lg:px-8 py-10 space-y-6">
        <div>
          <h1 className="text-xl font-semibold text-gray-900">Get the browser extension</h1>
          <p className="text-sm text-gray-600 mt-1">
            Capture a paper straight from your browser into your library. This is a beta build - it&apos;s not on the
            Chrome Web Store or Firefox Add-ons yet, so for now it installs manually from the zip below. Once it&apos;s
            published, this page will link there instead.
          </p>
        </div>

        <div className="grid sm:grid-cols-2 gap-4">
          <BrowserCard
            name="Chrome / Edge"
            browser="chrome"
            steps={[
              'Download and unzip it somewhere you’ll keep it (deleting the folder later removes the extension).',
              'Go to chrome://extensions and turn on "Developer mode" (top right).',
              'Click "Load unpacked" and select the unzipped folder.',
            ]}
          />
          <BrowserCard
            name="Firefox"
            browser="firefox"
            steps={[
              'Download and unzip it.',
              'Go to about:debugging#/runtime/this-firefox.',
              'Click "Load Temporary Add-on…" and pick manifest.json inside the unzipped folder.',
              'Firefox removes temporary add-ons on restart, so you’ll need to reload it there each time - normal for a beta build, not a bug.',
            ]}
          />
        </div>

        <div className="text-sm text-gray-600 border-t border-gray-200 pt-4">
          After installing, open the extension&apos;s options page and set your server URL and an API token - create one
          on the{' '}
          <Link to="/settings" className="text-indigo-600 hover:text-indigo-800">
            Settings page
          </Link>
          .
        </div>

        <div className="border-t border-gray-200 pt-6">
          <div className="border border-gray-200 rounded-lg p-6 bg-white space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-gray-900">Kobo sync agent skill</h2>
              <a
                href={api.skillUrl('kobo-sync')}
                className="px-3 py-1.5 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
              >
                Download
              </a>
            </div>
            <p className="text-sm text-gray-600">
              A skill for Claude Code and other AI coding agents that support the <code>.agents/skills</code> format.
              Once installed, you can just ask your agent to sync your Kobo and it drives the API for you - queue items
              from the Sync page above, then say something like &quot;sync my kobo&quot; with the device plugged in.
              Requires <strong>Python 3</strong> to be installed (no other dependencies) - most Mac and Linux machines
              already have it; check with <code>python3 --version</code> in a terminal.
            </p>
            <ol className="list-decimal list-inside text-sm text-gray-600 space-y-1">
              <li>Download and unzip it.</li>
              <li>
                Move the unzipped <code>kobo-sync</code> folder into <code>~/.agents/skills/</code> (create that folder
                if it doesn&apos;t exist yet).
              </li>
              <li>Restart your agent so it picks up the new skill.</li>
            </ol>
          </div>
        </div>
      </div>
    </>
  );
};

export default Extension;
