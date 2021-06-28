import m from '../vendor/mithril.js'

import Header from './components/Header.js'
import FileList from './components/FileList.js'
import FileUploader from './components/FileUploader.js'

import listFiles from './actions/listFiles.js'
import uploadFile from './actions/uploadFile.js'

document.addEventListener('DOMContentLoaded', function () {
  const context = {
    files: []
  };

  function handleFileUpload (files) {
    Array.from(files).forEach(async file => {
      await uploadFile(file)
      listFiles(context);
    })
  }

  function render () {
    const ui = m('div',
      m(Header, { context }),
      m(FileList, { context }),
      m(FileUploader, { context, onChange: handleFileUpload }, 'Upload a file')
    );

    m.render(document.body, ui);
  }

  m.redraw = render;
  context.redraw = render;
  render();
});
