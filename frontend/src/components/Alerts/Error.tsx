import React from 'react';
import { XCircleIcon } from '@heroicons/react/solid';

type ErrorProps = {
  text: string;
};

export const Error: React.FC<ErrorProps> = ({ text }) => {
  return (
    <div className="rounded-md bg-red-50 p-4">
      <div className="flex">
        <div className="flex-shrink-0">
          <XCircleIcon className="h-5 w-5 text-red-400" aria-hidden="true" />
        </div>
        <div className="ml-3">
          <h3 className="text-sm font-medium text-red-800">{text}</h3>
          {/* <div className="mt-2 text-sm text-red-700">
            <ul className="list-disc pl-5 space-y-1">
              <li>Your password must be at least 8 characters</li>
              <li>Your password must include at least one pro wrestling finishing move</li>
            </ul>
          </div> */}
        </div>
      </div>
    </div>
  );
};

export default Error;
