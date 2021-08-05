const fs = require('fs');
const path = require('path');

function dirTree(filename) {
  let stats = fs.lstatSync(filename);
  let name = path.basename(filename);

  let info = {
    name: name,
  };

  if (stats.isDirectory()) {
    info.type = 'folder';
    info.children = fs.readdirSync(filename).map(function (child) {
      return dirTree(filename + '/' + child);
    });
  } else {
    let ext = name.substr(name.length - 3, 3);
    if (ext !== '.js' && ext !== 'son') {
      info.type = ext;
    }
  }

  return info;
}

if (module.parent == undefined) {
  // node dirTree.js ~/foo/bar
  var util = require('util');
  console.log(JSON.stringify(dirTree(process.argv[2])));
  // console.log(util.inspect(dirTree(process.argv[2]), false, null));
}
