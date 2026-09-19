declare const window: any;

export const log = (msg: string, ...args: any[]) => {
  if (window.rm_log) window.rm_log(msg, args);
  else console.log(msg, args);
};

export default log;
