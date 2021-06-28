import m from '../../vendor/mithril.js'
import listFiles from '../actions/listFiles.js'

export default function FileList () {
  const state = {
    files: []
  }

  return {
    oncreate: async () => {
      state.files = await listFiles();
      m.redraw();
    },

    view: () => {
      return m('fb-filelist',
        m('ul',
          state.files.map(file => {
            return m('li', { key: file.Name },
              m('a', { href: '/api/download/' + file.Name },
                m('img', { src: './images/file-icon.png' }),
                file.Name
              )
            )
          })
        )
      )
    }
  }
}
