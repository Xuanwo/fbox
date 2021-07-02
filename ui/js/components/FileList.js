import m from '../../vendor/mithril.js';
import listFiles from '../actions/listFiles.js';

export default function FileList (vnode) {
  const { context } = vnode.attrs;

  return {
    oncreate: () => {
      listFiles(context);
    },

    view: () => {
      return m('fb-filelist',
        m('ul',
          context.files.map(file => {
            return m('li', { key: file.Name },
              m('a', { target: '_NEW', href: '/download/' + file.Name },
                m('img', { src: './img/file-icon.png' }),
                file.Name
              )
            );
          })
        )
      );
    }
  };
}
