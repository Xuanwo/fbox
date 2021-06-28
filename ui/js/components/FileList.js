import m from '../../vendor/mithril.js'
import listFiles from '../actions/listFiles.js'
import uploadFile from '../actions/uploadFile.js'

import FileUploader from './FileUploader.js'

export default function FileList () {
  const state = {
    files: []
  }

  async function refresh () {
    state.files = await listFiles();
    m.redraw();
  }

  function handleFileUpload (files) {
    Array.from(files).forEach(async file => {
      await uploadFile(file)
      refresh();
    })
  }

  return {
    oncreate: refresh,

    view: () => {
      return m('fb-filelist',
        m('ul',
          state.files.map(file => {
            return m('li', { key: file.Name },
              m('a', { target: '_NEW', href: '/api/download/' + file.Name },
                m('img', { src: './images/file-icon.png' }),
                file.Name
              )
            )
          })
        ),
        m(FileUploader, { onChange: handleFileUpload }, 'Upload a file')
      )
    }
  }
}
