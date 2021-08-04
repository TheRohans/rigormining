import Amplify from '@aws-amplify/core';
import Storage from '@aws-amplify/storage';

export function configureAmplify() {
  Amplify.configure(
    {
      Auth: {
        identityPoolId: process.env.KNOTSET_identityPoolId,
        region: process.env.KNOTSET_region,
        userPoolId: process.env.KNOTSET_userPoolId,
        userPoolWebClientId: process.env.KNOTSET_userPoolWebClientId,
      },
      Storage: {
        bucket: process.env.KNOTSET_bucket_name,
        region: process.env.KNOTSET_region,
        identityPoolId: process.env.KNOTSET_identityPoolId
      }
    }
  );
}

export function SetS3Config(bucket: string, level: string) {
  Storage.configure({
    bucket: bucket,
    level: level,
    region: process.env.KNOTSET_region,
    identityPoolId: process.env.KNOTSET_identityPoolId,
  });
}

