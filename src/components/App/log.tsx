declare const window: any;

export const log = (msg: string, ...args: any[]) => {
  if (window.ks_log) window.ks_log(msg, args);
  else console.log(msg, args);
};

export default log;
