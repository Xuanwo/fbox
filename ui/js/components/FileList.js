import m from '../../vendor/mithril.js'
import listFiles from '../actions/listFiles.js'

export default function FileList (vnode) {
  const { context } = vnode.attrs;

  const state = {
    files: []
  }

  return {
    oncreate: () => {
      listFiles(context);
    },

    view: () => {
      return m('fb-filelist',
        m('ul',
          context.files.map(file => {
            return m('li', { key: file.Name },
              m('a', { target: '_NEW', href: '/api/download/' + file.Name },
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
