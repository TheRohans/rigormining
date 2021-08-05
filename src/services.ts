import Amplify from '@aws-amplify/core';
import Storage from '@aws-amplify/storage';

export function configureAmplify() {
  Amplify.configure(
    {
      Auth: {
        // REQUIRED only for Federated Authentication - Amazon Cognito Identity Pool ID
        identityPoolId: process.env.KNOTSET_identityPoolId,
        // REQUIRED - Amazon Cognito Region
        region: process.env.KNOTSET_region,
        // OPTIONAL - Amazon Cognito Federated Identity Pool Region
        // Required only if it's different from Amazon Cognito Region
        identityPoolRegion: process.env.KNOTSET_region,
        // OPTIONAL - Amazon Cognito User Pool ID
        userPoolId: process.env.KNOTSET_userPoolId,
        // OPTIONAL - Amazon Cognito Web Client ID (26-char alphanumeric string)
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

