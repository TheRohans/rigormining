
import {log} from './components/App/log'

export function StorageConfig(bucket: string, level: string) {
  
  log(bucket, level);

  // Storage.configure({
  //   bucket: bucket,
  //   level: level,
  // });
}

