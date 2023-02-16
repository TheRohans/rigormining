// import Amplify from '@aws-amplify/core';
// import Storage from '@aws-amplify/storage';
// import awsconfig from './aws-exports';

export function configureAmplify() {
  // Amplify.configure(
  //   {
  //     Auth: {
  //       // REQUIRED only for Federated Authentication - Amazon Cognito Identity Pool ID
  //       identityPoolId: process.env.KNOTSET_identityPoolId,
  //       // REQUIRED - Amazon Cognito Region
  //       region: awsconfig.aws_cognito_region, //  process.env.KNOTSET_region,
  //       // OPTIONAL - Amazon Cognito Federated Identity Pool Region
  //       // Required only if it's different from Amazon Cognito Region
  //       identityPoolRegion: process.env.KNOTSET_region,
  //       // OPTIONAL - Amazon Cognito User Pool ID
  //       userPoolId: awsconfig.aws_user_pools_id, // process.env.KNOTSET_userPoolId,
  //       // OPTIONAL - Amazon Cognito Web Client ID (26-char alphanumeric string)
  //       userPoolWebClientId: awsconfig.aws_user_pools_web_client_id, // process.env.KNOTSET_userPoolWebClientId,
  //     },
  //     Storage: {
  //       bucket: awsconfig.aws_user_files_s3_bucket, // process.env.KNOTSET_bucket_name,
  //       region: awsconfig.aws_user_files_s3_bucket_region, // process.env.KNOTSET_region,
  //       identityPoolId: process.env.KNOTSET_identityPoolId
  //     },
  //     // API: {
  //     // }
  //     'aws_appsync_graphqlEndpoint': awsconfig.aws_appsync_graphqlEndpoint,
  //     'aws_appsync_region': awsconfig.aws_appsync_region,
  //     'aws_appsync_authenticationType': awsconfig.aws_appsync_authenticationType,
  //   }
  // );
}

export function SetS3Config(bucket: string, level: string) {
  // Storage.configure({
  //   bucket: bucket,
  //   level: level,
  // });
}

