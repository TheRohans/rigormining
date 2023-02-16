declare module '*.module.css' {
  const content: Record<string, string>;
  export default content;
}

declare module '*.svg' {
  const content: any;
  export default content;
}

declare module '*.jpg' {
  const content: any;
  export default content;
}
declare module '*.png' {
  const content: any;
  export default content;
}

declare module "worker-loader!*" {
  // You need to change `Worker`, if you specified a different 
  // value for the `workerType` option
  class WebpackWorker extends Worker {
    constructor();
  }

  // Uncomment this if you set the `esModule` option to `false`
  // export = WebpackWorker;
  export default WebpackWorker;
}